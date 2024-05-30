package models

import (
	"log"
	"time"

	"github.com/go-playground/validator/v10"
)

type Measurement string

const (
	Weight                 Measurement = "weight"
	SystolicBloodPressure  Measurement = "sbp"
	DiastolicBloodPressure Measurement = "dbp"
)

var ValidUnits = map[Measurement][]string{
	Weight:                 {"kg", "lbs"},
	SystolicBloodPressure:  {"mmHg"},
	DiastolicBloodPressure: {"mmHg"},
}

var DisplayNames = map[string]map[Measurement]string{
	"en": {
		Weight:                 "Weight",
		SystolicBloodPressure:  "Systolic Blood Pressure",
		DiastolicBloodPressure: "Diastolic Blood Pressure",
	},
	"fi": {
		Weight:                 "Paino",
		SystolicBloodPressure:  "Verenpaine, yläpaine",
		DiastolicBloodPressure: "Verenpaine, alapaine",
	},
}

type DataPointIn struct {
	Measurement Measurement `json:"measurement" binding:"required,oneof=weight sbp dbp"`
	Value       float64     `json:"value" binding:"required"`
	Unit        string      `json:"unit" binding:"required,unit"`
}

type DataPoint struct {
	ID          int         `json:"id"`
	Timestamp   time.Time   `json:"timestamp"`
	Measurement Measurement `json:"measurement"`
	Value       float64     `json:"value"`
	Unit        string      `json:"unit"`
}

func UnitValidation(fl validator.FieldLevel) bool {
	unit := fl.Field().String()

	if dp, ok := fl.Parent().Interface().(DataPointIn); ok {
		if units, exists := ValidUnits[dp.Measurement]; exists {
			for _, validUnit := range units {
				if unit == validUnit {
					return true
				}
			}
		}
	}
	return false
}

func RegisterCustomValidations(v *validator.Validate) {
	if err := v.RegisterValidation("unit", UnitValidation); err != nil {
		log.Fatalf("Failed to register custom validation for unit: %v", err)
	}
}
