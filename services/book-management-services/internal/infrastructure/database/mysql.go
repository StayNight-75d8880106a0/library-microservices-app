package mysql

import (
	"book-management-services/internal/config"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectMySQL() error {

	config := config.NewMySQLConfig()

	dsn := fmt.Sprintf(`%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local`,
		config.User, config.Password, config.Host, config.Port, config.Database, config.Charset)

	ConnectDatabase, ErrorConnect := gorm.Open(mysql.Open(dsn), &gorm.Config{
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

	log.Println("Success Connect To MySQL ✅")

	DB = ConnectDatabase

	return nil

}
