// Package db provides database interaction logic for analytics aggregations.
package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/analytics/models"
)

// AggregationRepository provides methods to interact with the aggregation data.
// It implements the AggregationRepositoryInterface.
type AggregationRepository struct {
	db *pgxpool.Pool
}

// NewAggregationRepository creates a new instance of AggregationRepository.
func NewAggregationRepository(db *pgxpool.Pool) *AggregationRepository {
	return &AggregationRepository{db: db}
}

// Upsert creates or updates an aggregation for a time bucket.
// It ensures that the aggregation data remains up-to-date.
func (r *AggregationRepository) Upsert(ctx context.Context, agg *models.Aggregation) error {
	query := `
		INSERT INTO analytics.aggregations (metric_type, service, value, time_bucket, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (metric_type, service, time_bucket)
		DO UPDATE SET value = EXCLUDED.value, created_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, agg.MetricType, agg.Service, agg.Value, agg.TimeBucket)
	return err
}

// FindByRange retrieves aggregations within a specified time range.
// It filters results by metric type and service.
func (r *AggregationRepository) FindByRange(ctx context.Context, metricType models.MetricType, service string, start, end time.Time) ([]*models.Aggregation, error) {
	query := `
		SELECT id, metric_type, service, value, time_bucket, created_at
		FROM analytics.aggregations
		WHERE metric_type = $1 AND service = $2 AND time_bucket BETWEEN $3 AND $4
		ORDER BY time_bucket ASC
	`
	rows, err := r.db.Query(ctx, query, metricType, service, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aggregations []*models.Aggregation
	for rows.Next() {
		agg := &models.Aggregation{}
		if err := rows.Scan(&agg.ID, &agg.MetricType, &agg.Service, &agg.Value, &agg.TimeBucket, &agg.CreatedAt); err != nil {
			return nil, err
		}
		aggregations = append(aggregations, agg)
	}
	return aggregations, nil
}

// FindAllServices returns a list of all services that have aggregations.
// The services are returned in alphabetical order.
func (r *AggregationRepository) FindAllServices(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT service FROM analytics.aggregations ORDER BY service`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	return services, nil
}

// AggregationRepositoryInterface defines the methods for aggregation repository operations.
// This ensures compatibility with services.
type AggregationRepositoryInterface interface {
	Upsert(ctx context.Context, agg *models.Aggregation) error
	FindByRange(ctx context.Context, metricType models.MetricType, service string, start, end time.Time) ([]*models.Aggregation, error)
	FindAllServices(ctx context.Context) ([]string, error)
}
