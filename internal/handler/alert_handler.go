package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
	"github.com/quckapp/search-service/internal/service"
)

type AlertHandler struct {
	service *service.AlertService
	logger  *logrus.Logger
}

func NewAlertHandler(svc *service.AlertService, logger *logrus.Logger) *AlertHandler {
	return &AlertHandler{service: svc, logger: logger}
}

func (h *AlertHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alert, err := h.service.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert"})
		return
	}
	c.JSON(http.StatusCreated, alert)
}

func (h *AlertHandler) List(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	alerts, err := h.service.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list alerts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}

func (h *AlertHandler) Update(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	alertID := c.Param("id")
	var req models.UpdateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alert, err := h.service.Update(c.Request.Context(), userID, alertID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update alert"})
		return
	}
	c.JSON(http.StatusOK, alert)
}

func (h *AlertHandler) Delete(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	alertID := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), userID, alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete alert"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *AlertHandler) GetHistory(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alert ID is required"})
		return
	}

	history, err := h.service.GetHistory(c.Request.Context(), alertID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get alert history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"history": history})
}
