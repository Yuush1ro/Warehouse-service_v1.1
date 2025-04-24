package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Inventory struct {
	ID          uuid.UUID       `json:"id"`
	ProductID   uuid.UUID       `json:"product_id"`
	WarehouseID uuid.UUID       `json:"warehouse_id"`
	Quantity    int             `json:"quantity"`
	Price       decimal.Decimal `json:"price"`
	Discount    decimal.Decimal `json:"discount"`
}

type InventoryWithNames struct {
	ID            uuid.UUID       `json:"id"`
	ProductID     uuid.UUID       `json:"product_id"`
	WarehouseID   uuid.UUID       `json:"warehouse_id"`
	Quantity      int             `json:"quantity"`
	Price         decimal.Decimal `json:"price"`
	Discount      decimal.Decimal `json:"discount"`
	WarehouseName string          `json:"warehouse_name"`
	ProductName   string          `json:"product_name"`
}
