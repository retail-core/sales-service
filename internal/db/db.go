package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/retail-core/sales-service/internal/config"
	"github.com/retail-core/sales-service/internal/logger"
	"github.com/retail-core/sales-service/internal/models"
)

func InitDB(config config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.DB_SOURCE), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	logger.L().Info("Running GORM auto-migrations...")
	err = db.AutoMigrate(
		&models.Order{}, 
		&models.OrderItem{},
	)

	// TODO: ping the database connection to ensure it's alive
	
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.L().Info("Database connection established and migrations completed.")
	return db, nil
}