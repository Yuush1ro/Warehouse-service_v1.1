package models

import "github.com/google/uuid"

type Warehouse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Description string    `json:"description,omitempty"`
}
