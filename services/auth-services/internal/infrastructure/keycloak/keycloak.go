package keycloak

import (
	"auth-services/internal/config"
	"auth-services/internal/dto"
	"auth-services/internal/helper"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type KeycloakClientInterface interface {
	Login(ctx context.Context, username string, password string) (map[string]interface{}, error)
	GetAdminToken(ctx context.Context) (string, error)
	RegisterUser(ctx context.Context, adminToken string, payload map[string]interface{}) error
	RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

type KeycloakClient struct {
	cfg *config.KeycloakConfig
}

func NewKeycloakClientRegistry(keycloakConfig *config.KeycloakConfig) *KeycloakClient {
	return &KeycloakClient{
		cfg: keycloakConfig,
	}
}

func (kc *KeycloakClient) Login(ctx context.Context, username string, password string) (map[string]interface{}, error) {

	endpoint := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", kc.cfg.KeycloakURL, kc.cfg.Realm)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", kc.cfg.ClientID)
	data.Set("client_secret", kc.cfg.ClientSecret)
	data.Set("username", username)
	data.Set("password", password)

	request, errRequest := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))

	if errRequest != nil {
		return nil, helper.NewInternalServerError("An Error During Login Request To Keycloak", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return nil, helper.NewInternalServerError("An Error During Login Request To Keycloak", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, helper.NewUnauthorizedError("Username or password is incorrect!", helper.ErrorDetail{Detail: "Invalid credentials from keycloak!"})
	}

	var result map[string]interface{}

	json.NewDecoder(response.Body).Decode(&result)

	return result, nil
}

func (kc *KeycloakClient) GetAdminToken(ctx context.Context) (string, error) {

	endpoint := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", kc.cfg.KeycloakURL, kc.cfg.Realm)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", kc.cfg.ClientID)
	data.Set("client_secret", kc.cfg.ClientSecret)

	request, errRequest := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))

	if errRequest != nil {
		return "", helper.NewInternalServerError("Failed to create admin token request to keycloak!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return "", helper.NewInternalServerError("Failed to get admin token from keycloak!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	var result map[string]interface{}

	json.NewDecoder(response.Body).Decode(&result)

	return result["access_token"].(string), nil

}

func (kc *KeycloakClient) RegisterUser(ctx context.Context, adminToken string, payload map[string]interface{}) error {

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users", kc.cfg.KeycloakURL, kc.cfg.Realm)

	jsonData, errJson := json.Marshal(payload)

	if errJson != nil {
		return helper.NewInternalServerError("Failed to marshal payload for user registration!", helper.ErrorDetail{Detail: errJson.Error()})
	}

	request, errRequets := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	request.Header.Set("Authorization", "Bearer "+adminToken)
	request.Header.Set("Content-Type", "application/json")

	if errRequets != nil {
		return helper.NewInternalServerError("Failed to create user registration request to keycloak!", helper.ErrorDetail{Detail: errRequets.Error()})
	}

	httpClient := &http.Client{}

	response, errResponse := httpClient.Do(request)

	if errResponse != nil {
		return helper.NewInternalServerError("Failed to send user registration request to keycloak!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(response.Body)
		return helper.NewInternalServerError("Failed to register user in keycloak!", helper.ErrorDetail{Detail: string(bodyBytes)})
	}

	return nil
}

func (kc *KeycloakClient) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", kc.cfg.KeycloakURL, kc.cfg.Realm)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", kc.cfg.ClientID)
	data.Set("client_secret", kc.cfg.ClientSecret)
	data.Set("refresh_token", refreshToken)

	request, errRequest := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))

	if errRequest != nil {
		return nil, helper.NewInternalServerError("Failed to create refresh token request to keycloak!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return nil, helper.NewInternalServerError("Failed to send refresh token request to keycloak!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, helper.NewUnauthorizedError("Invalid refresh token from keycloak!", helper.ErrorDetail{Detail: "Refresh token is invalid or expired!"})
	}

	var tokenResponse dto.KeycloakTokenResponse

	errDecode := json.NewDecoder(response.Body).Decode(&tokenResponse)

	if errDecode != nil {
		return nil, helper.NewInternalServerError("Failed to decode refresh token response from keycloak!", helper.ErrorDetail{Detail: errDecode.Error()})
	}

	result := &dto.LoginResponse{
		AccessToken:           tokenResponse.AccessToken,
		RefreshToken:          tokenResponse.RefreshToken,
		AccessTokenExpiresIn:  tokenResponse.ExpiresIn,
		RefreshTokenExpiresIn: tokenResponse.RefreshExpiresIn,
		TokenType:             tokenResponse.TokenType,
	}

	return result, nil

}

func (kc *KeycloakClient) Logout(ctx context.Context, refreshToken string) error {

	logoutURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/logout", kc.cfg.KeycloakURL, kc.cfg.Realm)

	data := url.Values{}
	data.Set("client_id", kc.cfg.ClientID)
	data.Set("client_secret", kc.cfg.ClientSecret)
	data.Set("refresh_token", refreshToken)

	request, errRequest := http.NewRequestWithContext(ctx, "POST", logoutURL, strings.NewReader(data.Encode()))

	if errRequest != nil {
		return helper.NewInternalServerError("Failed to create logout request to keycloak!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return helper.NewInternalServerError("Failed to send logout request to keycloak!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return helper.NewUnauthorizedError("Failed to revoke session on Keycloak!", helper.ErrorDetail{Detail: "Refresh token might be invalid or already revoked!"})
	}

	return nil

}
