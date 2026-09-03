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
	GRPC string
}

func NewPortConfig() *PortConfig {
	return &PortConfig{
		PORT: os.Getenv("BORROWING_MANAGEMENT_PORT"),
		GRPC: os.Getenv("USER_MANAGEMENT_GRPC_PORT"),
	}
}

type MySQLConfig struct {
	Host      string
	User      string
	Port      string
	Charset   string
	Collation string
	Database  string
	Password  string
}

func NewMySQLConfig() *MySQLConfig {
	return &MySQLConfig{
		Host:      os.Getenv("borrowingManagement_MYSQL_HOST"),
		User:      os.Getenv("borrowingManagement_MYSQL_USER"),
		Port:      os.Getenv("MYSQL_PORT_GLOBAL"),
		Charset:   os.Getenv("borrowingManagement_MYSQL_CHARSET"),
		Collation: os.Getenv("borrowingManagement_MYSQL_COLLATION"),
		Database:  os.Getenv("borrowingManagement_MYSQL_DATABASE"),
		Password:  os.Getenv("borrowingManagement_MYSQL_PASSWORD"),
	}
}

type RedisCacheConfig struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisCacheTTL time.Duration
}

func NewRedisCacheConfig() *RedisCacheConfig {
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	detail, errDetail := time.ParseDuration(os.Getenv("REDIS_CACHE_TTL"))

	if errDetail != nil {
		detail = 11 * time.Minute
	}

	return &RedisCacheConfig{
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
	Brokers                []string
	GroupID                string
	TopicUserAuthenticated string
	TopicUserStatusUpdated string
}

func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Brokers:                []string{os.Getenv("KAFKA_BROKERS")},
		GroupID:                os.Getenv("BORROWING_MANAGEMENT_GROUP_ID"),
		TopicUserAuthenticated: os.Getenv("KAFKA_TOPIC_USER_LOGIN"),
		TopicUserStatusUpdated: os.Getenv("KAFKA_TOPIC_USER_UPDATED"),
	}
}

type AppConfig struct {
	PortConfig       *PortConfig
	MySQLConfig      *MySQLConfig
	RedisCacheConfig *RedisCacheConfig
	Keycloak         *KeycloakConfig
	Kafka            *KafkaConfig
}

func NewAppConfig() *AppConfig {

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Println("Error loading .env file")
		}
	}

	return &AppConfig{
		PortConfig:       NewPortConfig(),
		MySQLConfig:      NewMySQLConfig(),
		RedisCacheConfig: NewRedisCacheConfig(),
		Keycloak:         NewKeycloakConfig(),
		Kafka:            NewKafkaConfig(),
	}
}
