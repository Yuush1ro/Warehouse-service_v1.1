package models

import "github.com/google/uuid"

type Inventory struct {
	ID          uuid.UUID `json:"id"`
	ProductID   uuid.UUID `json:"product_id"`
	WarehouseID uuid.UUID `json:"warehouse_id"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
	Discount    float64   `json:"discount"`
}

type InventoryWithNames struct {
	ID            uuid.UUID `json:"id"`
	ProductID     uuid.UUID `json:"product_id"`
	WarehouseID   uuid.UUID `json:"warehouse_id"`
	Quantity      int       `json:"quantity"`
	Price         float64   `json:"price"`
	Discount      float64   `json:"discount"`
	WarehouseName string    `json:"warehouse_name"`
	ProductName   string    `json:"product_name"`
}
