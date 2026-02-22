package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/service"
)

type AnalyticsHandler struct {
	service *service.AnalyticsService
	logger  *logrus.Logger
}

func NewAnalyticsHandler(svc *service.AnalyticsService, logger *logrus.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{service: svc, logger: logger}
}

func (h *AnalyticsHandler) GetAnalytics(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	analytics, err := h.service.GetAnalytics(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analytics"})
		return
	}
	c.JSON(http.StatusOK, analytics)
}

func (h *AnalyticsHandler) GetPopularQueries(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	limit := int64(10)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	queries, err := h.service.GetPopularQueries(c.Request.Context(), workspaceID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get popular queries"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"popular_queries": queries})
}

func (h *AnalyticsHandler) ClearAnalytics(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	if err := h.service.ClearAnalytics(c.Request.Context(), workspaceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear analytics"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Analytics cleared"})
}
