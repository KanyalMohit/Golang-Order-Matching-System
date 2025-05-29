package main

import (
	"database/sql"
	"log"
	"orderSystem/internal/api"
	"orderSystem/internal/config"
	"orderSystem/internal/migration"
	"orderSystem/internal/repository"
	"orderSystem/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load(logger)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database connection
	db, err := sql.Open("mysql", cfg.DatabaseDSN)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Run database migrations
	if err := migration.RunMigrations(db); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Initialize repository and service
	repo := repository.NewMySQLRepository(db)
	matchingService := service.NewMatchingService(repo, logger)

	// Initialize router
	router := gin.Default()
	handler := api.NewHandler(matchingService, logger)
	api.SetupRoutes(router, handler)

	// Start server
	logger.Info("Starting server", zap.String("address", cfg.ServerAddr))
	if err := router.Run(cfg.ServerAddr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
