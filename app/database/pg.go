// Package database provides PostgreSQL database connection utilities.
package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// New creates a new PostgreSQL database connection and returns a cleanup function.
// Returns an error if the connection fails, allowing the caller to handle it appropriately.
func New(user, password, dbname, port string) (db *gorm.DB, close func() error, err error) {
	dsn := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", user, password, port, dbname)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	return db, sqlDB.Close, nil
}
