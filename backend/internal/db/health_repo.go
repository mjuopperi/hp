package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mjuopperi/hp/backend/internal/models"
)

func AddDataPoint(ctx context.Context, dataPoint *models.DataPointIn) error {
	fmt.Printf("Adding data point %v\n", dataPoint)
	query := `
		insert into health_data (timestamp, measurement, value, unit) 
		values (@timestamp, @measurement, @value, @unit)`
	args := pgx.NamedArgs{
		"timestamp":   time.Now(),
		"measurement": dataPoint.Measurement,
		"value":       dataPoint.Value,
		"unit":        dataPoint.Unit,
	}
	_, err := Pool.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to insert data point: %w", err)
	}

	return nil
}

func GetDataPoints(ctx context.Context, measurementType models.Measurement) ([]models.DataPoint, error) {
	query := `select id, timestamp, measurement, value, unit from health_data where measurement = $1;`
	rows, err := Pool.Query(ctx, query, measurementType)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dataPoints := make([]models.DataPoint, 0)
	for rows.Next() {
		var dp models.DataPoint
		if err := rows.Scan(&dp.ID, &dp.Timestamp, &dp.Measurement, &dp.Value, &dp.Unit); err != nil {
			return nil, err
		}
		dataPoints = append(dataPoints, dp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dataPoints, nil
}
