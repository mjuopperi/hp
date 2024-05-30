package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/mjuopperi/hp/backend/internal/db"
	"github.com/mjuopperi/hp/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	RegisterRoutes(router)
	return router
}

const DBName = "hp"
const DBUser = "user"
const DBPassword = "password"

func setupPostgres(t *testing.T) func() {
	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase(DBName),
		postgres.WithUsername(DBUser),
		postgres.WithPassword(DBPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	host, _ := postgresContainer.Host(ctx)
	port, _ := postgresContainer.MappedPort(ctx, "5432")
	err = db.InitDB(db.ConnectionURI(host, port.Int(), DBUser, DBPassword, DBName))
	if err != nil {
		t.Fatalf("failed to initialize database: %s", err)
	}

	return func() {
		db.Close()

		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}
}

func TestGetValidUnits(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest(http.MethodGet, "/health/valid-units", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	expectedBody := `{"dbp":["mmHg"],"sbp":["mmHg"],"weight":["kg","lbs"]}`
	assert.JSONEq(t, expectedBody, resp.Body.String())
}

func TestGetDisplayNames(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest(http.MethodGet, "/health/display-names", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	expectedBody := `{
		"en": {
			"weight": "Weight",
			"sbp": "Systolic Blood Pressure",
			"dbp": "Diastolic Blood Pressure"
		},
		"fi": {
			"weight": "Paino",
			"sbp": "Verenpaine, yläpaine",
			"dbp": "Verenpaine, alapaine"
		}
	}`
	assert.JSONEq(t, expectedBody, resp.Body.String())
}

func TestGetMeasurements(t *testing.T) {
	router := setupRouter()
	setupPostgres(t)

	now := time.Now()

	measurements := []models.DataPoint{
		models.DataPoint{
			ID:          0,
			Timestamp:   now,
			Measurement: "weight",
			Value:       98,
			Unit:        "kg",
		},
		models.DataPoint{
			ID:          1,
			Timestamp:   now.AddDate(0, 0, -1),
			Measurement: "weight",
			Value:       99,
			Unit:        "kg",
		},
		models.DataPoint{
			ID:          2,
			Timestamp:   now.AddDate(0, 0, -2),
			Measurement: "weight",
			Value:       100,
			Unit:        "kg",
		},
	}

	addMeasurements(t, measurements)

	req, _ := http.NewRequest(http.MethodGet, "/health?measurement=weight", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var responseMeasurements []models.DataPoint
	if err := json.Unmarshal(resp.Body.Bytes(), &responseMeasurements); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	assert.Equal(t, len(measurements), len(responseMeasurements))
}

func TestPostMeasurement(t *testing.T) {
	router := setupRouter()
	cleanup := setupPostgres(t)
	defer cleanup()

	body := models.DataPointIn{Measurement: "weight", Value: 100, Unit: "kg"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/health", bytes.NewBuffer(jsonBody))
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	storedMeasurements, err := db.GetDataPoints(context.Background(), "weight")
	assert.NoError(t, err)
	assert.Len(t, storedMeasurements, 1)
	assert.Equal(t, models.Measurement("weight"), storedMeasurements[0].Measurement)
	assert.Equal(t, 100.0, storedMeasurements[0].Value)
	assert.Equal(t, "kg", storedMeasurements[0].Unit)
	assert.WithinDuration(t, time.Now(), storedMeasurements[0].Timestamp, time.Second)
}

func TestPostMeasurementUnsupportedMeasurement(t *testing.T) {
	router := setupRouter()
	cleanup := setupPostgres(t)
	defer cleanup()

	body := models.DataPointIn{Measurement: "enlightenment", Value: 100, Unit: "lumen"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/health", bytes.NewBuffer(jsonBody))
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestPostMeasurementUnsupportedUnit(t *testing.T) {
	router := setupRouter()
	cleanup := setupPostgres(t)
	defer cleanup()

	body := models.DataPointIn{Measurement: "weight", Value: 100, Unit: "rock"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/health", bytes.NewBuffer(jsonBody))
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func addMeasurements(t *testing.T, measurements []models.DataPoint) {
	batch := &pgx.Batch{}

	for _, m := range measurements {
		query := "insert into health_data (id, timestamp, measurement, value, unit) values ($1, $2, $3, $4, $5)"
		batch.Queue(query, m.ID, m.Timestamp, m.Measurement, m.Value, m.Unit)
	}

	br := db.Pool.SendBatch(context.Background(), batch)
	defer br.Close()

	_, err := br.Exec()
	if err != nil {
		t.Fatalf("Failed to execute batch: %v", err)
	}
}
