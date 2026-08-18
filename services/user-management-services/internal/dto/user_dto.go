package dto

type UserUpdatePasswordRequest struct {
	NewPassword     *string `json:"newPassword"`
	ConfirmPassword *string `json:"confirmPassword"`
}

type UserUpdateProfileRequest struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

type CreateAdminRequest struct {
	Username  *string `json:"username"`
	Email     *string `json:"email" binding:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Password  *string `json:"password"`
}

type AdminResponse struct {
	ID            *string `json:"id"`
	Username      *string `json:"username"`
	FirstName     *string `json:"firstName"`
	LastName      *string `json:"lastName"`
	Email         *string `json:"email"`
	EmailVerified *bool   `json:"emailVerified"`
	CreatedAt     string  `json:"createdAt"`
}

type UserResponse struct {
	ID            *string `json:"id"`
	Username      *string `json:"username"`
	FirstName     *string `json:"firstName"`
	LastName      *string `json:"lastName"`
	Email         *string `json:"email"`
	EmailVerified *bool   `json:"emailVerified"`
	CreatedAt     string  `json:"createdAt"`
}
