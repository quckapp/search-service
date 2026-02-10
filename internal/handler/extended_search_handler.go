package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
	"github.com/quckapp/search-service/internal/service"
)

type ExtendedSearchHandler struct {
	service *service.ExtendedSearchService
	logger  *logrus.Logger
}

func NewExtendedSearchHandler(svc *service.ExtendedSearchService, logger *logrus.Logger) *ExtendedSearchHandler {
	return &ExtendedSearchHandler{service: svc, logger: logger}
}

// ── Search Endpoints ──

func (h *ExtendedSearchHandler) SearchBookmarks(c *gin.Context) {
	var params models.SearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if params.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	result, err := h.service.SearchBookmarks(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ExtendedSearchHandler) SearchTasks(c *gin.Context) {
	var params models.SearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if params.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	result, err := h.service.SearchTasks(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ExtendedSearchHandler) SearchEmoji(c *gin.Context) {
	query := c.Query("q")
	workspaceID := c.Query("workspace_id")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	result, err := h.service.SearchEmoji(c.Request.Context(), query, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ExtendedSearchHandler) AdvancedSearch(c *gin.Context) {
	var req models.AdvancedSearchParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.AdvancedSearch(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Advanced search failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ── Aggregation ──

func (h *ExtendedSearchHandler) Aggregate(c *gin.Context) {
	var req models.AggregationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Aggregate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Aggregation failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ── Batch Operations ──

func (h *ExtendedSearchHandler) BatchDelete(c *gin.Context) {
	var req models.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.service.BatchDelete(c.Request.Context(), &req)
	c.JSON(http.StatusOK, result)
}

func (h *ExtendedSearchHandler) UpdateDocument(c *gin.Context) {
	index := c.Param("index")
	id := c.Param("id")
	if index == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Index and ID are required"})
		return
	}

	var req models.UpdateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateDocument(c.Request.Context(), index, id, req.Document); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": true})
}

// ── Typed Index Endpoints ──

func (h *ExtendedSearchHandler) IndexUser(c *gin.Context) {
	var req models.IndexUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.IndexUser(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"indexed": true, "id": req.ID})
}

func (h *ExtendedSearchHandler) IndexChannel(c *gin.Context) {
	var req models.IndexChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.IndexChannel(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index channel"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"indexed": true, "id": req.ID})
}

func (h *ExtendedSearchHandler) IndexBookmark(c *gin.Context) {
	var req models.IndexBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.IndexBookmark(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index bookmark"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"indexed": true, "id": req.ID})
}

func (h *ExtendedSearchHandler) IndexTask(c *gin.Context) {
	var req models.IndexTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.IndexTask(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index task"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"indexed": true, "id": req.ID})
}

func (h *ExtendedSearchHandler) CountDocuments(c *gin.Context) {
	index := c.Param("index")
	if index == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Index name is required"})
		return
	}

	count, err := h.service.CountDocuments(c.Request.Context(), index)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count documents"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"index": index, "count": count})
}

// ── Popular Queries (from analytics) ──

func (h *ExtendedSearchHandler) GetPopularQueries(c *gin.Context) {
	// This is a pass-through to show popular queries can be surfaced in search context
	workspaceID := c.Query("workspace_id")
	limit := int64(10)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	_ = workspaceID
	_ = limit
	c.JSON(http.StatusOK, gin.H{"message": "Use /analytics/popular-queries endpoint"})
}
