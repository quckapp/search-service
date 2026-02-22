package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
	"github.com/quckapp/search-service/internal/service"
)

type IndexManagementHandler struct {
	service *service.IndexManagementService
	logger  *logrus.Logger
}

func NewIndexManagementHandler(svc *service.IndexManagementService, logger *logrus.Logger) *IndexManagementHandler {
	return &IndexManagementHandler{service: svc, logger: logger}
}

func (h *IndexManagementHandler) ListIndices(c *gin.Context) {
	indices, err := h.service.ListIndices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list indices"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"indices": indices})
}

func (h *IndexManagementHandler) GetIndexInfo(c *gin.Context) {
	index := c.Param("index")
	if index == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Index name is required"})
		return
	}

	info, err := h.service.GetIndexInfo(c.Request.Context(), index)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get index info"})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (h *IndexManagementHandler) CreateIndex(c *gin.Context) {
	var req struct {
		Index    string                 `json:"index" binding:"required"`
		Mappings map[string]interface{} `json:"mappings"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateIndex(c.Request.Context(), req.Index, req.Mappings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Index created", "index": req.Index})
}

func (h *IndexManagementHandler) DeleteIndex(c *gin.Context) {
	index := c.Param("index")
	if index == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Index name is required"})
		return
	}

	if err := h.service.DeleteIndex(c.Request.Context(), index); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Index deleted", "index": index})
}

func (h *IndexManagementHandler) PutMapping(c *gin.Context) {
	var req models.IndexMapping
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.PutMapping(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update mapping"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Mapping updated"})
}

func (h *IndexManagementHandler) UpdateSettings(c *gin.Context) {
	var req models.IndexSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateSettings(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}

func (h *IndexManagementHandler) CreateAlias(c *gin.Context) {
	var req models.IndexAliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateAlias(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alias"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Alias created"})
}

func (h *IndexManagementHandler) DeleteAlias(c *gin.Context) {
	index := c.Query("index")
	alias := c.Query("alias")
	if index == "" || alias == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both index and alias are required"})
		return
	}

	if err := h.service.DeleteAlias(c.Request.Context(), index, alias); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete alias"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Alias deleted"})
}

func (h *IndexManagementHandler) RefreshIndex(c *gin.Context) {
	index := c.Param("index")
	if index == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Index name is required"})
		return
	}

	if err := h.service.RefreshIndex(c.Request.Context(), index); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh index"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Index refreshed", "index": index})
}

func (h *IndexManagementHandler) FlushIndex(c *gin.Context) {
	index := c.Param("index")
	if index == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Index name is required"})
		return
	}

	if err := h.service.FlushIndex(c.Request.Context(), index); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to flush index"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Index flushed", "index": index})
}
