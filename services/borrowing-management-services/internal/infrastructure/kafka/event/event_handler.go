package event

import (
	"borrowing-management-services/internal/client"
	"borrowing-management-services/internal/dto"
	"borrowing-management-services/internal/repository"
	"context"
	"encoding/json"
	"log"
	"time"
)

type EventHandler struct {
	cacheRepo  repository.UserRedisCacheInterface
	grpcClient client.UserGrpcClientInterface
}

func NewEventHandler(cacheRepo repository.UserRedisCacheInterface, grpcClient client.UserGrpcClientInterface) *EventHandler {
	return &EventHandler{
		cacheRepo:  cacheRepo,
		grpcClient: grpcClient,
	}
}

func (h *EventHandler) HandleUserAuthEvent(ctx context.Context, messageBytes []byte) error {

	var event dto.UserAuthKafkaPayloadConsumer

	errUnmarshal := json.Unmarshal(messageBytes, &event)

	if errUnmarshal != nil {
		log.Printf("[Kafka Handler Error] Failed to unmarshal Auth Event: %v", errUnmarshal)
		return errUnmarshal
	}

	switch event.EventType {
	case "USER_LOGIN":
		log.Printf("📥 [Event USER_LOGIN] Processing user_id: %s", event.KeycloakID)

		userStatus, errGrpc := h.grpcClient.GetUserStatus(ctx, event.KeycloakID)

		if errGrpc != nil {
			log.Printf("[gRPC Error] Failed to get user status for user_id %s: %v", event.KeycloakID, errGrpc)
			return errGrpc
		}

		errCache := h.cacheRepo.SetUserStatus(ctx, event.KeycloakID, userStatus, 24*time.Hour)

		if errCache != nil {
			log.Printf("[Redis Cache Error] Failed to set user status for user_id %s: %v", event.KeycloakID, errCache)
			return errCache
		}
	case "USER_LOGOUT":
		log.Printf("📥 [Event USER_LOGOUT] Processing user_id: %s", event.KeycloakID)

		errCache := h.cacheRepo.DeleteUserStatus(ctx, event.KeycloakID)

		if errCache != nil {
			log.Printf("[Redis Cache Error] Failed to delete user status for user_id %s: %v", event.KeycloakID, errCache)
			return errCache
		}
		log.Printf("✅ [Event USER_LOGOUT] Successfully deleted user status for user_id: %s", event.KeycloakID)
	default:
		log.Printf("[Kafka Handler Error] Unknown event type: %s", event.EventType)
		return nil
	}

	return nil

}

func (h *EventHandler) HandleUserStatusUpdateEvent(ctx context.Context, messageBytes []byte) error {

	var event dto.UserStatusUpdatedKafkaPayloadConsumer

	errUnmarshal := json.Unmarshal(messageBytes, &event)

	if errUnmarshal != nil {
		log.Printf("[Kafka Handler Error] Failed to unmarshal User Status Update Event: %v", errUnmarshal)
		return errUnmarshal
	}

	log.Printf("📥 [Event USER_STATUS_UPDATED] User: %s, New Status: %s", event.UserID, event.Status)

	errCache := h.cacheRepo.SetUserStatus(ctx, event.UserID, event.Status, 24*time.Hour)

	if errCache != nil {
		log.Printf("[Redis Cache Error] Failed to set user status for user_id %s: %v", event.UserID, errCache)
		return errCache
	}

	log.Printf("✅ [Event USER_STATUS_UPDATED] Successfully updated user status for user_id: %s", event.UserID)

	return nil

}
