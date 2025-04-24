package config

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/warehouse-service/internal/handlers"
	"github.com/yourusername/warehouse-service/internal/middleware"
	"github.com/yourusername/warehouse-service/internal/repository"
	"go.uber.org/zap"
)

// NewServer создает и возвращает настроенный HTTP сервер.
func NewServer(port string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// SetupDependencies инициализирует репозитории, обработчики и маршруты.
func SetupDependencies(logger *zap.Logger, dbpool *pgxpool.Pool) *mux.Router {
	// Репозитории
	warehouseRepo := repository.NewWarehouseRepository(dbpool)
	productRepo := repository.NewProductRepository(dbpool)
	inventoryRepo := repository.NewInventoryRepository(dbpool)
	analyticsRepo := repository.NewAnalyticsRepository(dbpool, logger)

	// Обработчики
	warehouseHandler := handlers.NewWarehouseHandler(warehouseRepo)
	productHandler := handlers.NewProductHandler(productRepo, logger)
	inventoryHandler := handlers.NewInventoryHandler(inventoryRepo, analyticsRepo, logger)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsRepo, logger)

	// Настройка маршрутов
	router := SetupRoutes(logger, warehouseHandler, productHandler, inventoryHandler, analyticsHandler)
	router.Use(middleware.LoggingMiddleware(logger))

	return router
}

// SetupRoutes настраивает маршруты для приложения.
func SetupRoutes(
	logger *zap.Logger,
	warehouseHandler *handlers.WarehouseHandler,
	productHandler *handlers.ProductHandler,
	inventoryHandler *handlers.InventoryHandler,
	analyticsHandler *handlers.AnalyticsHandler,
) *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware(logger))

	// Health-check
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("200")); err != nil {
			logger.Error("Failed to write health check response", zap.Error(err))
		}
	}).Methods("GET")

	// Warehouse routes
	router.HandleFunc("/api/warehouses", warehouseHandler.GetAllHandler).Methods("GET")
	router.HandleFunc("/api/warehouse", warehouseHandler.CreateHandler).Methods("POST")
	router.HandleFunc("/api/warehouse/update/{id}", warehouseHandler.UpdateHandler).Methods("PUT")
	router.HandleFunc("/api/warehouse/delete/{id}", warehouseHandler.DeleteHandler).Methods("DELETE")

	// Product routes
	router.HandleFunc("/api/products", productHandler.GetAllHandler).Methods("GET")
	router.HandleFunc("/api/product", productHandler.CreateHandler).Methods("POST")
	router.HandleFunc("/api/product/update/{id}", productHandler.UpdateHandler).Methods("PUT")
	router.HandleFunc("/api/product/delete/{id}", productHandler.DeleteHandler).Methods("DELETE")

	// Inventory routes
	router.HandleFunc("/api/inventory", inventoryHandler.CreateHandler).Methods("POST")
	router.HandleFunc("/api/inventory/update/{warehouseId}/{productId}", inventoryHandler.UpdateQuantityHandler).Methods("PUT")
	router.HandleFunc("/api/inventory/discount/{warehouseId}", inventoryHandler.SetDiscountHandler).Methods("PUT")
	router.HandleFunc("/api/inventory/{warehouseId}", inventoryHandler.GetByWarehouseHandler).Methods("GET")
	router.HandleFunc("/api/inventory/{warehouseId}/{productId}", inventoryHandler.GetProductHandler).Methods("GET")
	router.HandleFunc("/api/inventory/calculate/{warehouseId}", inventoryHandler.CalculateTotalHandler).Methods("POST")
	router.HandleFunc("/api/inventory/purchase/{warehouseId}", inventoryHandler.PurchaseHandler).Methods("POST")
	router.HandleFunc("/api/inventory/{warehouseId}/{productId}", inventoryHandler.DeleteProductFromWarehouseHandler).Methods("GET")
	router.HandleFunc("/api/inventory/{inventoryID}", inventoryHandler.DeleteInventoryHandler).Methods("DELETE")

	// Analytics routes
	router.HandleFunc("/api/analytics/top", analyticsHandler.GetTopWarehousesHandler).Methods("GET")
	router.HandleFunc("/api/analytics/{warehouseId}", analyticsHandler.GetWarehouseAnalyticsHandler).Methods("GET")
	router.HandleFunc("/api/analytics/delete/{warehouseId}/{productId}", analyticsHandler.DeleteAnalyticsHandler).Methods("DELETE")

	return router
}
