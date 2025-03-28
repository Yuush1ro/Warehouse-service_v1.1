package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/warehouse-service/internal/models"
	"github.com/yourusername/warehouse-service/internal/repository"
	"go.uber.org/zap"
)

type InventoryHandler struct {
	Repo          repository.InventoryRepository
	AnalyticsRepo repository.AnalyticsRepository
	Logger        *zap.Logger
}

func NewInventoryHandler(repo repository.InventoryRepository, analyticsRepo repository.AnalyticsRepository, logger *zap.Logger) *InventoryHandler {
	return &InventoryHandler{
		Repo:          repo,
		AnalyticsRepo: analyticsRepo,
		Logger:        logger,
	}
}

// 1. Создание связи товара и склада (указание цены)
func (h *InventoryHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	var inventory models.Inventory
	if err := json.NewDecoder(r.Body).Decode(&inventory); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.Logger.Debug("Received inventory data", zap.Any("inventory", inventory))

	inventory.ID = uuid.New()

	err := h.Repo.Create(r.Context(), inventory)
	if err != nil {
		h.Logger.Error("Failed to create inventory record", zap.Error(err))
		http.Error(w, "Failed to create inventory record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

// 2. Обновление количества товара (поступление на склад)
func (h *InventoryHandler) UpdateQuantityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["productId"])
	warehouseID, err2 := uuid.Parse(vars["warehouseId"])
	if err != nil || err2 != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	var request struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.Repo.UpdateQuantity(r.Context(), productID, warehouseID, request.Quantity)
	if err != nil {
		http.Error(w, "Failed to update quantity", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// 3. Установка скидки
func (h *InventoryHandler) SetDiscountHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ProductIDs []uuid.UUID `json:"product_ids"`
		Discount   float64     `json:"discount"`
	}
	vars := mux.Vars(r)
	warehouseID, err := uuid.Parse(vars["warehouseId"])
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.Repo.SetDiscount(r.Context(), request.ProductIDs, warehouseID, request.Discount)
	if err != nil {
		http.Error(w, "Failed to set discount", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "discount applied"})
}

// 4. Получение списка товаров на складе (с пагинацией)
func (h *InventoryHandler) GetByWarehouseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.Logger.Info("Received warehouse ID", zap.String("warehouseId", vars["warehouseId"]))

	warehouseID, err := uuid.Parse(vars["warehouseId"])
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	inventory, err := h.Repo.GetByWarehouse(r.Context(), warehouseID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get inventory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventory)
}

// 5. Получение информации о товаре на складе
func (h *InventoryHandler) GetProductHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := uuid.Parse(vars["productId"])
	warehouseID, err2 := uuid.Parse(vars["warehouseId"])
	if err != nil || err2 != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	inventory, err := h.Repo.GetProductInWarehouse(r.Context(), productID, warehouseID)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventory)
}

// 6. Подсчёт стоимости корзины
func (h *InventoryHandler) CalculateTotalHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Items map[uuid.UUID]int `json:"items"`
	}
	vars := mux.Vars(r)
	warehouseID, err := uuid.Parse(vars["warehouseId"])
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	total, err := h.Repo.CalculateTotal(r.Context(), warehouseID, request.Items)
	if err != nil {
		http.Error(w, "Failed to calculate total", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"total": total})
}

// 7. Покупка товаров
func (h *InventoryHandler) PurchaseHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Items map[uuid.UUID]int `json:"items"`
	}
	vars := mux.Vars(r)
	warehouseID, err := uuid.Parse(vars["warehouseId"])
	if err != nil {
		http.Error(w, "Invalid warehouse ID", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.Repo.Purchase(r.Context(), warehouseID, request.Items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for productID, quantity := range request.Items {
		price, err := h.Repo.GetProductPrice(r.Context(), warehouseID, productID)
		if err != nil {
			h.Logger.Error("Failed to get product price", zap.Error(err))
			continue
		}
		discount, err := h.Repo.GetProductDiscount(r.Context(), warehouseID, productID)
		if err != nil {
			if err.Error() == "no rows in result set" {
				h.Logger.Warn("No discount found for product, defaulting to 0",
					zap.String("warehouseID", warehouseID.String()),
					zap.String("productID", productID.String()))
				discount = 0 // Устанавливаем скидку в 0, если запись отсутствует
			} else {
				h.Logger.Error("Failed to get product discount", zap.Error(err))
				continue
			}
		}

		// Применяем скидку
		finalPrice := price * (1 - discount/100)
		totalPrice := float64(quantity) * finalPrice

		h.Logger.Info("Recording sale in analytics",
			zap.String("warehouseID", warehouseID.String()),
			zap.String("productID", productID.String()),
			zap.Int("quantity", quantity),
			zap.Float64("totalPrice", totalPrice))

		err = h.AnalyticsRepo.RecordSale(r.Context(), warehouseID, productID, quantity, totalPrice)
		if err != nil {
			h.Logger.Error("Failed to record sale in analytics",
				zap.String("warehouseID", warehouseID.String()),
				zap.String("productID", productID.String()),
				zap.Int("quantity", quantity),
				zap.Float64("totalPrice", totalPrice),
				zap.Error(err)) // Логируем ошибку
			continue
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "purchase successful"})

}

func (h *InventoryHandler) DeleteProductFromWarehouseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	warehouseID, err := uuid.Parse(vars["warehouseId"])
	productID, err2 := uuid.Parse(vars["productId"])
	if err != nil || err2 != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	err = h.Repo.DeleteProductFromWarehouse(r.Context(), warehouseID, productID)
	if err != nil {
		http.Error(w, "Failed to delete product from warehouse", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *InventoryHandler) DeleteInventoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.Logger.Info("Received delete request", zap.Any("vars", vars))

	inventoryID, err := uuid.Parse(vars["inventoryID"])
	if err != nil {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	err = h.Repo.DeleteInventory(r.Context(), inventoryID)
	if err != nil {
		h.Logger.Error("Failed to delete inventory record", zap.Error(err))
		http.Error(w, "Failed to delete inventory record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
