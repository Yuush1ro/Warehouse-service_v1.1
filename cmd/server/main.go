package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/yourusername/warehouse-service/internal/config"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("Failed to sync logger", zap.Error(err))
		}
	}()

	// Загрузить .env файл
	if err := godotenv.Load(); err != nil {
		logger.Warn(".env file not found, falling back to environment variables")
	}

	//database connection
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		logger.Fatal("DB_URL is not set in environment")
	}
	dbpool, err := pgxpool.New(context.Background(), dbURL)

	// Инициализация зависимостей
	router := config.SetupDependencies(logger, dbpool)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	server := config.NewServer(port, router)

	//server start
	go func() {
		logger.Info("Starting server", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("ListenAndServe failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Shutdown
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server Shutdown Failed", zap.Error(err))
	}

	logger.Info("Server exited properly")
}
