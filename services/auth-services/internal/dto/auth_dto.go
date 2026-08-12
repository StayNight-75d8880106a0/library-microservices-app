package dto

type RegisterUser struct {
	Username  *string `json:"username" binding:"required"`
	Email     *string `json:"email" binding:"required"`
	FirstName *string `json:"firstName" binding:"required"`
	LastName  *string `json:"lastName" binding:"required"`
	Password  *string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username *string `json:"username" binding:"required"`
	Password *string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken           string  `json:"accessToken"`
	RefreshToken          string  `json:"refreshToken"`
	AccessTokenExpiresIn  float64 `json:"accessTokenExpiresIn"`
	RefreshTokenExpiresIn float64 `json:"refreshTokenExpiresIn"`
	TokenType             string  `json:"tokenType"`
}

type MeResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Roles    string `json:"role"`
}

type RefreshTokenRequest struct {
	RefreshToken *string `json:"refreshToken" binding:"required"`
}

type KeycloakTokenResponse struct {
	AccessToken      string  `json:"access_token"`
	RefreshToken     string  `json:"refresh_token"`
	ExpiresIn        float64 `json:"expires_in"`
	RefreshExpiresIn float64 `json:"refresh_expires_in"`
	TokenType        string  `json:"token_type"`
}
