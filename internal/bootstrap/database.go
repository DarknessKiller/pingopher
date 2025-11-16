package bootstrap

import (
	"log"

	"github.com/DarknessKiller/pingopher/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitiateDatabase(cfg *config.Config) *gorm.DB {

	db, err := gorm.Open(sqlite.Open(cfg.SQLitePath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to SQLite:", err)
	}

	log.Println("SQLite connected! File →", cfg.SQLitePath)
	return db
}
