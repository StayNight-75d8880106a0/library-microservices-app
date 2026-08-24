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

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Passwrod string
	Name     string
	SSLMode  string
	Timezone string
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     os.Getenv("userManagement_DB_HOST"),
		Port:     os.Getenv("DB_PORT_GLOBAL"),
		User:     os.Getenv("userManagement_DB_USER"),
		Passwrod: os.Getenv("userManagement_DB_PASSWORD"),
		Name:     os.Getenv("userManagement_DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		Timezone: os.Getenv("DB_TIMEZONE"),
	}
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers: []string{os.Getenv("KAFKA_BROKERS")},
		Topic:   os.Getenv("KAFKA_TOPIC_USER_CREATED"),
		GroupID: os.Getenv("KAFKA_CONSUMER_GROUP_USER_MANAGEMENT"),
	}
}

type AppConfig struct {
	Port     *PortConfig
	Keycloak *KeycloakConfig
	DB       *DatabaseConfig
	Kafka    *KafkaConfig
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
		DB:       NewDatabaseConfig(),
		Kafka:    NewKafkaConfig(),
	}
}
