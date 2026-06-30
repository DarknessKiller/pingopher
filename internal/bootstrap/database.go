package bootstrap

import (
	"fmt"
	"log"

	"github.com/DarknessKiller/pingopher/internal/config"
	"github.com/DarknessKiller/pingopher/internal/model"
	"github.com/kofj/gorm-driver-d1/gormd1"
	"github.com/libtnb/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitiateDatabase(cfg *config.Config) *gorm.DB {
	var db *gorm.DB
	switch cfg.DatabaseType {
	case "sqlite":
		db = InitiateSQLiteDatabase(cfg)
	case "d1":
		db = InitiateD1Database(cfg)
	default:
		log.Fatal("Unsupported database type:", cfg.DatabaseType)
		return nil
	}

	if cfg.Env != "production" {
		db = db.Debug()
	}

	if err := db.AutoMigrate(
		&model.Host{},
		&model.History{},
		&model.Notification{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	return db
}

// SQLite
func InitiateSQLiteDatabase(cfg *config.Config) *gorm.DB {

	db, err := gorm.Open(sqlite.Open(cfg.SQLitePath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to SQLite:", err)
	}

	log.Println("SQLite connected! File:", cfg.SQLitePath)
	return db
}

// Cloudflare D1
func InitiateD1Database(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("d1://%s:%s@%s",
		cfg.CloudflareD1.AccountID,
		cfg.CloudflareD1.AuthToken,
		cfg.CloudflareD1.DatabaseString)

	gormDB, err := gorm.Open(gormd1.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatal("Failed to initialize GORM with D1:", err)
	}

	log.Println("D1 connected! Database:", cfg.CloudflareD1.DatabaseString)
	return gormDB
}
