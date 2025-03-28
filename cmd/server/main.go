package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/warehouse-service/internal/handlers"
	"github.com/yourusername/warehouse-service/internal/repository"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	//database connection
	dbpool, err := pgxpool.New(context.Background(), "postgres://user:password@db:5432/warehouse?sslmode=disable")
	if err != nil {
		logger.Fatal("Unable to connect to database", zap.Error(err))
	}
	defer dbpool.Close()

	//repository
	warehouseRepo := repository.NewWarehouseRepository(dbpool)
	productRepo := repository.NewProductRepository(dbpool)
	inventoryRepo := repository.NewInventoryRepository(dbpool)
	analyticsRepo := repository.NewAnalyticsRepository(dbpool, logger)

	//handlers
	warehouseHandler := handlers.NewWarehouseHandler(warehouseRepo)
	productHandler := handlers.NewProductHandler(productRepo, logger)
	inventoryHandler := handlers.NewInventoryHandler(inventoryRepo, analyticsRepo, logger)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsRepo, logger)

	router := mux.NewRouter()

	// Health-check
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200"))
	}).Methods("GET")

	//warehouse routes
	router.HandleFunc("/api/warehouses", warehouseHandler.GetAllHandler).Methods("GET")
	router.HandleFunc("/api/warehouse", warehouseHandler.CreateHandler).Methods("POST")
	router.HandleFunc("/api/warehouse/update/{id}", warehouseHandler.UpdateHandler).Methods("PUT")
	router.HandleFunc("/api/warehouse/delete/{id}", warehouseHandler.DeleteHandler).Methods("DELETE")

	//product routes
	router.HandleFunc("/api/products", productHandler.GetAllHandler).Methods("GET")
	router.HandleFunc("/api/product", productHandler.CreateHandler).Methods("POST")
	router.HandleFunc("/api/product/update/{id}", productHandler.UpdateHandler).Methods("PUT")
	router.HandleFunc("/api/product/delete/{id}", productHandler.DeleteHandler).Methods("DELETE")

	//inventory routes
	router.HandleFunc("/api/inventory", inventoryHandler.CreateHandler).Methods("POST")
	router.HandleFunc("/api/inventory/update/{warehouseId}/{productId}", inventoryHandler.UpdateQuantityHandler).Methods("PUT")
	router.HandleFunc("/api/inventory/discount/{warehouseId}", inventoryHandler.SetDiscountHandler).Methods("PUT")
	router.HandleFunc("/api/inventory/{warehouseId}", inventoryHandler.GetByWarehouseHandler).Methods("GET")
	router.HandleFunc("/api/inventory/{warehouseId}/{productId}", inventoryHandler.GetProductHandler).Methods("GET")
	router.HandleFunc("/api/inventory/calculate/{warehouseId}", inventoryHandler.CalculateTotalHandler).Methods("POST")
	router.HandleFunc("/api/inventory/purchase/{warehouseId}", inventoryHandler.PurchaseHandler).Methods("POST")
	router.HandleFunc("/api/inventory/{warehouseId}/{productId}", inventoryHandler.DeleteProductFromWarehouseHandler).Methods("GET")
	router.HandleFunc("/api/inventory/{inventoryID}", inventoryHandler.DeleteInventoryHandler).Methods("DELETE")

	//analytics routes
	router.HandleFunc("/api/analytics/top", analyticsHandler.GetTopWarehousesHandler).Methods("GET")
	router.HandleFunc("/api/analytics/{warehouseId}", analyticsHandler.GetWarehouseAnalyticsHandler).Methods("GET")
	router.HandleFunc("/api/analytics/delete/{warehouseId}/{productId}", analyticsHandler.DeleteAnalyticsHandler).Methods("DELETE")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	//server start
	go func() {
		logger.Info("Starting server on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
