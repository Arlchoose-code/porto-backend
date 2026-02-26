package database

import (
	"arlchoose/backend-api/config"
	"arlchoose/backend-api/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {

	// Load konfigurasi database dari .env
	dbUser := config.GetEnv("DB_USER", "root")
	dbPass := config.GetEnv("DB_PASS", "")
	dbHost := config.GetEnv("DB_HOST", "localhost")
	dbPort := config.GetEnv("DB_PORT", "3306")
	dbName := config.GetEnv("DB_NAME", "")

	// Format DSN untuk MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// Koneksi ke database
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Database connected successfully!")

	// **Auto Migrate Models**
	err = DB.AutoMigrate(
		&models.User{},
		&models.Profile{},
		&models.Setting{},
		&models.Contact{},
		&models.Skill{},
		&models.Education{},
		&models.Course{},
		&models.Experience{},
		&models.ExperienceImage{},
		&models.Project{},
		&models.ProjectTechStack{},
		&models.ProjectImage{},
		&models.Tag{},
		&models.Blog{},
		&models.Bookmark{},
		&models.BookmarkTopic{},
		&models.Tool{},
		&models.ToolUsage{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("Database migrated successfully!")
}
