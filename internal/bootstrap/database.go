package bootstrap

import (
	"fmt"
	"log"
	"strings"

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

func InitiateSQLiteDatabase(cfg *config.Config) *gorm.DB {

	db, err := gorm.Open(sqlite.Open(sqliteDSN(cfg.SQLitePath)), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to SQLite:", err)
	}

	log.Println("SQLite connected! File:", cfg.SQLitePath)
	return db
}

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

// sqliteDSN appends the durability pragmas used for concurrent probe writes: WAL lets readers work
// during a write, NORMAL sync keeps commits cheap. busy_timeout(5000) is added by the driver itself.
func sqliteDSN(path string) string {
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + "_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
}
