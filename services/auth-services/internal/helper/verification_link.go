package helper

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateVerificationLink(baseURL string, KeycloakID string, secretKey string, expiry int) (string, error) {

	claims := jwt.MapClaims{
		"sub":  KeycloakID,
		"type": "EMAIL_VERIFICATION",
		"exp":  time.Now().Add(time.Duration(expiry) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	signedToken, errToken := token.SignedString([]byte(secretKey))

	if errToken != nil {
		return "", errToken
	}

	verificationLink := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", baseURL, signedToken)

	return verificationLink, nil

}

func ParseVerificationToken(tokenString string, secretKey string) (jwt.MapClaims, error) {

	token, errParse := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if errParse != nil {
		return nil, errParse
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid verification token")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "EMAIL_VERIFICATION" {
		return nil, errors.New("invalid token type for email verification")
	}

	return claims, nil

}
