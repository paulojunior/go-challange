// Command server starts the HTTP API server.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/app/categories"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/app/logger"
	"github.com/mytheresa/go-hiring-challenge/app/middleware"
	"github.com/mytheresa/go-hiring-challenge/app/services"
	"github.com/mytheresa/go-hiring-challenge/models"
)

func main() {
	// Load environment variables from .env file.
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// Initialize structured logger.
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	logger.Init(env)
	logger.Info("Starting application", "env", env)

	// Set up signal handling for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize database connection.
	db, close, err := database.New(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := close(); err != nil {
			logger.Error("Failed to close database", "error", err)
		}
	}()
	logger.Info("Database connected successfully")

	// Initialize repositories.
	prodRepo := models.NewProductsRepository(db)
	catRepo := models.NewCategoriesRepository(db)

	// Initialize services.
	catalogService := services.NewCatalogService(prodRepo)
	categoriesService := services.NewCategoriesService(catRepo)

	// Initialize handlers.
	catalogHandler := catalog.NewCatalogHandler(catalogService)
	categoriesHandler := categories.NewCategoriesHandler(categoriesService)

	// Set up routing.
	mux := http.NewServeMux()

	// API v1 routes
	mux.Handle("GET /v1/catalog", api.ErrorHandler(catalogHandler.HandleGet))
	mux.Handle("GET /v1/catalog/{code}", api.ErrorHandler(catalogHandler.HandleGetByCode))
	mux.Handle("GET /v1/categories", api.ErrorHandler(categoriesHandler.HandleGet))
	mux.Handle("POST /v1/categories", api.ErrorHandler(categoriesHandler.HandlePost))

	// Legacy routes (kept for assignment compatibility)
	mux.Handle("GET /catalog", api.ErrorHandler(catalogHandler.HandleGet))
	mux.Handle("GET /catalog/{code}", api.ErrorHandler(catalogHandler.HandleGetByCode))
	mux.Handle("GET /categories", api.ErrorHandler(categoriesHandler.HandleGet))
	mux.Handle("POST /categories", api.ErrorHandler(categoriesHandler.HandlePost))

	logger.Info("Routes registered", "version", "v1", "legacy_routes_enabled", true)

	// Set up the HTTP server with middlewares.
	// Middlewares are applied in reverse order (last = innermost)
	// Final order: RequestID -> Logger -> Recovery -> mux
	var handler http.Handler = mux
	handler = middleware.Recovery(handler)
	handler = middleware.Logger(handler)
	handler = middleware.RequestID(handler)

	srv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%s", os.Getenv("HTTP_PORT")),
		Handler: handler,
	}

	// Start the server.
	go func() {
		logger.Info("Starting HTTP server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down server...")

	// Create a new context with timeout for graceful shutdown.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed", "error", err)
	} else {
		logger.Info("Server stopped gracefully")
	}

	stop()
}
