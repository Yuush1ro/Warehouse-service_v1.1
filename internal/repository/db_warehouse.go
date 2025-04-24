package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/warehouse-service/internal/models"
)

var ErrWarehouseNotFound = errors.New("warehouse not found")

type WarehouseRepository interface {
	CreateWarehouse(ctx context.Context, name string, address string) (uuid.UUID, error)
	GetAllWarehouses(ctx context.Context) ([]models.Warehouse, error)
	UpdateWarehouse(ctx context.Context, id uuid.UUID, newAddress string) error
	DeleteWarehouse(ctx context.Context, id uuid.UUID) error
	GetWarehouseByID(ctx context.Context, id uuid.UUID) (*models.Warehouse, error)
}

type WarehouseRepositoryImpl struct {
	db *pgxpool.Pool
}

var _ WarehouseRepository = (*WarehouseRepositoryImpl)(nil)

func NewWarehouseRepository(db *pgxpool.Pool) *WarehouseRepositoryImpl {
	return &WarehouseRepositoryImpl{db: db}
}

// Добавить склад
func (r *WarehouseRepositoryImpl) CreateWarehouse(ctx context.Context, name string, address string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := r.db.Exec(ctx, "INSERT INTO warehouses (id, name, address) VALUES ($1, $2, $3)", id, name, address)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}

// Получить все склады
func (r *WarehouseRepositoryImpl) GetAllWarehouses(ctx context.Context) ([]models.Warehouse, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name, address, description FROM warehouses")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warehouses []models.Warehouse
	for rows.Next() {
		var w models.Warehouse
		var description sql.NullString // Используем sql.NullString для обработки NULL
		if err := rows.Scan(&w.ID, &w.Name, &w.Address, &description); err != nil {
			return nil, err
		}
		w.Description = description.String // Преобразуем в строку
		warehouses = append(warehouses, w)
	}
	return warehouses, nil
}

// Обновить склад
func (r *WarehouseRepositoryImpl) UpdateWarehouse(ctx context.Context, id uuid.UUID, newAddress string) error {
	commandTag, err := r.db.Exec(ctx, "UPDATE warehouses SET address = $1 WHERE id = $2", newAddress, id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrWarehouseNotFound
	}

	return nil
}

// Удалить склад
func (r *WarehouseRepositoryImpl) DeleteWarehouse(ctx context.Context, id uuid.UUID) error {
	commandTag, err := r.db.Exec(ctx, "DELETE FROM warehouses WHERE id = $1", id)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrWarehouseNotFound
	}

	return nil
}

func (r *WarehouseRepositoryImpl) GetWarehouseByID(ctx context.Context, id uuid.UUID) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	var description sql.NullString // Используем sql.NullString для обработки NULL
	query := `SELECT id, name, address, description FROM warehouses WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&warehouse.ID, &warehouse.Name, &warehouse.Address, &description)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrWarehouseNotFound
		}
		return nil, err
	}
	warehouse.Description = description.String // Преобразуем в строку
	return &warehouse, nil
}
