package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type PortConfig struct {
	PORT string
}

func NewPortConfig() *PortConfig {
	return &PortConfig{
		PORT: os.Getenv("AUTH_APP_PORT"),
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

type RedisConfig struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
}

func NewRedisConfig() *RedisConfig {
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	return &RedisConfig{
		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       redisDB,
	}
}

type KafkaConfig struct {
	Brokers                []string
	TopicUserCreated       string
	TopicUserAuthenticated string
}

func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers:                []string{os.Getenv("KAFKA_BROKERS")},
		TopicUserCreated:       os.Getenv("KAFKA_TOPIC_USER_CREATED"),
		TopicUserAuthenticated: os.Getenv("KAFKA_TOPIC_USER_LOGIN"),
	}
}

type AppConfig struct {
	Port     *PortConfig
	Keycloak *KeycloakConfig
	Redis    *RedisConfig
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
		Redis:    NewRedisConfig(),
		Kafka:    NewKafkaConfig(),
	}
}
