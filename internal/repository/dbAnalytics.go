package repository

import (
	"context"

	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/warehouse-service/internal/models"
	"go.uber.org/zap"
)

type AnalyticsRepository interface {
	RecordSale(ctx context.Context, warehouseID, productID uuid.UUID, quantity int, totalSum float64) error
	GetWarehouseAnalytics(ctx context.Context, warehouseID uuid.UUID) ([]models.Analytics, error)
	GetTopWarehouses(ctx context.Context, limit int) ([]struct {
		WarehouseID uuid.UUID `json:"warehouse_id"`
		Address     string    `json:"address"`
		TotalSum    float64   `json:"total_sum"`
	}, error)
	DeleteAnalytics(ctx context.Context, warehouseID, productID uuid.UUID) error
}

type AnalyticsRepositoryImpl struct {
	db     *pgxpool.Pool
	Logger *zap.Logger
}

func NewAnalyticsRepository(db *pgxpool.Pool, logger *zap.Logger) *AnalyticsRepositoryImpl {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AnalyticsRepositoryImpl{db: db, Logger: logger}
}

// 1. Запись продажи в аналитику
func (r *AnalyticsRepositoryImpl) RecordSale(ctx context.Context, warehouseID, productID uuid.UUID, quantity int, totalSum float64) error {
	r.Logger.Info("Recording sale",
		zap.String("warehouseID", warehouseID.String()),
		zap.String("productID", productID.String()),
		zap.Int("quantity", quantity),
		zap.Float64("totalPrice", totalSum))

	_, err := r.db.Exec(ctx, `
		INSERT INTO analytics (warehouse_id, product_id, sold_quantity, total_sum) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (warehouse_id, product_id) 
		DO UPDATE SET 
			sold_quantity = analytics.sold_quantity + EXCLUDED.sold_quantity, 
			total_sum = analytics.total_sum + EXCLUDED.total_sum
	`, warehouseID, productID, quantity, totalSum)

	if err != nil {
		r.Logger.Error("Failed to execute RecordSale query", zap.Error(err))
	} else {
		r.Logger.Info("RecordSale query executed successfully",
			zap.String("warehouseID", warehouseID.String()),
			zap.String("productID", productID.String()),
			zap.Int("quantity", quantity),
			zap.Float64("totalSum", totalSum))
	}
	return err
}

// 2. Получение аналитики по складу
func (r *AnalyticsRepositoryImpl) GetWarehouseAnalytics(ctx context.Context, warehouseID uuid.UUID) ([]models.Analytics, error) {
	r.Logger.Info("Executing analytics query", zap.String("warehouseId", warehouseID.String()))

	rows, err := r.db.Query(ctx, `
		SELECT id, warehouse_id, product_id, sold_quantity, total_sum 
		FROM analytics WHERE warehouse_id = $1`, warehouseID)

	if err != nil {
		log.Println("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	var analytics []models.Analytics
	for rows.Next() {
		var a models.Analytics
		if err := rows.Scan(&a.ID, &a.WarehouseID, &a.ProductID, &a.Quantity, &a.TotalSum); err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}
		analytics = append(analytics, a)
	}
	log.Println("Returning analytics data:", analytics)
	return analytics, nil
}

// 3. Топ-10 складов по выручке
func (r *AnalyticsRepositoryImpl) GetTopWarehouses(ctx context.Context, limit int) ([]struct {
	WarehouseID uuid.UUID `json:"warehouse_id"`
	Address     string    `json:"address"`
	TotalSum    float64   `json:"total_sum"`
}, error) {
	r.Logger.Info("Executing query to fetch top warehouses by revenue", zap.Int("limit", limit))

	rows, err := r.db.Query(ctx, `
		SELECT w.id, w.address, COALESCE(SUM(a.total_sum), 0) AS total_sum
		FROM analytics a
		JOIN warehouses w ON a.warehouse_id = w.id
		GROUP BY w.id, w.address
		ORDER BY total_sum DESC
		LIMIT $1
	`, limit)
	if err != nil {
		r.Logger.Error("Failed to execute query for top warehouses", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		WarehouseID uuid.UUID `json:"warehouse_id"`
		Address     string    `json:"address"`
		TotalSum    float64   `json:"total_sum"`
	}
	for rows.Next() {
		var res struct {
			WarehouseID uuid.UUID `json:"warehouse_id"`
			Address     string    `json:"address"`
			TotalSum    float64   `json:"total_sum"`
		}
		if err := rows.Scan(&res.WarehouseID, &res.Address, &res.TotalSum); err != nil {
			r.Logger.Error("Error scanning row for top warehouses", zap.Error(err))
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}

func (r *AnalyticsRepositoryImpl) DeleteAnalytics(ctx context.Context, warehouseID, productID uuid.UUID) error {
	r.Logger.Info("Deleting analytics data",
		zap.String("warehouseID", warehouseID.String()),
		zap.String("productID", productID.String()))

	_, err := r.db.Exec(ctx, `
		DELETE FROM analytics 
		WHERE warehouse_id = $1 AND product_id = $2
	`, warehouseID, productID)
	return err
}
