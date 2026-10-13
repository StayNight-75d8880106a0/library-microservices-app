package cronjob

import (
	redisdb "borrowing-management-services/internal/infrastructure/redis"
	"borrowing-management-services/internal/usecase"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type CronService struct {
	usecase usecase.WaitingListUsecaseInterface
	cron    *cron.Cron
	locker  redisdb.LockerInterface
}

func NewCronService(waitingListUsecase usecase.WaitingListUsecaseInterface, cron *cron.Cron, locker redisdb.LockerInterface) *CronService {
	return &CronService{
		usecase: waitingListUsecase,
		cron:    cron,
		locker:  locker,
	}
}

func (c *CronService) Start() error {
	if c.cron == nil {
		return errors.New("cron scheduler is not initialized")
	}

	if _, err := c.cron.AddFunc("0 */15 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		release, errLock := c.locker.Acquire(ctx, "lock:cron:v1:waiting-list-expired", 6*time.Minute)

		if errLock != nil {
			if errors.Is(errLock, redisdb.ErrLockNotAcquired) {
				log.Println("[CRON] Skipped — another instance is already running this job.")
				return
			}
			log.Printf("[CRON ERROR] Failed to acquire lock: %v", errLock)
			return
		}

		defer release()

		log.Println("[CRON] Running ProcessExpiredNotified...")
		if err := c.usecase.ProcessExpiredWaitingLists(ctx); err != nil {
			log.Printf("[CRON ERROR] Failed to process expired notifications: %v\n", err)
		}
	}); err != nil {
		return fmt.Errorf("failed to register expired-waiting-list job: %w", err)
	}

	c.cron.Start()
	log.Println("[CRON] Cron job started successfully.")
	return nil
}

func (c *CronService) Stop(ctx context.Context) error {
	if c.cron == nil {
		log.Println("[CRON] Stop called but scheduler was never started; nothing to do.")
		return nil
	}

	log.Println("[CRON] Stopping cron job, waiting for running jobs to finish...")

	done := c.cron.Stop()

	select {
	case <-done.Done():
		log.Println("[CRON] Cron job stopped gracefully.")
		return nil
	case <-ctx.Done():
		log.Printf("[CRON WARN] Graceful shutdown timed out, abandoning in-flight jobs: %v", ctx.Err())
		return ctx.Err()
	}
}
