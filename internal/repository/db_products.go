package repository

import (
	"context"
	"errors"

	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/warehouse-service/internal/models"
)

type ProductRepository interface {
	GetAll(ctx context.Context) ([]models.Product, error)
	Create(ctx context.Context, product models.Product) error
	Update(ctx context.Context, product models.Product) error
	Delete(ctx context.Context, id string) error
}

type ProductRepositoryImpl struct {
	db *pgxpool.Pool
}

var _ ProductRepository = (*ProductRepositoryImpl)(nil)

func NewProductRepository(db *pgxpool.Pool) *ProductRepositoryImpl {
	return &ProductRepositoryImpl{db: db}
}

func (r *ProductRepositoryImpl) GetAll(ctx context.Context) ([]models.Product, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name, description, attributes, weight, barcode FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		var attributes []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &attributes, &p.Weight, &p.Barcode); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(attributes, &p.Attributes); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepositoryImpl) Create(ctx context.Context, product models.Product) error {
	_, err := r.db.Exec(ctx, "INSERT INTO products (id, name, description, attributes, weight, barcode) VALUES ($1, $2, $3, $4, $5, $6)",
		product.ID, product.Name, product.Description, product.Attributes, product.Weight, product.Barcode)
	return err
}

func (r *ProductRepositoryImpl) Update(ctx context.Context, product models.Product) error {
	query := `
		UPDATE products
		SET 
			name = COALESCE(NULLIF($1, ''), name),
			description = COALESCE(NULLIF($2, ''), description),
			attributes = COALESCE($3, attributes),
			weight = COALESCE(NULLIF($4, 0), weight),
			barcode = COALESCE(NULLIF($5, ''), barcode)
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query,
		product.Name, product.Description, product.Attributes,
		product.Weight, product.Barcode, product.ID,
	)
	return err
}

func (r *ProductRepositoryImpl) Delete(ctx context.Context, id string) error {
	commandTag, err := r.db.Exec(ctx, "DELETE FROM products WHERE id=$1", id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return errors.New("product not found")
	}
	return nil
}
