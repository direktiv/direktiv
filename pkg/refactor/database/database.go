package database

import (
	_ "embed"
	"log"
	"os"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed db_schema.sql
var Schema string

func sqlLiteSchema() string {
	convertTypes := map[string]string{
		"uuid":        "text",
		"timestamptz": "datetime",
		"bytea":       "blob",
		"boolean":     "numeric",
		"serial":      "integer",
	}

	liteSchema := Schema

	for k, v := range convertTypes {
		liteSchema = strings.ReplaceAll(liteSchema, " "+k+",", " "+v+",")
		liteSchema = strings.ReplaceAll(liteSchema, " "+k+" ", " "+v+" ")
	}
	liteSchema = strings.ReplaceAll(liteSchema, "CREATE INDEX", "--")

	return liteSchema
}

func NewMockGorm() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,   // Slow SQL threshold
				LogLevel:      logger.Silent, // Log level
			},
		),
	})
	if err != nil {
		return nil, err
	}

	res := db.Exec(sqlLiteSchema())

	if res.Error != nil {
		return nil, res.Error
	}

	return db, nil
}
