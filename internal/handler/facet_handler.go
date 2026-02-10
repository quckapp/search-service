package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
	"github.com/quckapp/search-service/internal/service"
)

type FacetHandler struct {
	service *service.FacetService
	logger  *logrus.Logger
}

func NewFacetHandler(svc *service.FacetService, logger *logrus.Logger) *FacetHandler {
	return &FacetHandler{service: svc, logger: logger}
}

func (h *FacetHandler) GetFacets(c *gin.Context) {
	var req models.FacetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.GetFacets(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get facets"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *FacetHandler) GetFacetedSearch(c *gin.Context) {
	index := c.Query("index")
	query := c.Query("q")
	facetField := c.Query("facet_field")

	if index == "" || query == "" || facetField == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index, q, and facet_field are required"})
		return
	}

	results, facets, err := h.service.GetFacetedSearch(c.Request.Context(), index, query, facetField, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Faceted search failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results, "facets": facets})
}

func (h *FacetHandler) GetFilterOptions(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	options, err := h.service.GetFilterOptions(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get filter options"})
		return
	}
	c.JSON(http.StatusOK, options)
}
