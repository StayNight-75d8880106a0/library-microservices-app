package middleware

import (
	"strings"
	"user-management-services/internal/config"
	"user-management-services/internal/helper"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwks keyfunc.Keyfunc, cfg *config.AppConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		authHeader := ctx.GetHeader("Authorization")

		if !strings.HasPrefix(authHeader, "Bearer ") {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Authorization header required!", helper.ErrorDetail{Detail: "Authorization header must be Bearer!"}))
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, errToken := jwt.Parse(tokenString, jwks.Keyfunc,
			jwt.WithValidMethods([]string{"RS512"}),
			jwt.WithIssuer(cfg.Keycloak.KeycloakIssuer),
			jwt.WithExpirationRequired(),
		)

		if errToken != nil {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Invalid token!", helper.ErrorDetail{Detail: errToken.Error()}))
			ctx.Abort()
			return
		}

		if !token.Valid {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Invalid token!", helper.ErrorDetail{Detail: "Token is not valid"}))
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Invalid token!", helper.ErrorDetail{Detail: "Token claims is not valid"}))
			ctx.Abort()
			return
		}

		if azp, _ := claims["azp"].(string); azp != cfg.Keycloak.ClientID {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Invalid Audience!", helper.ErrorDetail{Detail: "Token is not intended for this client!"}))
			ctx.Abort()
			return
		}

		userID, _ := claims["sub"].(string)

		if userID == "" {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Invalid Token!", helper.ErrorDetail{Detail: "Subject claim is missing!"}))
			ctx.Abort()
			return
		}

		ctx.Set("userID", userID)
		ctx.Set("claims", claims)

	}
}
