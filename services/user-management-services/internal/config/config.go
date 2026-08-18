package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type PortConfig struct {
	PORT string
}

func NewPortConfig() *PortConfig {
	return &PortConfig{
		PORT: os.Getenv("USER_MANAGEMENT_PORT"),
	}
}

type KeycloakConfig struct {
	Realm          string
	ClientID       string
	ClientSecret   string
	KeycloakURL    string
	KeycloakIssuer string
}

func NewKeycloakConfig() *KeycloakConfig {
	return &KeycloakConfig{
		Realm:          os.Getenv("KEYCLOAK_REALM"),
		ClientID:       os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret:   os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		KeycloakURL:    os.Getenv("KEYCLOAK_URL"),
		KeycloakIssuer: os.Getenv("KEYCLOAK_ISSUER"),
	}
}

type AppConfig struct {
	Port     *PortConfig
	Keycloak *KeycloakConfig
}

func NewAppConfig() *AppConfig {

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Println("Error loading .env file")
		}
	}

	return &AppConfig{
		Port:     NewPortConfig(),
		Keycloak: NewKeycloakConfig(),
	}
}
