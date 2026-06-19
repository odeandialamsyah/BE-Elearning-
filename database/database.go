package database

import (
	"backend-elearning/config"
	"backend-elearning/models"
	"crypto/tls"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var DB *gorm.DB

func ConnectDB(cfg *config.Config) {

	mysqlDriver.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
	})

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?tls=tidb&charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect DB: ", err)
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.Course{},
		&models.Module{},
		&models.Quiz{},
		&models.QuizResult{},
		&models.Enrollment{},
		&models.Feedback{},
	)
	if err != nil {
		log.Fatal("❌ Failed to migrate: ", err)
	}

	log.Println("✅ Database connected & migrated")
}