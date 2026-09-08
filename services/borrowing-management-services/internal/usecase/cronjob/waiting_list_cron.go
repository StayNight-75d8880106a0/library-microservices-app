package cronjob

import (
	"borrowing-management-services/internal/usecase"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/robfig/cron/v3"
)

type CronService struct {
	usecase usecase.WaitingListUsecaseInterface
	cron    *cron.Cron
}

func NewCronService(waitingListUsecase usecase.WaitingListUsecaseInterface, cron *cron.Cron) *CronService {
	return &CronService{
		usecase: waitingListUsecase,
		cron:    cron,
	}
}

func (c *CronService) Start() error {
	if c.cron == nil {
		return errors.New("cron scheduler is not initialized")
	}

	if _, err := c.cron.AddFunc("0 */15 * * * *", func() {
		log.Println("[CRON] Running ProcessExpiredNotified...")
		if err := c.usecase.ProcessExpiredWaitingLists(context.Background()); err != nil {
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
