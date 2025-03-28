package services

import (
	"context"
	"errors"

	"github.com/yourusername/warehouse-service/internal/models"

	//"your_project/models"

	"github.com/google/uuid"
)

// Временное хранилище складов
var warehouses = make(map[uuid.UUID]models.Warehouse)

// AddWarehouse добавляет новый склад
func AddWarehouse(ctx context.Context, warehouse models.Warehouse) (models.Warehouse, error) {
	warehouse.ID = uuid.New()
	warehouses[warehouse.ID] = warehouse
	return warehouse, nil
}

// GetWarehouses возвращает все склады
func GetWarehouses(ctx context.Context) ([]models.Warehouse, error) {
	var result []models.Warehouse
	for _, warehouse := range warehouses {
		result = append(result, warehouse)
	}
	return result, nil
}

// UpdateWarehouse обновляет склад
func UpdateWarehouse(ctx context.Context, id uuid.UUID, warehouse models.Warehouse) (models.Warehouse, error) {
	existingWarehouse, exists := warehouses[id]
	if !exists {
		return models.Warehouse{}, errors.New("warehouse not found")
	}
	// Обновление данных склада
	existingWarehouse.Name = warehouse.Name
	existingWarehouse.Address = warehouse.Address
	//existingWarehouse.Capacity = warehouse.Capacity
	existingWarehouse.Description = warehouse.Description
	warehouses[id] = existingWarehouse
	return existingWarehouse, nil
}

// DeleteWarehouse удаляет склад
func DeleteWarehouse(ctx context.Context, id uuid.UUID) error {
	if _, exists := warehouses[id]; !exists {
		return errors.New("warehouse not found")
	}
	delete(warehouses, id)
	return nil
}
