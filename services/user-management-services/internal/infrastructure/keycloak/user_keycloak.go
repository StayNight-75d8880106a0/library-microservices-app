package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"user-management-services/internal/config"
	"user-management-services/internal/helper"
)

type KeycloakUserInterface interface {
	GetAdminToken(ctx context.Context) (string, error)
	UpdateProfile(ctx context.Context, token string, userID string, payload map[string]interface{}) error
	UpdatePassword(ctx context.Context, token string, userID string, newPassword string) error
	CreateUserAdmin(ctx context.Context, token string, payload map[string]interface{}) (string, error)
	AssignRoleToUserAdmin(ctx context.Context, token string, userID string, roleName string) error
	GetUserRoles(ctx context.Context, token string, userID string) ([]string, error)
	// UpdateUserProfile(ctx context.Context, token string, userID string, payload map[string]interface{}) error
	DeleteUser(ctx context.Context, token string, userID string) error
	GetUsers(ctx context.Context, token string, first int, max int) ([]map[string]interface{}, int64, error)
	countUsers(ctx context.Context, token string) (int64, error)
	GetMyProfile(ctx context.Context, token string, userID string) (map[string]interface{}, error)
}

type KeycloakClientUser struct {
	cfg *config.KeycloakConfig
}

func NewKeycloakClientUserRegistry(cfg *config.KeycloakConfig) *KeycloakClientUser {
	return &KeycloakClientUser{
		cfg: cfg,
	}
}

func (kc *KeycloakClientUser) GetAdminToken(ctx context.Context) (string, error) {

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

func (kc *KeycloakClientUser) UpdateProfile(ctx context.Context, token string, userID string, payload map[string]interface{}) error {

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users/%s", kc.cfg.KeycloakURL, kc.cfg.Realm, userID)

	bodyBytes, errMarshal := json.Marshal(payload)

	if errMarshal != nil {
		return helper.NewInternalServerError("An Error During Marshal Json!", helper.ErrorDetail{Detail: errMarshal.Error()})
	}

	request, errRequest := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewBuffer(bodyBytes))

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	if errRequest != nil {
		return helper.NewInternalServerError("An Error During Create Request!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return helper.NewInternalServerError("An Error During Request!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusOK {
		return helper.NewInternalServerError("Invalid credentials from keycloak!", helper.ErrorDetail{Detail: fmt.Sprintf("Response Status Code: %d", response.StatusCode)})
	}

	return nil
}

func (kc *KeycloakClientUser) UpdatePassword(ctx context.Context, token string, userID string, newPassword string) error {

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users/%s/reset-password", kc.cfg.KeycloakURL, kc.cfg.Realm, userID)

	payload := map[string]interface{}{
		"type":      "password",
		"value":     newPassword,
		"temporary": false,
	}

	bodyBytes, errMarshal := json.Marshal(payload)

	if errMarshal != nil {
		return helper.NewInternalServerError("An Error During Marshal Json!", helper.ErrorDetail{Detail: errMarshal.Error()})
	}

	request, errRequest := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewBuffer(bodyBytes))

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	if errRequest != nil {
		return helper.NewInternalServerError("An Error During Create Request!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return helper.NewInternalServerError("An Error During Request!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusOK {
		return helper.NewInternalServerError("Keycloak Reset Password Failed!", helper.ErrorDetail{Detail: fmt.Sprintf("keycloak reset password failed, status: %d", response.StatusCode)})
	}

	return nil
}

func (kc *KeycloakClientUser) CreateUserAdmin(ctx context.Context, token string, payload map[string]interface{}) (string, error) {

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users", kc.cfg.KeycloakURL, kc.cfg.Realm)

	bodyBytes, errMarshal := json.Marshal(payload)

	if errMarshal != nil {
		return "", helper.NewInternalServerError("An Error During Marshal Json!", helper.ErrorDetail{Detail: errMarshal.Error()})
	}

	request, errRequest := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(bodyBytes))

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	if errRequest != nil {
		return "", helper.NewInternalServerError("An Error During Create Request!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return "", helper.NewInternalServerError("An Error During Request!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return "", helper.NewInternalServerError("An Error During User Creation!", helper.ErrorDetail{Detail: fmt.Sprintf("Response Status Code: %d", response.StatusCode)})
	}

	location := response.Header.Get("location")

	parts := strings.Split(location, "/")

	return parts[len(parts)-1], nil

}

func (kc *KeycloakClientUser) AssignRoleToUserAdmin(ctx context.Context, token string, userID string, roleName string) error {

	roleEndpoint := fmt.Sprintf("%s/admin/realms/%s/roles/%s", kc.cfg.KeycloakURL, kc.cfg.Realm, roleName)

	requestRole, errRequestRole := http.NewRequestWithContext(ctx, "GET", roleEndpoint, nil)

	requestRole.Header.Set("Authorization", "Bearer "+token)

	if errRequestRole != nil {
		return helper.NewInternalServerError("An Error During Create Request!", helper.ErrorDetail{Detail: errRequestRole.Error()})
	}

	responseRole, errResponseRole := http.DefaultClient.Do(requestRole)

	if errResponseRole != nil {
		return helper.NewInternalServerError("An Error During Request!", helper.ErrorDetail{Detail: errResponseRole.Error()})
	}

	defer responseRole.Body.Close()

	if responseRole.StatusCode != http.StatusOK {
		return helper.NewInternalServerError("An Error During Get Role!", helper.ErrorDetail{Detail: fmt.Sprintf("Response Status Code: %d", responseRole.StatusCode)})
	}

	var roleDetail map[string]interface{}

	_ = json.NewDecoder(responseRole.Body).Decode(&roleDetail)

	roleID, _ := roleDetail["id"].(string)
	name, _ := roleDetail["name"].(string)

	if roleID == "" || name == "" {
		return helper.NewInternalServerError("An Error During Get Role!", helper.ErrorDetail{Detail: "Role ID or Name is empty"})
	}

	assignRoleEndpoint := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm", kc.cfg.KeycloakURL, kc.cfg.Realm, userID)

	payload := []map[string]interface{}{
		{
			"id":   roleID,
			"name": name,
		},
	}

	bodyBytes, errMarshal := json.Marshal(payload)

	if errMarshal != nil {
		return helper.NewInternalServerError("An Error During Marshal Json!", helper.ErrorDetail{Detail: errMarshal.Error()})
	}

	requestAssignRole, errRequestAssignRole := http.NewRequestWithContext(ctx, "POST", assignRoleEndpoint, bytes.NewBuffer(bodyBytes))

	requestAssignRole.Header.Set("Authorization", "Bearer "+token)
	requestAssignRole.Header.Set("Content-Type", "application/json")

	if errRequestAssignRole != nil {
		return helper.NewInternalServerError("An Error During Create Request!", helper.ErrorDetail{Detail: errRequestAssignRole.Error()})
	}

	responseAssignRole, errResponseAssignRole := http.DefaultClient.Do(requestAssignRole)

	if errResponseAssignRole != nil {
		return helper.NewInternalServerError("An Error During Request!", helper.ErrorDetail{Detail: errResponseAssignRole.Error()})
	}

	defer responseAssignRole.Body.Close()

	if responseAssignRole.StatusCode != http.StatusNoContent && responseAssignRole.StatusCode != http.StatusOK {
		return helper.NewInternalServerError("An Error During Assign Role!", helper.ErrorDetail{Detail: fmt.Sprintf("Response Status Code: %d", responseAssignRole.StatusCode)})
	}

	return nil
}

func (kc *KeycloakClientUser) GetUserRoles(ctx context.Context, token string, userID string) ([]string, error) {

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users/%s/role-mappings/realm", kc.cfg.KeycloakURL, kc.cfg.Realm, userID)

	request, errRequest := http.NewRequestWithContext(ctx, "GET", endpoint, nil)

	request.Header.Set("Authorization", "Bearer "+token)

	if errRequest != nil {
		return nil, helper.NewInternalServerError("An Error During Create Request!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return nil, helper.NewInternalServerError("An Error During Request!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	if response.StatusCode != http.StatusOK {
		return nil, helper.NewInternalServerError("An Error During Get User Roles!", helper.ErrorDetail{Detail: fmt.Sprintf("Response Status Code: %d", response.StatusCode)})
	}

	defer response.Body.Close()

	var rolesData []map[string]interface{}

	_ = json.NewDecoder(response.Body).Decode(&rolesData)

	roles := make([]string, 0)

	for _, r := range rolesData {
		if name, ok := r["name"].(string); ok {
			roles = append(roles, name)
		}
	}

	return roles, nil

}

// func (kc *KeycloakClientUser) UpdateUserProfile(ctx context.Context, token string, userID string, payload map[string]interface{}) error {

// 	endpoint := fmt.Sprintf("%s/admin/realms/%s/users/%s", kc.cfg.KeycloakURL, kc.cfg.Realm, userID)

// 	bodyBytes, errMarshal := json.Marshal(payload)

// 	if errMarshal != nil {
// 		return helper.NewInternalServerError("An Error During Marshal Json!", helper.ErrorDetail{Detail: errMarshal.Error()})
// 	}

// 	request, errRequest := http.NewRequestWithContext(ctx, "PUT", endpoint, bytes.NewBuffer(bodyBytes))

// 	request.Header.Set("Authorization", "Bearer "+token)
// 	request.Header.Set("Content-Type", "application/json")

// 	if errRequest != nil {
// 		return helper.NewInternalServerError("An Error During Create Request!", helper.ErrorDetail{Detail: errRequest.Error()})
// 	}

// 	response, errResponse := http.DefaultClient.Do(request)

// 	if errResponse != nil {
// 		return helper.NewInternalServerError("An Error During Request!", helper.ErrorDetail{Detail: errResponse.Error()})
// 	}

// 	if response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusOK {
// 		return helper.NewInternalServerError("An Error During Update User Profile!", helper.ErrorDetail{Detail: fmt.Sprintf("Response Status Code: %d", response.StatusCode)})
// 	}

// 	return nil

// }

func (kc *KeycloakClientUser) DeleteUser(ctx context.Context, token string, userID string) error {

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users/%s", kc.cfg.KeycloakURL, kc.cfg.Realm, userID)

	request, errRequest := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)

	request.Header.Set("Authorization", "Bearer "+token)

	if errRequest != nil {
		return helper.NewInternalServerError("An Error During Create Request!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return helper.NewInternalServerError("An Error During Request!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusOK {
		return helper.NewInternalServerError("An Error During Delete User!", helper.ErrorDetail{Detail: fmt.Sprintf("Response Status Code: %d", response.StatusCode)})
	}

	return nil

}

func (kc *KeycloakClientUser) GetUsers(ctx context.Context, token string, first int, max int) ([]map[string]interface{}, int64, error) {

	total, errCount := kc.countUsers(ctx, token)

	if errCount != nil {
		return nil, 0, helper.NewInternalServerError("An Error During Count Users!", helper.ErrorDetail{Detail: errCount.Error()})
	}

	if total == 0 {
		return []map[string]interface{}{}, 0, nil
	}

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users?first=%d&max=%d", kc.cfg.KeycloakURL, kc.cfg.Realm, first, max)

	request, errRequest := http.NewRequestWithContext(ctx, "GET", endpoint, nil)

	request.Header.Set("Authorization", "Bearer "+token)

	if errRequest != nil {
		return nil, 0, helper.NewInternalServerError("An Error During Create Request!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return nil, 0, helper.NewInternalServerError("An Error During Request!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, 0, helper.NewInternalServerError("An Error During Get Users!", helper.ErrorDetail{Detail: fmt.Sprintf("Response Status Code: %d", response.StatusCode)})
	}

	var users []map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&users); err != nil {
		return nil, 0, err
	}

	return users, total, nil

}

func (kc *KeycloakClientUser) countUsers(ctx context.Context, token string) (int64, error) {

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users/count", kc.cfg.KeycloakURL, kc.cfg.Realm)

	request, errRequest := http.NewRequestWithContext(ctx, "GET", endpoint, nil)

	request.Header.Set("Authorization", "Bearer "+token)

	if errRequest != nil {
		return 0, helper.NewInternalServerError("An Error During Create Request!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return 0, helper.NewInternalServerError("An Error During Request!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, helper.NewInternalServerError("An Error During Count Users!", helper.ErrorDetail{Detail: fmt.Sprintf("Response Status Code: %d", response.StatusCode)})
	}

	var count int64
	if err := json.NewDecoder(response.Body).Decode(&count); err != nil {
		return 0, err
	}

	return count, nil

}

func (kc *KeycloakClientUser) GetMyProfile(ctx context.Context, token string, userID string) (map[string]interface{}, error) {

	endpoint := fmt.Sprintf("%s/admin/realms/%s/users/%s", kc.cfg.KeycloakURL, kc.cfg.Realm, userID)

	request, errRequest := http.NewRequestWithContext(ctx, "GET", endpoint, nil)

	request.Header.Set("Authorization", "Bearer "+token)

	if errRequest != nil {
		return nil, helper.NewInternalServerError("An Error During Create Request!", helper.ErrorDetail{Detail: errRequest.Error()})
	}

	response, errResponse := http.DefaultClient.Do(request)

	if errResponse != nil {
		return nil, helper.NewInternalServerError("An Error During Request!", helper.ErrorDetail{Detail: errResponse.Error()})
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, helper.NewInternalServerError("An Error During Get My Profile!", helper.ErrorDetail{Detail: fmt.Sprintf("Response Status Code: %d", response.StatusCode)})
	}

	var profile map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&profile); err != nil {
		return nil, err
	}

	return profile, nil
}
