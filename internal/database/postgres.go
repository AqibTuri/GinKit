// Package database opens a single GORM connection to PostgreSQL for the whole process.
// Repositories receive *gorm.DB; they do not open their own connections.
package database

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open connects with optional SQL logging in development (debug=true).
func Open(dsn string, debug bool) (*gorm.DB, error) {
	cfg := &gorm.Config{
		PrepareStmt: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}
	if debug {
		cfg.Logger = logger.Default.LogMode(logger.Info)
	}
	return gorm.Open(postgres.Open(dsn), cfg)
}
