package helper

import "github.com/golang-jwt/jwt/v5"

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

	if userRole["ADMIN"] {
		return "Admin"
	} else if userRole["USER"] {
		return "User Public"
	} else if userRole["Super Admin"] {
		return "Super Admin"
	}

	return "USER"
}
