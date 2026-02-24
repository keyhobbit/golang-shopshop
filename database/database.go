package database

import (
	"log"
	"os"
	"path/filepath"

	"shoop-golang/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(dbPath string) *gorm.DB {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("failed to create db directory: %v", err)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := DB.AutoMigrate(
		&models.AdminUser{},
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Image{},
		&models.Order{},
		&models.OrderItem{},
		&models.Banner{},
		&models.CompanyInfo{},
		&models.AboutPage{},
		&models.SEOBanner{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	log.Println("Database initialized and migrated successfully")
	return DB
}
