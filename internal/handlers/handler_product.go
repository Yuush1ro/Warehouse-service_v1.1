package handlers

import (
	"encoding/json"
	"net/http"

	//"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/warehouse-service/internal/models"
	"github.com/yourusername/warehouse-service/internal/repository"
	"go.uber.org/zap"
)

type ProductHandler struct {
	Repo   repository.ProductRepository
	Logger *zap.Logger
}

func NewProductHandler(repo repository.ProductRepository, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{Repo: repo, Logger: logger}
}

func (h *ProductHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Received POST request to /api/product")
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		h.Logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if product.Name == "" || product.Barcode == "" {
		h.Logger.Error("Missing required fields")
		http.Error(w, "Name and Barcode are required", http.StatusBadRequest)
		return
	}

	product.ID = uuid.New().String()

	err := h.Repo.Create(r.Context(), product)
	if err != nil {
		h.Logger.Error("Failed to create product", zap.Error(err))
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "created", "id": product.ID}); err != nil {
		h.Logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	products, err := h.Repo.GetAll(r.Context())
	if err != nil {
		h.Logger.Error("Invalid warehouse ID", zap.Error(err))
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	if err := json.NewEncoder(w).Encode(products); err != nil {
		h.Logger.Error("Failed to encode products response", zap.Error(err))
		http.Error(w, "Failed to encode products response", http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.Logger.Error("Missing product ID")
		http.Error(w, "Missing product ID", http.StatusBadRequest)
		return
	}

	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		h.Logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	product.ID = id // Присваиваем ID из пути

	if err := h.Repo.Update(r.Context(), product); err != nil {
		h.Logger.Error("Failed to update product", zap.Error(err))
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProductHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Missing product ID", http.StatusBadRequest)
		return
	}
	h.Logger.Info("Deleting product", zap.String("id", id))

	err := h.Repo.Delete(r.Context(), id)
	if err != nil {
		h.Logger.Error("Failed to delete product", zap.Error(err))
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
