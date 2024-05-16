package models

import "time"

type Measurement string

type DataPointIn struct {
	Measurement Measurement `json:"measurement" binding:"required,oneof=weight"`
	Value       float64     `json:"value" binding:"required"`
	Unit        string      `json:"unit" binding:"required,oneof=kg"`
}

type DataPoint struct {
	ID          int         `json:"id"`
	Timestamp   time.Time   `json:"timestamp"`
	Measurement Measurement `json:"measurement"`
	Value       float64     `json:"value"`
	Unit        string      `json:"unit"`
}
