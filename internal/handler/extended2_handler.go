package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/service"
)

type Extended2Handler struct {
	service *service.Extended2Service
	logger  *logrus.Logger
}

func NewExtended2Handler(svc *service.Extended2Service, logger *logrus.Logger) *Extended2Handler {
	return &Extended2Handler{service: svc, logger: logger}
}

// ── Search Templates ──

func (h *Extended2Handler) CreateTemplate(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var req struct {
		Name        string         `json:"name" binding:"required"`
		Description string         `json:"description"`
		Query       string         `json:"query" binding:"required"`
		Filters     map[string]any `json:"filters"`
		IsPublic    bool           `json:"is_public"`
		WorkspaceID string         `json:"workspace_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t := &service.SearchTemplate{
		UserID:      userID,
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		Description: req.Description,
		Query:       req.Query,
		Filters:     req.Filters,
		IsPublic:    req.IsPublic,
	}
	if err := h.service.CreateTemplate(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": t})
}

func (h *Extended2Handler) GetTemplate(c *gin.Context) {
	userID := getUserID(c)
	t, err := h.service.GetTemplate(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": t})
}

func (h *Extended2Handler) ListTemplates(c *gin.Context) {
	userID := getUserID(c)
	results, err := h.service.ListTemplates(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list templates"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) UpdateTemplate(c *gin.Context) {
	userID := getUserID(c)
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateTemplate(c.Request.Context(), userID, c.Param("id"), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) DeleteTemplate(c *gin.Context) {
	userID := getUserID(c)
	if err := h.service.DeleteTemplate(c.Request.Context(), userID, c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) UseTemplate(c *gin.Context) {
	userID := getUserID(c)
	if err := h.service.IncrementTemplateUsage(c.Request.Context(), userID, c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record usage"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Search Bookmarks ──

func (h *Extended2Handler) BookmarkResult(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var req struct {
		Query      string `json:"query"`
		ResultID   string `json:"result_id" binding:"required"`
		ResultType string `json:"result_type" binding:"required"`
		Title      string `json:"title"`
		Snippet    string `json:"snippet"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b := &service.SearchBookmarkItem{
		UserID:     userID,
		Query:      req.Query,
		ResultID:   req.ResultID,
		ResultType: req.ResultType,
		Title:      req.Title,
		Snippet:    req.Snippet,
	}
	if err := h.service.BookmarkResult(c.Request.Context(), b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bookmark"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": b})
}

func (h *Extended2Handler) ListBookmarks(c *gin.Context) {
	userID := getUserID(c)
	results, err := h.service.ListBookmarks(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list bookmarks"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) DeleteBookmark(c *gin.Context) {
	userID := getUserID(c)
	if err := h.service.DeleteBookmark(c.Request.Context(), userID, c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bookmark"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Search Feedback ──

func (h *Extended2Handler) SubmitFeedback(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var req struct {
		Query    string `json:"query" binding:"required"`
		ResultID string `json:"result_id" binding:"required"`
		Rating   int    `json:"rating" binding:"required"`
		Comment  string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f := &service.SearchFeedback{
		UserID:   userID,
		Query:    req.Query,
		ResultID: req.ResultID,
		Rating:   req.Rating,
		Comment:  req.Comment,
	}
	if err := h.service.SubmitFeedback(c.Request.Context(), f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": f})
}

func (h *Extended2Handler) ListFeedback(c *gin.Context) {
	userID := getUserID(c)
	results, err := h.service.ListFeedback(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list feedback"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) GetFeedbackStats(c *gin.Context) {
	stats, err := h.service.GetFeedbackStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

// ── A/B Tests ──

func (h *Extended2Handler) CreateABTest(c *gin.Context) {
	var req struct {
		Name        string         `json:"name" binding:"required"`
		Description string         `json:"description"`
		ConfigA     map[string]any `json:"config_a"`
		ConfigB     map[string]any `json:"config_b"`
		SplitPct    int            `json:"split_pct"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t := &service.SearchABTest{
		Name:        req.Name,
		Description: req.Description,
		ConfigA:     req.ConfigA,
		ConfigB:     req.ConfigB,
		SplitPct:    req.SplitPct,
		IsActive:    true,
	}
	if err := h.service.CreateABTest(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create A/B test"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": t})
}

func (h *Extended2Handler) GetABTest(c *gin.Context) {
	t, err := h.service.GetABTest(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "A/B test not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": t})
}

func (h *Extended2Handler) ListABTests(c *gin.Context) {
	results, err := h.service.ListABTests(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list A/B tests"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) DeleteABTest(c *gin.Context) {
	if err := h.service.DeleteABTest(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete A/B test"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Search Pipelines ──

func (h *Extended2Handler) CreatePipeline(c *gin.Context) {
	var req struct {
		Name        string                `json:"name" binding:"required"`
		Description string                `json:"description"`
		Steps       []service.PipelineStep `json:"steps"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := &service.SearchPipeline{
		Name:        req.Name,
		Description: req.Description,
		Steps:       req.Steps,
		IsActive:    true,
	}
	if err := h.service.CreatePipeline(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pipeline"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": p})
}

func (h *Extended2Handler) GetPipeline(c *gin.Context) {
	p, err := h.service.GetPipeline(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": p})
}

func (h *Extended2Handler) ListPipelines(c *gin.Context) {
	results, err := h.service.ListPipelines(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list pipelines"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) UpdatePipeline(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdatePipeline(c.Request.Context(), c.Param("id"), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pipeline"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) DeletePipeline(c *gin.Context) {
	if err := h.service.DeletePipeline(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete pipeline"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Stop Words ──

func (h *Extended2Handler) AddStopWord(c *gin.Context) {
	var req struct {
		Word     string `json:"word" binding:"required"`
		Language string `json:"language"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	lang := req.Language
	if lang == "" { lang = "en" }
	sw := &service.StopWord{Word: req.Word, Language: lang}
	if err := h.service.AddStopWord(c.Request.Context(), sw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add stop word"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": sw})
}

func (h *Extended2Handler) ListStopWords(c *gin.Context) {
	lang := c.DefaultQuery("language", "en")
	results, err := h.service.ListStopWords(c.Request.Context(), lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list stop words"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) DeleteStopWord(c *gin.Context) {
	lang := c.DefaultQuery("language", "en")
	if err := h.service.DeleteStopWord(c.Request.Context(), lang, c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete stop word"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Query Rewrites ──

func (h *Extended2Handler) CreateRewrite(c *gin.Context) {
	var req struct {
		Pattern     string `json:"pattern" binding:"required"`
		Replacement string `json:"replacement" binding:"required"`
		Priority    int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r := &service.QueryRewrite{Pattern: req.Pattern, Replacement: req.Replacement, Priority: req.Priority}
	if err := h.service.CreateRewrite(c.Request.Context(), r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rewrite rule"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": r})
}

func (h *Extended2Handler) ListRewrites(c *gin.Context) {
	results, err := h.service.ListRewrites(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list rewrite rules"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) UpdateRewrite(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateRewrite(c.Request.Context(), c.Param("id"), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rewrite rule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) DeleteRewrite(c *gin.Context) {
	if err := h.service.DeleteRewrite(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rewrite rule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Index Schedules ──

func (h *Extended2Handler) CreateSchedule(c *gin.Context) {
	var req struct {
		IndexName string `json:"index_name" binding:"required"`
		Schedule  string `json:"schedule" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	is := &service.IndexSchedule{IndexName: req.IndexName, Schedule: req.Schedule}
	if err := h.service.CreateSchedule(c.Request.Context(), is); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": is})
}

func (h *Extended2Handler) ListSchedules(c *gin.Context) {
	results, err := h.service.ListSchedules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list schedules"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

func (h *Extended2Handler) UpdateSchedule(c *gin.Context) {
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateSchedule(c.Request.Context(), c.Param("id"), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update schedule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Extended2Handler) DeleteSchedule(c *gin.Context) {
	if err := h.service.DeleteSchedule(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete schedule"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
