package repository

import (
	"context"
	"errors"

	//"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/warehouse-service/internal/models"
)

type InventoryRepository interface {
	Create(ctx context.Context, inventory models.Inventory) error
	UpdateQuantity(ctx context.Context, productID, warehouseID uuid.UUID, quantity int) error
	SetDiscount(ctx context.Context, productIDs []uuid.UUID, warehouseID uuid.UUID, discount float64) error
	GetByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]models.InventoryWithNames, error)
	GetProductInWarehouse(ctx context.Context, productID, warehouseID uuid.UUID) (*models.Inventory, error)
	CalculateTotal(ctx context.Context, warehouseID uuid.UUID, items map[uuid.UUID]int) (float64, error)
	Purchase(ctx context.Context, warehouseID uuid.UUID, items map[uuid.UUID]int) error
	GetProductPrice(ctx context.Context, warehouseID uuid.UUID, productID uuid.UUID) (float64, error)
	GetProductDiscount(ctx context.Context, warehouseID, productID uuid.UUID) (float64, error)
	DeleteProductFromWarehouse(ctx context.Context, warehouseID uuid.UUID, productID uuid.UUID) error
	DeleteInventory(ctx context.Context, inventoryID uuid.UUID) error
}

type InventoryRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewInventoryRepository(db *pgxpool.Pool) *InventoryRepositoryImpl {
	return &InventoryRepositoryImpl{db: db}
}

// 1. Создание связи товара и склада (указание цены)
func (r *InventoryRepositoryImpl) Create(ctx context.Context, inventory models.Inventory) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO inventory (product_id, warehouse_id, quantity, price, discount) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (product_id, warehouse_id) 
		DO UPDATE SET 
			quantity = inventory.quantity + EXCLUDED.quantity,
			price = EXCLUDED.price,
			discount = EXCLUDED.discount
	`, inventory.ProductID, inventory.WarehouseID, inventory.Quantity, inventory.Price, inventory.Discount)
	return err
}

// 2. Обновление количества товара (поступление на склад)
func (r *InventoryRepositoryImpl) UpdateQuantity(ctx context.Context, productID, warehouseID uuid.UUID, quantity int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE inventory SET quantity = quantity + $1 WHERE product_id = $2 AND warehouse_id = $3
	`, quantity, productID, warehouseID)
	return err
}

// 3. Установка скидки на список товаров
func (r *InventoryRepositoryImpl) SetDiscount(ctx context.Context, productIDs []uuid.UUID, warehouseID uuid.UUID, discount float64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE inventory SET discount = $1 WHERE product_id = ANY($2) AND warehouse_id = $3
	`, discount, productIDs, warehouseID)
	return err
}

// 4. Получение списка товаров на складе (с пагинацией)
func (r *InventoryRepositoryImpl) GetByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]models.InventoryWithNames, error) {
	rows, err := r.db.Query(ctx, `
		SELECT 
			i.id, i.product_id, i.warehouse_id, i.quantity, i.price, i.discount,
			w.name AS warehouse_name, p.name AS product_name
		FROM inventory i
		JOIN warehouses w ON i.warehouse_id = w.id
		JOIN products p ON i.product_id = p.id
		WHERE i.warehouse_id = $1
		LIMIT $2 OFFSET $3
	`, warehouseID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventoryList []models.InventoryWithNames
	for rows.Next() {
		var inv models.InventoryWithNames
		if err := rows.Scan(&inv.ID, &inv.ProductID, &inv.WarehouseID, &inv.Quantity, &inv.Price, &inv.Discount, &inv.WarehouseName, &inv.ProductName); err != nil {
			return nil, err
		}
		inventoryList = append(inventoryList, inv)
	}
	return inventoryList, nil
}

// 5. Получение информации о товаре на складе
func (r *InventoryRepositoryImpl) GetProductInWarehouse(ctx context.Context, productID, warehouseID uuid.UUID) (*models.Inventory, error) {
	var inv models.Inventory
	err := r.db.QueryRow(ctx, `
		SELECT id, product_id, warehouse_id, quantity, price, discount FROM inventory
		WHERE product_id = $1 AND warehouse_id = $2
	`, productID, warehouseID).Scan(&inv.ID, &inv.ProductID, &inv.WarehouseID, &inv.Quantity, &inv.Price, &inv.Discount)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// 6. Подсчёт стоимости товаров
func (r *InventoryRepositoryImpl) CalculateTotal(ctx context.Context, warehouseID uuid.UUID, items map[uuid.UUID]int) (float64, error) {
	var total float64
	for productID, quantity := range items {
		var price, discount float64
		err := r.db.QueryRow(ctx, `
			SELECT price, discount FROM inventory WHERE product_id = $1 AND warehouse_id = $2
		`, productID, warehouseID).Scan(&price, &discount)
		if err != nil {
			return 0, err
		}
		total += float64(quantity) * (price * (1 - discount/100))
	}
	return total, nil
}

// 7. Покупка товаров (уменьшение количества)
func (r *InventoryRepositoryImpl) Purchase(ctx context.Context, warehouseID uuid.UUID, items map[uuid.UUID]int) error {
	for productID, quantity := range items {
		result, err := r.db.Exec(ctx, `
			UPDATE inventory SET quantity = quantity - $1 WHERE product_id = $2 AND warehouse_id = $3 AND quantity >= $1
		`, quantity, productID, warehouseID)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			return errors.New("not enough stock for product " + productID.String())
		}
	}
	return nil
}

func (r *InventoryRepositoryImpl) GetProductPrice(ctx context.Context, warehouseID uuid.UUID, productID uuid.UUID) (float64, error) {
	var price float64
	err := r.db.QueryRow(ctx, "SELECT price FROM inventory WHERE warehouse_id = $1 AND product_id = $2",
		warehouseID, productID).Scan(&price)
	if err != nil {
		return 0, err
	}
	return price, nil
}

func (r *InventoryRepositoryImpl) GetProductDiscount(ctx context.Context, warehouseID, productID uuid.UUID) (float64, error) {
	var discount float64
	err := r.db.QueryRow(ctx, `
		SELECT discount 
		FROM inventory 
		WHERE warehouse_id = $1 AND product_id = $2
	`, warehouseID, productID).Scan(&discount)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, nil // Если скидка отсутствует, возвращаем 0
		}
		return 0, err
	}
	return discount, nil
}

func (r *InventoryRepositoryImpl) DeleteProductFromWarehouse(ctx context.Context, productID, warehouseID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM inventory WHERE product_id = $1 AND warehouse_id = $2`, productID, warehouseID)
	return err
}

func (r *InventoryRepositoryImpl) DeleteInventory(ctx context.Context, inventoryID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM inventory WHERE id = $1`, inventoryID)
	return err
}
