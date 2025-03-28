package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/warehouse-service/internal/models"
	"github.com/yourusername/warehouse-service/internal/repository"
)

type WarehouseHandler struct {
	Repo repository.WarehouseRepository
}

// NewWarehouseHandler создает новый обработчик для работы со складами
func NewWarehouseHandler(repo repository.WarehouseRepository) *WarehouseHandler {
	return &WarehouseHandler{Repo: repo}
}

// CreateHandler обрабатывает запросы на создание склада
func (h *WarehouseHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	var request models.Warehouse

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	warehouseID, err := h.Repo.CreateWarehouse(r.Context(), request.Name, request.Address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createdWarehouse, err := h.Repo.GetWarehouseByID(r.Context(), warehouseID)
	if err != nil {
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
	json.NewEncoder(w).Encode(response)
}

// GetAllHandler обрабатывает запросы на получение всех складов
func (h *WarehouseHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	warehouses, err := h.Repo.GetAllWarehouses(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(warehouses)
}

// UpdateHandler обрабатывает запросы на обновление склада
func (h *WarehouseHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	var request struct {
		Location string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.Repo.UpdateWarehouse(r.Context(), id, request.Location)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedWarehouse, err := h.Repo.GetWarehouseByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to fetch updated warehouse", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedWarehouse)
}

// DeleteHandler обрабатывает запросы на удаление склада
func (h *WarehouseHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Println("Received ID:", vars["id"])
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	if err := h.Repo.DeleteWarehouse(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
