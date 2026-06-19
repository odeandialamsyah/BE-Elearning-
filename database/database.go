package database

import (
	"backend-elearning/config"
	"backend-elearning/models"
	"crypto/tls"
	"fmt"
	"log"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB(cfg *config.Config) {

	err := mysqlDriver.RegisterTLSConfig("tidb", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Fatal("❌ Failed to register TLS config:", err)
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?tls=tidb&charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

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