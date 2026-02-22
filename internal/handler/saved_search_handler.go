package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
	"github.com/quckapp/search-service/internal/service"
)

type SavedSearchHandler struct {
	service *service.SavedSearchService
	logger  *logrus.Logger
}

func NewSavedSearchHandler(svc *service.SavedSearchService, logger *logrus.Logger) *SavedSearchHandler {
	return &SavedSearchHandler{service: svc, logger: logger}
}

func (h *SavedSearchHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateSavedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	saved, err := h.service.Create(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save search"})
		return
	}
	c.JSON(http.StatusCreated, saved)
}

func (h *SavedSearchHandler) GetByID(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	searchID := c.Param("id")
	saved, err := h.service.GetByID(c.Request.Context(), userID, searchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Saved search not found"})
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (h *SavedSearchHandler) GetByUser(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	searches, err := h.service.GetByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get saved searches"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"saved_searches": searches})
}

func (h *SavedSearchHandler) Update(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	searchID := c.Param("id")
	var req models.UpdateSavedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	saved, err := h.service.Update(c.Request.Context(), userID, searchID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update saved search"})
		return
	}
	c.JSON(http.StatusOK, saved)
}

func (h *SavedSearchHandler) Delete(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	searchID := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), userID, searchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete saved search"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
