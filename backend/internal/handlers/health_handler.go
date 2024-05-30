package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/mjuopperi/hp/backend/internal/db"
	"github.com/mjuopperi/hp/backend/internal/models"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		models.RegisterCustomValidations(v)
	}

	health := router.Group("/health")
	{
		health.GET("/valid-units", getValidUnits)
		health.GET("/display-names", getDisplayNames)
		health.GET("", getMeasurements)
		health.POST("", addMeasurement)
	}
}

func getValidUnits(c *gin.Context) {
	c.JSON(http.StatusOK, models.ValidUnits)
}

func getDisplayNames(c *gin.Context) {
	c.JSON(http.StatusOK, models.DisplayNames)
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
		fmt.Printf("Binding error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.Struct(dpi); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	err := db.AddDataPoint(c.Request.Context(), &dpi)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}
