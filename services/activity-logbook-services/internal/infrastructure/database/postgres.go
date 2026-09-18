package database

import (
	"activity-logbook-services/internal/config"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() error {

	var ErrorConnect error

	config := config.NewPostgresConfig()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s client_encoding=UTF8",
		config.Host, config.User, config.Passwrod, config.Name, config.Port, config.SSLMode, config.Timezone)

	ConnectDatabase, ErrorConnect := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})

	if ErrorConnect != nil {
		log.Fatalf("An Error During Connect To Database! : %s", ErrorConnect)
		return ErrorConnect
	}

	ConnectSqlDB, err := ConnectDatabase.DB()

	if err != nil {
		log.Fatalf("failed to get master sql.DB!: %s", err)
		return err
	}

	ConnectSqlDB.SetMaxOpenConns(100)
	ConnectSqlDB.SetMaxIdleConns(25)
	ConnectSqlDB.SetConnMaxLifetime(30 * time.Minute)
	ConnectSqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := ConnectSqlDB.Ping(); err != nil {
		log.Fatalf("failed to ping master sql.DB!: %s", err)
		return err
	}

	log.Println("Success Connect To Postgres ✅")

	DB = ConnectDatabase

	return nil

}
