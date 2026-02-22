package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
	"github.com/quckapp/search-service/internal/service"
)

type SearchHandler struct {
	service *service.SearchService
	logger  *logrus.Logger
}

func NewSearchHandler(svc *service.SearchService, logger *logrus.Logger) *SearchHandler {
	return &SearchHandler{service: svc, logger: logger}
}

// ── Search Endpoints ──

func (h *SearchHandler) GlobalSearch(c *gin.Context) {
	var params models.SearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if params.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	result, err := h.service.GlobalSearch(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *SearchHandler) SearchMessages(c *gin.Context) {
	var params models.SearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if params.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	result, err := h.service.SearchMessages(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *SearchHandler) SearchFiles(c *gin.Context) {
	var params models.SearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if params.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	result, err := h.service.SearchFiles(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *SearchHandler) SearchUsers(c *gin.Context) {
	var params models.SearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if params.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	result, err := h.service.SearchUsers(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *SearchHandler) SearchChannels(c *gin.Context) {
	var params models.SearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if params.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	result, err := h.service.SearchChannels(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *SearchHandler) Suggest(c *gin.Context) {
	query := c.Query("q")
	workspaceID := c.Query("workspace_id")
	if query == "" {
		c.JSON(http.StatusOK, models.SuggestionResponse{Suggestions: []string{}})
		return
	}

	result, err := h.service.Suggest(c.Request.Context(), query, workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Suggestion failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ── Index Endpoints ──

func (h *SearchHandler) IndexDocument(c *gin.Context) {
	var req models.IndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.IndexDocument(c.Request.Context(), req.Index, req.ID, req.Document); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index document"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"indexed": true, "id": req.ID})
}

func (h *SearchHandler) IndexMessage(c *gin.Context) {
	var doc map[string]interface{}
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, ok := doc["id"].(string)
	if !ok || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document must have an 'id' field"})
		return
	}

	if err := h.service.IndexDocument(c.Request.Context(), "quckapp_messages", id, doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index message"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"indexed": true})
}

func (h *SearchHandler) IndexFile(c *gin.Context) {
	var doc map[string]interface{}
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, ok := doc["id"].(string)
	if !ok || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document must have an 'id' field"})
		return
	}

	if err := h.service.IndexDocument(c.Request.Context(), "quckapp_files", id, doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index file"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"indexed": true})
}

func (h *SearchHandler) BulkIndex(c *gin.Context) {
	var req models.BulkIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.service.BulkIndex(c.Request.Context(), req.Documents)
	c.JSON(http.StatusOK, result)
}

func (h *SearchHandler) DeleteFromIndex(c *gin.Context) {
	indexType := c.Param("type")
	id := c.Param("id")

	index := "quckapp_" + indexType
	if err := h.service.DeleteDocument(c.Request.Context(), index, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *SearchHandler) Reindex(c *gin.Context) {
	var req models.ReindexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Reindex(c.Request.Context(), req.Index); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Reindex failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reindex started", "index": req.Index})
}

// ── Health ──

func (h *SearchHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.HealthCheck())
}
