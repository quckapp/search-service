package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/service"
)

type SpellCheckHandler struct {
	service *service.SpellCheckService
	logger  *logrus.Logger
}

func NewSpellCheckHandler(svc *service.SpellCheckService, logger *logrus.Logger) *SpellCheckHandler {
	return &SpellCheckHandler{service: svc, logger: logger}
}

func (h *SpellCheckHandler) GetSuggestions(c *gin.Context) {
	text := c.Query("q")
	index := c.Query("index")
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	result, err := h.service.GetSuggestions(c.Request.Context(), text, index)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get spelling suggestions"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *SpellCheckHandler) DidYouMean(c *gin.Context) {
	text := c.Query("q")
	index := c.Query("index")
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	result, err := h.service.DidYouMean(c.Request.Context(), text, index)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get suggestions"})
		return
	}
	c.JSON(http.StatusOK, result)
}
