package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/warehouse-service/internal/models"
	"github.com/yourusername/warehouse-service/internal/repository"
	"go.uber.org/zap"
)

type WarehouseHandler struct {
	Repo   repository.WarehouseRepository
	logger *zap.Logger
}

func NewWarehouseHandler(repo repository.WarehouseRepository) *WarehouseHandler {
	logger, _ := zap.NewProduction() // Initialize a new logger
	return &WarehouseHandler{
		Repo:   repo,
		logger: logger,
	}
}

// CreateHandler обрабатывает запросы на создание склада
func (h *WarehouseHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	var request models.Warehouse

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	warehouseID, err := h.Repo.CreateWarehouse(r.Context(), request.Name, request.Address)
	if err != nil {
		h.logger.Error("Failed to create warehouse", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createdWarehouse, err := h.Repo.GetWarehouseByID(r.Context(), warehouseID)
	if err != nil {
		h.logger.Error("Failed to fetch created warehouse", zap.Error(err))
		http.Error(w, "Failed to fetch created warehouse", http.StatusInternalServerError)
		return
	}

	// Возвращаем только имя и адрес
	response := struct {
		ID      uuid.UUID `json:"id"`
		Name    string    `json:"name"`
		Address string    `json:"address"`
	}{
		ID:      createdWarehouse.ID,
		Name:    createdWarehouse.Name,
		Address: createdWarehouse.Address,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetAllHandler обрабатывает запросы на получение всех складов
func (h *WarehouseHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	warehouses, err := h.Repo.GetAllWarehouses(r.Context())
	if err != nil {
		h.logger.Error("Failed to fetch warehouses", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(warehouses); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateHandler обрабатывает запросы на обновление склада
func (h *WarehouseHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.logger.Error("Invalid warehouse ID", zap.Error(err))
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Location string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.Repo.UpdateWarehouse(r.Context(), id, request.Location)
	if err != nil {
		h.logger.Error("Failed to update warehouse", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedWarehouse, err := h.Repo.GetWarehouseByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch updated warehouse", zap.Error(err))
		http.Error(w, "Failed to fetch updated warehouse", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedWarehouse); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteHandler обрабатывает запросы на удаление склада
func (h *WarehouseHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Println("Received ID:", vars["id"])
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		h.logger.Error("Invalid warehouse ID", zap.Error(err))
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	if err := h.Repo.DeleteWarehouse(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete warehouse", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
