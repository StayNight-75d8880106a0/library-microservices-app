package middleware

import (
	"auth-services/internal/config"
	"auth-services/internal/helper"
	"auth-services/internal/repository"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwks keyfunc.Keyfunc, repo repository.AuthRepositoryInterface, cfg *config.AppConfig) gin.HandlerFunc {

	expectedAudience := cfg.Keycloak.ClientID

	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Authorization header required!", helper.ErrorDetail{Detail: "Authorization Header Must Be Bearer!"}))
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if repo.IsTokenBlacklisted(ctx, tokenString) {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Token is revoked/logged out!", helper.ErrorDetail{Detail: "Token is blacklisted!"}))
			ctx.Abort()
			return
		}

		token, errToken := jwt.Parse(tokenString, jwks.Keyfunc,
			jwt.WithValidMethods([]string{"RS512"}),
			jwt.WithIssuer(cfg.Keycloak.KeycloakIssuer),
			jwt.WithExpirationRequired(),
		)

		if errToken != nil {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Invalid Token!", helper.ErrorDetail{Detail: errToken.Error()}))
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Invalid Token Claims!", helper.ErrorDetail{Detail: "Failed to parse claims structure"}))
			ctx.Abort()
			return
		}

		if azp, exists := claims["azp"].(string); exists {
			if azp != expectedAudience {
				helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Invalid Audience!", helper.ErrorDetail{Detail: "Token is not intended for this client!"}))
				ctx.Abort()
				return
			}
		}

		if userID, ok := claims["sub"].(string); ok {
			ctx.Set("user_id", userID)
		}
		if username, ok := claims["preferred_username"].(string); ok {
			ctx.Set("username", username)
		}
		if email, ok := claims["email"].(string); ok {
			ctx.Set("email", email)
		}
		if name, oke := claims["name"].(string); oke {
			ctx.Set("name", name)
		}
		ctx.Set("claims", claims)

		ctx.Next()

	}
}
