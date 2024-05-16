package handlers

import (
	"net/http"

	"github.com/mjuopperi/hp/backend/internal/db"
	"github.com/mjuopperi/hp/backend/internal/models"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	health := router.Group("/health")
	{
		health.GET("", getMeasurements)
		health.POST("", addMeasurement)
	}
}

func getMeasurements(c *gin.Context) {
	measurementType := models.Measurement(c.Query("measurement"))

	measurements, err := db.GetDataPoints(c.Request.Context(), measurementType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, measurements)
}

func addMeasurement(c *gin.Context) {
	var dpi models.DataPointIn
	if err := c.ShouldBindJSON(&dpi); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := db.AddDataPoint(c.Request.Context(), &dpi)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}
