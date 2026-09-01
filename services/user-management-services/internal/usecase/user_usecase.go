package usecase

import (
	"context"
	"log"
	"time"
	"user-management-services/internal/config"
	"user-management-services/internal/dto"
	"user-management-services/internal/helper"
	"user-management-services/internal/infrastructure/kafka/event"
	"user-management-services/internal/infrastructure/kafka/producer"
	"user-management-services/internal/infrastructure/keycloak"
	"user-management-services/internal/models"
	"user-management-services/internal/repository"

	"gorm.io/gorm"
)

type UserProfileUsecaseInterface interface {
	UpdateMyProfile(ctx context.Context, userID string, request *dto.UserUpdateProfileRequest) error
	UpdateMyPassword(ctx context.Context, userID string, request *dto.UserUpdatePasswordRequest) error
	GetMyProfile(ctx context.Context, userID string) (*dto.UserResponse, error)
	CreateUserFromEvent(ctx context.Context, payload *dto.UserCreatedKafkaPayloadConsumer) error
	GetUserStatus(ctx context.Context, userID string) (string, error)
}

type UserProfileUsecase struct {
	keycloak      keycloak.KeycloakUserInterface
	repository    repository.UserRepositoryInterface
	kafkaProducer *producer.KafkaProducer
	cfg           *config.AppConfig
}

func NewUserProfileUsecaseRegistry(keycloakCLient keycloak.KeycloakUserInterface, userRepository repository.UserRepositoryInterface, kafkaProducer *producer.KafkaProducer, cfg *config.AppConfig) *UserProfileUsecase {
	return &UserProfileUsecase{
		keycloak:      keycloakCLient,
		repository:    userRepository,
		kafkaProducer: kafkaProducer,
		cfg:           cfg,
	}
}

func (u *UserProfileUsecase) UpdateMyProfile(ctx context.Context, userID string, request *dto.UserUpdateProfileRequest) error {

	if request.FirstName == nil || *request.FirstName == "" {
		return helper.NewUnprocessableEntityError("First Name Cannot Be Empty!", helper.ErrorDetail{Detail: "First Name Is Required!"})
	}

	if request.LastName == nil || *request.LastName == "" {
		return helper.NewUnprocessableEntityError("Last Name Cannot Be Empty!", helper.ErrorDetail{Detail: "Last Name Is Required!"})
	}

	if request.PhoneNumber == nil || *request.PhoneNumber == "" {
		return helper.NewUnprocessableEntityError("Phone Number Cannot Be Empty!", helper.ErrorDetail{Detail: "Phone Number Is Required!"})
	}

	if request.Address == nil || *request.Address == "" {
		return helper.NewUnprocessableEntityError("Address Cannot Be Empty!", helper.ErrorDetail{Detail: "Address Is Required!"})
	}

	errValidatePhone := helper.ValidateIndonesianPhoneNumber(*request.PhoneNumber)

	if errValidatePhone != nil {
		return helper.NewUnprocessableEntityError("Invalid Phone Number!", helper.ErrorDetail{Detail: errValidatePhone.Error()})
	}

	userDB, errGetUser := u.repository.GetUserByKeycloakID(ctx, userID)

	if errGetUser != nil {
		if errGetUser == gorm.ErrRecordNotFound {
			return helper.NewNotFoundError("User Not Found!", helper.ErrorDetail{Detail: "User With ID " + userID + " Not Found!"})
		}
		return helper.NewInternalServerError("Failed to Get User!", helper.ErrorDetail{Detail: errGetUser.Error()})
	}

	keycloakToken, errToken := u.keycloak.GetAdminToken(ctx)

	if errToken != nil {
		return helper.NewInternalServerError("Failed to Get Admin Token!", helper.ErrorDetail{Detail: errToken.Error()})
	}

	payload := map[string]interface{}{
		"firstName": *request.FirstName,
		"lastName":  *request.LastName,
	}

	errUpdateProfileKeycloak := u.keycloak.UpdateProfile(ctx, keycloakToken, userID, payload)

	if errUpdateProfileKeycloak != nil {
		return errUpdateProfileKeycloak
	}

	userDB.FirstName = *request.FirstName
	userDB.LastName = *request.LastName
	userDB.PhoneNumber = *request.PhoneNumber
	userDB.Address = *request.Address
	userDB.UpdatedAt = time.Now()

	if userDB.ProfileStatus == models.UserStatusIncomplete {
		userDB.ProfileStatus = models.UserStatusActive
	}

	errUpdateProfileDB := u.repository.UpdateUserProfileDB(ctx, userDB, userID)

	if errUpdateProfileDB != nil {
		return helper.NewInternalServerError("Failed to Update User Profile!", helper.ErrorDetail{Detail: errUpdateProfileDB.Error()})
	}

	event := &event.UserUpdatedEvent{
		EventType: "USER_STATUS_UPDATED",
		UserID:    userID,
		Status:    string(userDB.ProfileStatus),
		UpdatedAt: time.Now(),
	}

	go func() {
		errPublish := u.kafkaProducer.PublishEvent(context.Background(), event, userID, u.cfg.Kafka.TopicUserUpdated)
		if errPublish != nil {
			log.Printf("[Kafka Publish Error] Failed to send event for user %s: %v", userID, errPublish)
		}
	}()

	return nil
}

func (u *UserProfileUsecase) UpdateMyPassword(ctx context.Context, userID string, request *dto.UserUpdatePasswordRequest) error {

	if request.NewPassword == nil || *request.NewPassword == "" {
		return helper.NewUnprocessableEntityError("New Password Cannot Be Empty!", helper.ErrorDetail{Detail: "New Password Is Required"})
	}

	if request.ConfirmPassword == nil || *request.ConfirmPassword == "" {
		return helper.NewUnprocessableEntityError("Confirm Password Cannot Be Empty!", helper.ErrorDetail{Detail: "Confirm Password Is Required"})
	}

	if *request.NewPassword != *request.ConfirmPassword {
		return helper.NewUnprocessableEntityError("New Password and Confirm Password Is Not Matching", helper.ErrorDetail{Detail: "New Password and Confirm Password Must Match!"})
	}

	keycloakToken, errToken := u.keycloak.GetAdminToken(ctx)

	if errToken != nil {
		return helper.NewInternalServerError("Failed to Get Admin Token!", helper.ErrorDetail{Detail: errToken.Error()})
	}

	errUpdatePassword := u.keycloak.UpdatePassword(ctx, keycloakToken, userID, *request.ConfirmPassword)

	if errUpdatePassword != nil {
		return errUpdatePassword
	}

	return nil

}

func (u *UserProfileUsecase) GetMyProfile(ctx context.Context, userID string) (*dto.UserResponse, error) {

	keycloakToken, errToken := u.keycloak.GetAdminToken(ctx)

	if errToken != nil {
		return nil, helper.NewInternalServerError("Failed to Get Admin Token!", helper.ErrorDetail{Detail: errToken.Error()})
	}

	profile, errGetProfile := u.keycloak.GetMyProfile(ctx, keycloakToken, userID)

	if errGetProfile != nil {
		return nil, errGetProfile
	}

	profileDB, errGetProfileDB := u.repository.GetUserByKeycloakID(ctx, userID)

	if errGetProfileDB != nil {
		if errGetProfileDB == gorm.ErrRecordNotFound {
			return nil, helper.NewNotFoundError("User Not Found!", helper.ErrorDetail{Detail: "User With ID " + userID + " Not Found!"})
		}
		return nil, helper.NewInternalServerError("Failed to Get User!", helper.ErrorDetail{Detail: errGetProfileDB.Error()})
	}

	id, _ := profile["id"].(string)
	username, _ := profile["username"].(string)
	firstName, _ := profile["firstName"].(string)
	lastName, _ := profile["lastName"].(string)
	email, _ := profile["email"].(string)
	emailVerified, _ := profile["emailVerified"].(bool)
	createdAt, _ := profile["createdTimestamp"].(float64)

	result := &dto.UserResponse{
		ID:            &id,
		Username:      &username,
		FirstName:     &firstName,
		LastName:      &lastName,
		Email:         &email,
		EmailVerified: &emailVerified,
		CreatedAt:     helper.FormatEpochMillisRFC3339Jakarta(int64(createdAt)),
	}

	if profileDB != nil {
		result.PhoneNumber = &profileDB.PhoneNumber
		result.LibraryCardNumber = &profileDB.LibraryCardNumber
		result.Address = &profileDB.Address
		result.ProfileStatus = string(profileDB.ProfileStatus)
		result.UpdatedAt = helper.FormatEpochMillisRFC3339Jakarta(profileDB.UpdatedAt.UnixNano() / int64(time.Millisecond))
	}

	return result, nil
}

func (u *UserProfileUsecase) CreateUserFromEvent(ctx context.Context, payload *dto.UserCreatedKafkaPayloadConsumer) error {

	result := &models.Users{
		KeycloakUserID:    payload.KeycloakID,
		FirstName:         payload.FirstName,
		LastName:          payload.LastName,
		LibraryCardNumber: helper.GenerateLibraryCardNumber(),
		ProfileStatus:     "INCOMPLETE",
		CreatedAt:         payload.CreatedAt,
	}

	errCreate := u.repository.CreateUserFromEvent(ctx, result)

	if errCreate != nil {
		return helper.NewInternalServerError("Failed to Create User From Event!", helper.ErrorDetail{Detail: errCreate.Error()})
	}

	return nil
}

func (u *UserProfileUsecase) GetUserStatus(ctx context.Context, userID string) (string, error) {

	userDB, errGetUser := u.repository.GetUserByKeycloakID(ctx, userID)

	if errGetUser != nil {
		if errGetUser == gorm.ErrRecordNotFound {
			return "", helper.NewNotFoundError("User Not Found!", helper.ErrorDetail{Detail: "User With ID " + userID + " Not Found!"})
		}
		return "", helper.NewInternalServerError("Failed to Get User!", helper.ErrorDetail{Detail: errGetUser.Error()})
	}

	return string(userDB.ProfileStatus), nil

}
