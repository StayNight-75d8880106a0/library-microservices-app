package usecase

import (
	"auth-services/internal/dto"
	"auth-services/internal/helper"
	"auth-services/internal/infrastructure/keycloak"
	"auth-services/internal/repository"
	"context"
	"time"
)

type AuthUsecaseInterface interface {
	Login(ctx context.Context, request *dto.LoginRequest) (*dto.LoginResponse, error)
	RegisterUser(ctx context.Context, request *dto.RegisterUser) error
	Logout(ctx context.Context, token string, request *dto.RefreshTokenRequest) error
	Me(ctx context.Context, claims map[string]interface{}) (*dto.MeResponse, error)
	RefreshToken(ctx context.Context, request *dto.RefreshTokenRequest) (*dto.LoginResponse, error)
}

type AuthUsecase struct {
	keycloak   keycloak.KeycloakClientInterface
	repository repository.AuthRepositoryInterface
}

func NewAuthUsecaseRegistry(keycloakClient keycloak.KeycloakClientInterface, authRepository repository.AuthRepository) *AuthUsecase {
	return &AuthUsecase{
		keycloak:   keycloakClient,
		repository: &authRepository,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, request *dto.LoginRequest) (*dto.LoginResponse, error) {

	if request.Username == nil || *request.Username == "" {
		return nil, helper.NewUnprocessableEntityError("Username Cannot Be Empty!", helper.ErrorDetail{Detail: "Username is required!"})
	}

	if request.Password == nil || *request.Password == "" {
		return nil, helper.NewUnprocessableEntityError("Password Cannot Be Empty!", helper.ErrorDetail{Detail: "Password is required!"})
	}

	response, errReponse := u.keycloak.Login(ctx, *request.Username, *request.Password)

	if errReponse != nil {
		return nil, errReponse
	}

	result := &dto.LoginResponse{
		AccessToken:           response["access_token"].(string),
		RefreshToken:          response["refresh_token"].(string),
		AccessTokenExpiresIn:  response["expires_in"].(float64),
		RefreshTokenExpiresIn: response["refresh_expires_in"].(float64),
		TokenType:             "Bearer",
	}

	return result, nil

}

func (u *AuthUsecase) RegisterUser(ctx context.Context, request *dto.RegisterUser) error {

	if request.Username == nil || *request.Username == "" {
		return helper.NewUnprocessableEntityError("Username Cannot Be Empty!", helper.ErrorDetail{Detail: "Username is required!"})
	}

	if request.Password == nil || *request.Password == "" {
		return helper.NewUnprocessableEntityError("Password Cannot Be Empty!", helper.ErrorDetail{Detail: "Password is required!"})
	}

	if request.Email == nil || *request.Email == "" {
		return helper.NewUnprocessableEntityError("Email Cannot Be Empty!", helper.ErrorDetail{Detail: "Email is required!"})
	}

	if request.FirstName == nil || *request.FirstName == "" {
		return helper.NewUnprocessableEntityError("First Name Cannot Be Empty!", helper.ErrorDetail{Detail: "First Name is required!"})
	}

	if request.LastName == nil || *request.LastName == "" {
		return helper.NewUnprocessableEntityError("Last Name Cannot Be Empty!", helper.ErrorDetail{Detail: "Last Name is required!"})
	}

	adminToken, errAdminToken := u.keycloak.GetAdminToken(ctx)

	if errAdminToken != nil {
		return errAdminToken
	}

	payload := map[string]interface{}{
		"username":  *request.Username,
		"email":     *request.Email,
		"firstName": *request.FirstName,
		"lastName":  *request.LastName,
		"enabled":   true,
		"credentials": []map[string]interface{}{
			{
				"type":      "password",
				"value":     *request.Password,
				"temporary": false,
			},
		},
	}

	result := u.keycloak.RegisterUser(ctx, adminToken, payload)

	return result

}

func (u *AuthUsecase) Logout(ctx context.Context, token string, request *dto.RefreshTokenRequest) error {

	if request.RefreshToken == nil || *request.RefreshToken == "" {
		return helper.NewUnprocessableEntityError("Refresh Token Cannot Be Empty!", helper.ErrorDetail{Detail: "Refresh token is required!"})
	}

	errLogoutKeycloak := u.keycloak.Logout(ctx, *request.RefreshToken)

	if errLogoutKeycloak != nil {
		return errLogoutKeycloak
	}

	errLogoutRedis := u.repository.BlacklistToken(ctx, token, 1*time.Hour)

	if errLogoutRedis != nil {
		return helper.NewInternalServerError("Failed to logout user!", helper.ErrorDetail{Detail: errLogoutRedis.Error()})
	}

	return nil
}

func (u *AuthUsecase) Me(ctx context.Context, claims map[string]interface{}) (*dto.MeResponse, error) {

	sub, _ := claims["sub"].(string)
	username, _ := claims["preferred_username"].(string)
	name, _ := claims["name"].(string)
	email, _ := claims["email"].(string)

	role := helper.ExtractRoleFromClaims(claims)

	result := &dto.MeResponse{
		ID:       sub,
		Username: username,
		Email:    email,
		Name:     name,
		Roles:    role,
	}

	return result, nil

}

func (u *AuthUsecase) RefreshToken(ctx context.Context, request *dto.RefreshTokenRequest) (*dto.LoginResponse, error) {

	if request.RefreshToken == nil || *request.RefreshToken == "" {
		return nil, helper.NewUnprocessableEntityError("Refresh Token Cannot Be Empty!", helper.ErrorDetail{Detail: "Refresh token is required!"})
	}

	response, errResponse := u.keycloak.RefreshToken(ctx, *request.RefreshToken)

	if errResponse != nil {
		return nil, errResponse
	}

	result := &dto.LoginResponse{
		AccessToken:           response.AccessToken,
		RefreshToken:          response.RefreshToken,
		AccessTokenExpiresIn:  response.AccessTokenExpiresIn,
		RefreshTokenExpiresIn: response.RefreshTokenExpiresIn,
		TokenType:             "Bearer",
	}

	return result, nil
}
