package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
	"github.com/quckapp/search-service/internal/service"
)

type RelevanceHandler struct {
	service *service.RelevanceService
	logger  *logrus.Logger
}

func NewRelevanceHandler(svc *service.RelevanceService, logger *logrus.Logger) *RelevanceHandler {
	return &RelevanceHandler{service: svc, logger: logger}
}

func (h *RelevanceHandler) GetConfig(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	config, err := h.service.GetConfig(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get relevance config"})
		return
	}
	c.JSON(http.StatusOK, config)
}

func (h *RelevanceHandler) UpdateConfig(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	var req models.UpdateRelevanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.service.UpdateConfig(c.Request.Context(), workspaceID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update relevance config"})
		return
	}
	c.JSON(http.StatusOK, config)
}

func (h *RelevanceHandler) PreviewTuning(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	query := c.Query("q")
	index := c.Query("index")
	if workspaceID == "" || query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id and q are required"})
		return
	}

	preview, err := h.service.PreviewTuning(c.Request.Context(), workspaceID, query, index)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preview tuning"})
		return
	}
	c.JSON(http.StatusOK, preview)
}

func (h *RelevanceHandler) ResetToDefaults(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	config, err := h.service.ResetToDefaults(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset relevance config"})
		return
	}
	c.JSON(http.StatusOK, config)
}
