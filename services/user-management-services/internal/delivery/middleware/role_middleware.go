package middleware

import (
	"user-management-services/internal/helper"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedRoles))

	for _, role := range allowedRoles {
		allowed[role] = true
	}

	return func(ctx *gin.Context) {

		value, exists := ctx.Get("claims")

		if !exists {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Unauthorized!", helper.ErrorDetail{Detail: "Claims not found, Authenticate middleware is missing!"}))
			ctx.Abort()
			return
		}

		claims, ok := value.(jwt.MapClaims)

		if !ok {
			helper.NewErrorResponse(ctx, helper.NewUnauthorizedError("Unauthorized!", helper.ErrorDetail{Detail: "Invalid claims format!"}))
			ctx.Abort()
			return
		}

		realmAccess, _ := claims["realm_access"].(map[string]interface{})
		rolesRaw, _ := realmAccess["roles"].([]interface{})

		for _, raw := range rolesRaw {
			if role, ok := raw.(string); ok && allowed[role] {
				ctx.Next()
				return
			}
		}

		helper.NewErrorResponse(ctx, helper.NewForbiddenError("Access Denied!", helper.ErrorDetail{Detail: "You do not have the required role to access this resource!"}))
		ctx.Abort()
	}
}
