package database

import (
	"backend-elearning/config"
	"backend-elearning/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB(cfg *config.Config) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect DB: ", err)
	}

	// migrate
	err = DB.AutoMigrate(&models.Course{}, &models.Module{})
	if err != nil {
		log.Fatal("❌ Failed to migrate: ", err)
	}

	// migrate
	// err = DB.AutoMigrate(&models.User{}, &models.Course{}, &models.Module{})
	// if err != nil {
	// 	log.Fatal("❌ Failed to migrate: ", err)
	// }

	log.Println("✅ Database connected & migrated")
}
