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
		PORT: os.Getenv("BOOK_MANAGEMENT_PORT"),
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
		Host:      os.Getenv("bookManagement_MYSQL_HOST"),
		User:      os.Getenv("bookManagement_MYSQL_USER"),
		Port:      os.Getenv("MYSQL_PORT_GLOBAL"),
		Charset:   os.Getenv("bookManagement_MYSQL_CHARSET"),
		Collation: os.Getenv("bookManagement_MYSQL_COLLATION"),
		Database:  os.Getenv("bookManagement_MYSQL_DATABASE"),
		Password:  os.Getenv("bookManagement_MYSQL_PASSWORD"),
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

type ElasticsearchConfig struct {
	ElasticsearchURL string
}

func NewElasticsearchConfig() *ElasticsearchConfig {
	return &ElasticsearchConfig{
		ElasticsearchURL: os.Getenv("ELASTICSEARCH_URL"),
	}
}

type OpenLibraryAPIConfig struct {
	OpenLibraryAPIURL string
}

func NewOpenLibraryAPIConfig() *OpenLibraryAPIConfig {
	return &OpenLibraryAPIConfig{
		OpenLibraryAPIURL: os.Getenv("OPEN_LIBRARY_API_URL"),
	}
}

type AppConfig struct {
	PortConfig     *PortConfig
	MySQLConfig    *MySQLConfig
	RedisConfig    *RedisConfig
	Keycloak       *KeycloakConfig
	Elasticsearch  *ElasticsearchConfig
	OpenLibraryAPI *OpenLibraryAPIConfig
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
		MySQLConfig:    NewMySQLConfig(),
		RedisConfig:    NewRedisConfig(),
		Keycloak:       NewKeycloakConfig(),
		Elasticsearch:  NewElasticsearchConfig(),
		OpenLibraryAPI: NewOpenLibraryAPIConfig(),
	}
}
