package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/warehouse-service/internal/repository"
	"go.uber.org/zap"
)

type AnalyticsHandler struct {
	Repo   repository.AnalyticsRepository
	Logger *zap.Logger
}

func NewAnalyticsHandler(repo repository.AnalyticsRepository, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{Repo: repo, Logger: logger}
}

// 1. Получение аналитики по складу
func (h *AnalyticsHandler) GetWarehouseAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.Logger.Info("Fetching warehouse analytics", zap.String("warehouseId", vars["warehouseId"]))
	warehouseID, err := uuid.Parse(vars["warehouseId"])
	if err != nil {
		h.Logger.Error("Invalid UUID format", zap.String("warehouseId", vars["warehouseId"]))
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	analytics, err := h.Repo.GetWarehouseAnalytics(r.Context(), warehouseID)
	if err != nil {
		http.Error(w, "Failed to fetch analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// 2. Получение топ-10 складов по выручке
func (h *AnalyticsHandler) GetTopWarehousesHandler(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10 // По умолчанию возвращаем топ-10
	}

	h.Logger.Info("Fetching top warehouses by revenue", zap.Int("limit", limit))

	warehouses, err := h.Repo.GetTopWarehouses(r.Context(), limit)
	if err != nil {
		h.Logger.Error("Failed to fetch top warehouses", zap.Error(err))
		http.Error(w, "Failed to fetch top warehouses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(warehouses)
}

func (h *AnalyticsHandler) DeleteAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.Logger.Info("Deleting analytics data", zap.String("warehouseId", vars["warehouseId"]), zap.String("productId", vars["productId"]))

	warehouseID, err := uuid.Parse(vars["warehouseId"])
	if err != nil {
		h.Logger.Error("Invalid UUID format for warehouseId", zap.String("warehouseId", vars["warehouseId"]))
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	productID, err := uuid.Parse(vars["productId"])
	if err != nil {
		h.Logger.Error("Invalid UUID format for productId", zap.String("productId", vars["productId"]))
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if err := h.Repo.DeleteAnalytics(r.Context(), warehouseID, productID); err != nil {
		h.Logger.Error("Failed to delete analytics data", zap.Error(err))
		http.Error(w, "Failed to delete analytics data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
