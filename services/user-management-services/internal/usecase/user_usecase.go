package usecase

import (
	"context"
	"user-management-services/internal/dto"
	"user-management-services/internal/helper"
	"user-management-services/internal/infrastructure/keycloak"
	"user-management-services/internal/models"
	"user-management-services/internal/repository"
)

type UserProfileUsecaseInterface interface {
	UpdateMyProfile(ctx context.Context, userID string, request *dto.UserUpdateProfileRequest) error
	UpdateMyPassword(ctx context.Context, userID string, request *dto.UserUpdatePasswordRequest) error
	GetMyProfile(ctx context.Context, userID string) (*dto.UserResponse, error)
	CreateUserFromEvent(ctx context.Context, payload *dto.UserCreatedKafkaPayloadConsumer) error
}

type UserProfileUsecase struct {
	keycloak   keycloak.KeycloakUserInterface
	repository repository.UserRepositoryInterface
}

func NewUserProfileUsecaseRegistry(keycloakCLient keycloak.KeycloakUserInterface, userRepository repository.UserRepositoryInterface) *UserProfileUsecase {
	return &UserProfileUsecase{
		keycloak:   keycloakCLient,
		repository: userRepository,
	}
}

func (u *UserProfileUsecase) UpdateMyProfile(ctx context.Context, userID string, request *dto.UserUpdateProfileRequest) error {

	if request.FirstName == nil || *request.FirstName == "" {
		return helper.NewUnprocessableEntityError("First Name Cannot Be Empty!", helper.ErrorDetail{Detail: "First Name Is Required!"})
	}

	if request.LastName == nil || *request.LastName == "" {
		return helper.NewUnprocessableEntityError("Last Name Cannot Be Empty!", helper.ErrorDetail{Detail: "Last Name Is Required!"})
	}

	keycloakToken, errToken := u.keycloak.GetAdminToken(ctx)

	if errToken != nil {
		return helper.NewInternalServerError("Failed to Get Admin Token!", helper.ErrorDetail{Detail: errToken.Error()})
	}

	payload := map[string]interface{}{
		"firstName": *request.FirstName,
		"lastName":  *request.LastName,
	}

	errUpdateProfile := u.keycloak.UpdateProfile(ctx, keycloakToken, userID, payload)

	if errUpdateProfile != nil {
		return errUpdateProfile
	}

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
