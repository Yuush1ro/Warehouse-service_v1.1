package models

type Product struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Attributes  map[string]string `json:"attributes"`
	Weight      float64           `json:"weight"`
	Barcode     string            `json:"barcode"`
}
