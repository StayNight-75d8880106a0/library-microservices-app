package helper

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func ExtractRoleFromClaims(claims jwt.MapClaims) string {

	realmAccess, ok := claims["realm_access"].(map[string]interface{})

	if !ok {
		return "USER"
	}

	roleRaw, ok := realmAccess["roles"].([]interface{})

	if !ok {
		return "USER"
	}

	userRole := make(map[string]bool)

	for _, value := range roleRaw {
		if roleString, ok := value.(string); ok {
			userRole[roleString] = true
		}
	}

	if userRole["SUPER_ADMIN"] {
		return "SUPER_ADMIN"
	}
	if userRole["ADMIN"] {
		return "ADMIN"
	}
	if userRole["USER"] {
		return "USER"
	}

	return "USER"
}

func CalculateTokenRemainingTTL(tokenString string) time.Duration {

	cleanToken := strings.TrimPrefix(tokenString, "Bearer ")

	token, _, err := new(jwt.Parser).ParseUnverified(cleanToken, jwt.MapClaims{})

	if err != nil {
		return 0
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return 0
	}

	exp, ok := claims["exp"].(float64)

	if !ok {
		return 0
	}

	expirationTime := time.Unix(int64(exp), 0)
	remainingTTL := time.Until(expirationTime)

	if remainingTTL <= 0 {
		return 0
	}

	return remainingTTL

}
