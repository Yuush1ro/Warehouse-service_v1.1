package models

import "github.com/google/uuid"

type Analytics struct {
	ID          uuid.UUID `json:"id"`
	WarehouseID uuid.UUID `json:"warehouse_id"`
	ProductID   uuid.UUID `json:"product_id"`
	Quantity    int       `json:"sold_quantity"`
	TotalSum    float64   `json:"total_sum"`
}
