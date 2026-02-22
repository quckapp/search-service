package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/service"
)

type HistoryHandler struct {
	service *service.HistoryService
	logger  *logrus.Logger
}

func NewHistoryHandler(svc *service.HistoryService, logger *logrus.Logger) *HistoryHandler {
	return &HistoryHandler{service: svc, logger: logger}
}

func (h *HistoryHandler) GetHistory(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit := int64(20)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	history, err := h.service.GetHistory(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get search history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"history": history})
}

func (h *HistoryHandler) ClearHistory(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.service.ClearHistory(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "History cleared"})
}

func (h *HistoryHandler) DeleteHistoryItem(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	historyID := c.Param("id")
	if historyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "History ID is required"})
		return
	}

	if err := h.service.DeleteHistoryItem(c.Request.Context(), userID, historyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete history item"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "History item deleted"})
}

// getUserID extracts user ID from the gin context (set by auth middleware)
func getUserID(c *gin.Context) string {
	uid, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	if s, ok := uid.(string); ok {
		return s
	}
	return ""
}
