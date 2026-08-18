package usecase

import (
	"context"
	"math"
	"user-management-services/internal/dto"
	"user-management-services/internal/helper"
	"user-management-services/internal/infrastructure/keycloak"
)

type AdminUsecaseInterface interface {
	GetAllUsers(ctx context.Context, page int, limit int) ([]dto.AdminResponse, helper.PaginationMeta, error)
	CreateAdmin(ctx context.Context, request *dto.CreateAdminRequest) error
	UpdateAdmin(ctx context.Context, adminID string, request *dto.UserUpdatePasswordRequest) error
	DeleteAdmin(ctx context.Context, adminID string) error
}

type AdminUsecase struct {
	keycloak keycloak.KeycloakUserInterface
}

func NewAdminUsecaseRegistry(keycloakClient keycloak.KeycloakUserInterface) *AdminUsecase {
	return &AdminUsecase{
		keycloak: keycloakClient,
	}
}

func (u *AdminUsecase) GetAllUsers(ctx context.Context, page int, limit int) ([]dto.AdminResponse, helper.PaginationMeta, error) {

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	keycloakToken, errToken := u.keycloak.GetAdminToken(ctx)

	if errToken != nil {
		return []dto.AdminResponse{}, helper.PaginationMeta{}, helper.NewInternalServerError("Failed to Get Admin Token!", helper.ErrorDetail{Detail: errToken.Error()})
	}

	usersData, totalData, errGetUsers := u.keycloak.GetUsers(ctx, keycloakToken, offset, limit)

	if errGetUsers != nil {
		return []dto.AdminResponse{}, helper.PaginationMeta{}, errGetUsers
	}

	totalPage := int(math.Ceil(float64(totalData) / float64(limit)))

	pagination := &helper.PaginationMeta{
		TotalData: totalData,
		TotalPage: totalPage,
		Page:      page,
		Limit:     limit,
		Keywords:  nil,
	}

	var users []dto.AdminResponse

	for _, value := range usersData {
		id, _ := value["id"].(string)
		username, _ := value["username"].(string)
		firstName, _ := value["firstName"].(string)
		lastName, _ := value["lastName"].(string)
		email, _ := value["email"].(string)
		emailVerified, _ := value["emailVerified"].(bool)
		createdAt, _ := value["createdTimestamp"].(float64)

		users = append(users, dto.AdminResponse{
			ID:            &id,
			Username:      &username,
			FirstName:     &firstName,
			LastName:      &lastName,
			Email:         &email,
			EmailVerified: &emailVerified,
			CreatedAt:     helper.FormatEpochMillisRFC3339Jakarta(int64(createdAt)),
		})
	}

	return users, *pagination, nil

}

func (u *AdminUsecase) CreateAdmin(ctx context.Context, request *dto.CreateAdminRequest) error {

	keycloakToken, errToken := u.keycloak.GetAdminToken(ctx)

	if errToken != nil {
		return helper.NewInternalServerError("Failed to Get Admin Token!", helper.ErrorDetail{Detail: errToken.Error()})
	}

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

	payload := map[string]interface{}{
		"username":      *request.Username,
		"email":         *request.Email,
		"firstName":     *request.FirstName,
		"lastName":      *request.LastName,
		"enabled":       true,
		"emailVerified": true,
		"credentials": []map[string]interface{}{
			{
				"type":      "password",
				"value":     *request.Password,
				"temporary": false,
			},
		},
	}

	newUser, errCreateUser := u.keycloak.CreateUserAdmin(ctx, keycloakToken, payload)

	if errCreateUser != nil {
		return errCreateUser
	}

	errAssignRole := u.keycloak.AssignRoleToUserAdmin(ctx, keycloakToken, newUser, "ADMIN")

	if errAssignRole != nil {
		return errAssignRole
	}

	return nil

}

func (u *AdminUsecase) UpdateAdmin(ctx context.Context, adminID string, request *dto.UserUpdatePasswordRequest) error {

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

	roles, errRoles := u.keycloak.GetUserRoles(ctx, keycloakToken, adminID)

	if errRoles != nil {
		return errRoles
	}

	if !helper.IsRoleMatch(roles, "ADMIN") || helper.IsRoleMatch(roles, "USER") {
		return helper.NewForbiddenError("Super Admin Cannot Update User Public Account!", helper.ErrorDetail{Detail: "This Forbidden!"})
	}

	errUpdate := u.keycloak.UpdatePassword(ctx, keycloakToken, adminID, *request.ConfirmPassword)

	if errUpdate != nil {
		return errUpdate
	}

	return nil
}

func (u *AdminUsecase) DeleteAdmin(ctx context.Context, adminID string) error {

	keycloakToken, errToken := u.keycloak.GetAdminToken(ctx)

	if errToken != nil {
		return helper.NewInternalServerError("Failed to Get Admin Token!", helper.ErrorDetail{Detail: errToken.Error()})
	}

	roles, errRoles := u.keycloak.GetUserRoles(ctx, keycloakToken, adminID)

	if errRoles != nil {
		return errRoles
	}

	if !helper.IsRoleMatch(roles, "ADMIN") || helper.IsRoleMatch(roles, "USER") {
		return helper.NewForbiddenError("Super Admin Cannot Update User Public Account!", helper.ErrorDetail{Detail: "This Forbidden!"})
	}

	errDelete := u.keycloak.DeleteUser(ctx, keycloakToken, adminID)

	if errDelete != nil {
		return errDelete
	}

	return nil

}
