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

type AppConfig struct {
	PortConfig  *PortConfig
	MySQLConfig *MySQLConfig
	RedisConfig *RedisConfig
}

func NewAppConfig() *AppConfig {

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Println("Error loading .env file")
		}
	}

	return &AppConfig{
		PortConfig:  NewPortConfig(),
		MySQLConfig: NewMySQLConfig(),
		RedisConfig: NewRedisConfig(),
	}
}
