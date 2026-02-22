package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
	"github.com/quckapp/search-service/internal/service"
)

type SearchScopeHandler struct {
	service *service.SearchScopeService
	logger  *logrus.Logger
}

func NewSearchScopeHandler(svc *service.SearchScopeService, logger *logrus.Logger) *SearchScopeHandler {
	return &SearchScopeHandler{service: svc, logger: logger}
}

func (h *SearchScopeHandler) SetScope(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.SetSearchScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scope, err := h.service.SetScope(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set search scope"})
		return
	}
	c.JSON(http.StatusOK, scope)
}

func (h *SearchScopeHandler) GetScope(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	scope, err := h.service.GetScope(c.Request.Context(), userID, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get search scope"})
		return
	}
	c.JSON(http.StatusOK, scope)
}

func (h *SearchScopeHandler) ListAvailableScopes(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	scopes, err := h.service.ListAvailableScopes(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list available scopes"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"available_scopes": scopes})
}
