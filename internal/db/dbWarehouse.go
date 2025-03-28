package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yourusername/warehouse-service/internal/models"
)

type WarehouseRepository struct {
	DB *pgx.Conn
}

// NewWarehouseRepository создает новый репозиторий для работы с складами
func NewWarehouseRepository(db *pgx.Conn) *WarehouseRepository {
	return &WarehouseRepository{DB: db}
}

// Create создает новый склад в базе данных
func (r *WarehouseRepository) Create(ctx context.Context, warehouse models.Warehouse) (*models.Warehouse, error) {
	warehouse.ID = uuid.New()
	query := `INSERT INTO warehouses (id, address) VALUES ($1, $2) RETURNING id, address`
	err := r.DB.QueryRow(ctx, query, warehouse.ID, warehouse.Address).Scan(&warehouse.ID, &warehouse.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create warehouse: %w", err)
	}
	return &warehouse, nil
}

// GetAll возвращает список всех складов
func (r *WarehouseRepository) GetAll(ctx context.Context) ([]models.Warehouse, error) {
	rows, err := r.DB.Query(ctx, "SELECT id, address FROM warehouses")
	if err != nil {
		return nil, fmt.Errorf("failed to get warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []models.Warehouse
	for rows.Next() {
		var warehouse models.Warehouse
		if err := rows.Scan(&warehouse.ID, &warehouse.Address); err != nil {
			return nil, fmt.Errorf("failed to scan warehouse: %w", err)
		}
		warehouses = append(warehouses, warehouse)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return warehouses, nil
}

// Update обновляет информацию о складе
func (r *WarehouseRepository) Update(ctx context.Context, id uuid.UUID, location string) (*models.Warehouse, error) {
	query := `UPDATE warehouses SET address = $1 WHERE id = $2`
	_, err := r.DB.Exec(ctx, query, location, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update warehouse: %w", err)
	}
	return &models.Warehouse{ID: id, Address: location}, nil
}

// Delete удаляет склад по id
func (r *WarehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM warehouses WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete warehouse: %w", err)
	}
	return nil
}
