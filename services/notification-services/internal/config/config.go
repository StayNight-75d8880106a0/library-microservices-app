package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type PortConfig struct {
	PORT string
}

func NewPortConfig() *PortConfig {
	return &PortConfig{
		PORT: os.Getenv("NOTIFICATION_SERVICE_PORT"),
	}
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Passwrod string
	Name     string
	SSLMode  string
	Timezone string
}

func NewPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:     os.Getenv("notificationService_POSTGRES_HOST"),
		Port:     os.Getenv("DB_PORT_GLOBAL"),
		User:     os.Getenv("notificationService_POSTGRES_USER"),
		Passwrod: os.Getenv("notificationService_POSTGRES_PASSWORD"),
		Name:     os.Getenv("notificationService_POSTGRES_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		Timezone: os.Getenv("DB_TIMEZONE"),
	}
}

type RedisConfig struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisCacheTTL time.Duration
}

func NewRedisConfig() *RedisConfig {
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	detail, errDetail := time.ParseDuration(os.Getenv("REDIS_CACHE_TTL"))

	if errDetail != nil {
		detail = 11 * time.Minute
	}

	return &RedisConfig{
		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisCacheTTL: detail,
		RedisDB:       redisDB,
	}
}

type KeycloakConfig struct {
	KeycloakURL    string
	KeycloakRealm  string
	KeycloakIssuer string
	ClientID       string
	Realm          string
}

func NewKeycloakConfig() *KeycloakConfig {
	return &KeycloakConfig{
		KeycloakURL:    os.Getenv("KEYCLOAK_URL"),
		KeycloakRealm:  os.Getenv("KEYCLOAK_REALM"),
		KeycloakIssuer: os.Getenv("KEYCLOAK_ISSUER"),
		ClientID:       os.Getenv("KEYCLOAK_CLIENT_ID"),
		Realm:          os.Getenv("KEYCLOAK_REALM"),
	}
}

type KafkaConfig struct {
	Brokers          []string
	GroupID          string
	TopicUserCreated string
}

func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers:          []string{os.Getenv("KAFKA_BROKERS")},
		GroupID:          os.Getenv("KAFKA_CONSUMER_GROUP_BOOK_MANAGEMENT"),
		TopicUserCreated: os.Getenv("KAFKA_TOPIC_USER_CREATED"),
	}
}

type AppConfig struct {
	PortConfig     *PortConfig
	PostgresConfig *PostgresConfig
	RedisConfig    *RedisConfig
	Keycloak       *KeycloakConfig
	KafkaConfig    *KafkaConfig
}

func NewAppConfig() *AppConfig {

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Println("Error loading .env file")
		}
	}

	return &AppConfig{
		PortConfig:     NewPortConfig(),
		PostgresConfig: NewPostgresConfig(),
		RedisConfig:    NewRedisConfig(),
		Keycloak:       NewKeycloakConfig(),
		KafkaConfig:    NewKafkaConfig(),
	}
}
