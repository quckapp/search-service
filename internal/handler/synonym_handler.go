package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
	"github.com/quckapp/search-service/internal/service"
)

type SynonymHandler struct {
	service *service.SynonymService
	logger  *logrus.Logger
}

func NewSynonymHandler(svc *service.SynonymService, logger *logrus.Logger) *SynonymHandler {
	return &SynonymHandler{service: svc, logger: logger}
}

func (h *SynonymHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateSynonymRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.service.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create synonym group"})
		return
	}
	c.JSON(http.StatusCreated, group)
}

func (h *SynonymHandler) List(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	groups, err := h.service.List(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list synonym groups"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"synonym_groups": groups})
}

func (h *SynonymHandler) Update(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	synonymID := c.Param("id")
	if workspaceID == "" || synonymID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id and synonym ID are required"})
		return
	}

	var req models.UpdateSynonymRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.service.Update(c.Request.Context(), workspaceID, synonymID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update synonym group"})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *SynonymHandler) Delete(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	synonymID := c.Param("id")
	if workspaceID == "" || synonymID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id and synonym ID are required"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), workspaceID, synonymID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete synonym group"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *SynonymHandler) ApplyToIndex(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	count, err := h.service.ApplyToIndex(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply synonyms"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Synonyms applied", "groups_applied": count})
}
