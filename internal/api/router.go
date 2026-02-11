package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/config"
	"github.com/quckapp/search-service/internal/handler"
	"github.com/quckapp/search-service/internal/middleware"
)

func NewRouter(
	searchHandler *handler.SearchHandler,
	historyHandler *handler.HistoryHandler,
	savedSearchHandler *handler.SavedSearchHandler,
	indexMgmtHandler *handler.IndexManagementHandler,
	extSearchHandler *handler.ExtendedSearchHandler,
	analyticsHandler *handler.AnalyticsHandler,
	facetHandler *handler.FacetHandler,
	synonymHandler *handler.SynonymHandler,
	relevanceHandler *handler.RelevanceHandler,
	alertHandler *handler.AlertHandler,
	spellCheckHandler *handler.SpellCheckHandler,
	searchScopeHandler *handler.SearchScopeHandler,
	ext2Handler *handler.Extended2Handler,
	cfg *config.Config,
	logger *logrus.Logger,
) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	// Health (no auth)
	r.GET("/health", searchHandler.Health)

	api := r.Group("/api/v1")
	api.Use(middleware.Auth(cfg.JWTSecret))
	{
		// -- Core Search --
		api.GET("/search", searchHandler.GlobalSearch)
		api.GET("/search/messages", searchHandler.SearchMessages)
		api.GET("/search/files", searchHandler.SearchFiles)
		api.GET("/search/users", searchHandler.SearchUsers)
		api.GET("/search/channels", searchHandler.SearchChannels)
		api.GET("/search/suggest", searchHandler.Suggest)

		// -- Extended Search --
		api.GET("/search/bookmarks", extSearchHandler.SearchBookmarks)
		api.GET("/search/tasks", extSearchHandler.SearchTasks)
		api.GET("/search/emoji", extSearchHandler.SearchEmoji)
		api.POST("/search/advanced", extSearchHandler.AdvancedSearch)

		// -- Search History --
		api.GET("/search/history", historyHandler.GetHistory)
		api.DELETE("/search/history", historyHandler.ClearHistory)
		api.DELETE("/search/history/:id", historyHandler.DeleteHistoryItem)

		// -- Saved Searches --
		api.POST("/search/saved", savedSearchHandler.Create)
		api.GET("/search/saved", savedSearchHandler.GetByUser)
		api.GET("/search/saved/:id", savedSearchHandler.GetByID)
		api.PUT("/search/saved/:id", savedSearchHandler.Update)
		api.DELETE("/search/saved/:id", savedSearchHandler.Delete)

		// -- Analytics --
		api.GET("/analytics", analyticsHandler.GetAnalytics)
		api.GET("/analytics/popular-queries", analyticsHandler.GetPopularQueries)
		api.DELETE("/analytics", analyticsHandler.ClearAnalytics)

		// -- Aggregation --
		api.POST("/search/aggregate", extSearchHandler.Aggregate)

		// -- Core Index Management --
		api.POST("/index", searchHandler.IndexDocument)
		api.POST("/index/message", searchHandler.IndexMessage)
		api.POST("/index/file", searchHandler.IndexFile)
		api.POST("/index/bulk", searchHandler.BulkIndex)
		api.DELETE("/index/:type/:id", searchHandler.DeleteFromIndex)
		api.POST("/index/reindex", searchHandler.Reindex)

		// -- Typed Index Endpoints --
		api.POST("/index/user", extSearchHandler.IndexUser)
		api.POST("/index/channel", extSearchHandler.IndexChannel)
		api.POST("/index/bookmark", extSearchHandler.IndexBookmark)
		api.POST("/index/task", extSearchHandler.IndexTask)

		// -- Batch Operations --
		api.POST("/index/batch-delete", extSearchHandler.BatchDelete)
		api.PUT("/index/:index/:id", extSearchHandler.UpdateDocument)

		// -- Document Count --
		api.GET("/index/:index/count", extSearchHandler.CountDocuments)

		// -- Index Administration --
		api.GET("/indices", indexMgmtHandler.ListIndices)
		api.GET("/indices/:index", indexMgmtHandler.GetIndexInfo)
		api.POST("/indices", indexMgmtHandler.CreateIndex)
		api.DELETE("/indices/:index", indexMgmtHandler.DeleteIndex)
		api.PUT("/indices/mappings", indexMgmtHandler.PutMapping)
		api.PUT("/indices/settings", indexMgmtHandler.UpdateSettings)
		api.POST("/indices/aliases", indexMgmtHandler.CreateAlias)
		api.DELETE("/indices/aliases", indexMgmtHandler.DeleteAlias)
		api.POST("/indices/:index/refresh", indexMgmtHandler.RefreshIndex)
		api.POST("/indices/:index/flush", indexMgmtHandler.FlushIndex)

		// -- Search Facets/Filters --
		api.POST("/search/facets", facetHandler.GetFacets)
		api.GET("/search/faceted", facetHandler.GetFacetedSearch)
		api.GET("/search/filters", facetHandler.GetFilterOptions)

		// -- Synonym Management --
		api.POST("/synonyms", synonymHandler.Create)
		api.GET("/synonyms", synonymHandler.List)
		api.PUT("/synonyms/:id", synonymHandler.Update)
		api.DELETE("/synonyms/:id", synonymHandler.Delete)
		api.POST("/synonyms/apply", synonymHandler.ApplyToIndex)

		// -- Relevance Tuning --
		api.GET("/relevance", relevanceHandler.GetConfig)
		api.PUT("/relevance", relevanceHandler.UpdateConfig)
		api.GET("/relevance/preview", relevanceHandler.PreviewTuning)
		api.POST("/relevance/reset", relevanceHandler.ResetToDefaults)

		// -- Search Alerts --
		api.POST("/alerts", alertHandler.Create)
		api.GET("/alerts", alertHandler.List)
		api.PUT("/alerts/:id", alertHandler.Update)
		api.DELETE("/alerts/:id", alertHandler.Delete)
		api.GET("/alerts/:id/history", alertHandler.GetHistory)

		// -- Spell Check / Did You Mean --
		api.GET("/search/spellcheck", spellCheckHandler.GetSuggestions)
		api.GET("/search/didyoumean", spellCheckHandler.DidYouMean)

		// -- Search Permissions/Scoping --
		api.POST("/search/scope", searchScopeHandler.SetScope)
		api.GET("/search/scope", searchScopeHandler.GetScope)
		api.GET("/search/scopes", searchScopeHandler.ListAvailableScopes)

		// -- Search Templates --
		api.POST("/search/templates", ext2Handler.CreateTemplate)
		api.GET("/search/templates", ext2Handler.ListTemplates)
		api.GET("/search/templates/:id", ext2Handler.GetTemplate)
		api.PUT("/search/templates/:id", ext2Handler.UpdateTemplate)
		api.DELETE("/search/templates/:id", ext2Handler.DeleteTemplate)
		api.POST("/search/templates/:id/use", ext2Handler.UseTemplate)

		// -- Search Result Bookmarks --
		api.POST("/search/result-bookmarks", ext2Handler.BookmarkResult)
		api.GET("/search/result-bookmarks", ext2Handler.ListBookmarks)
		api.DELETE("/search/result-bookmarks/:id", ext2Handler.DeleteBookmark)

		// -- Search Feedback --
		api.POST("/search/feedback", ext2Handler.SubmitFeedback)
		api.GET("/search/feedback", ext2Handler.ListFeedback)
		api.GET("/search/feedback/stats", ext2Handler.GetFeedbackStats)

		// -- A/B Tests --
		api.POST("/search/ab-tests", ext2Handler.CreateABTest)
		api.GET("/search/ab-tests", ext2Handler.ListABTests)
		api.GET("/search/ab-tests/:id", ext2Handler.GetABTest)
		api.DELETE("/search/ab-tests/:id", ext2Handler.DeleteABTest)

		// -- Search Pipelines --
		api.POST("/search/pipelines", ext2Handler.CreatePipeline)
		api.GET("/search/pipelines", ext2Handler.ListPipelines)
		api.GET("/search/pipelines/:id", ext2Handler.GetPipeline)
		api.PUT("/search/pipelines/:id", ext2Handler.UpdatePipeline)
		api.DELETE("/search/pipelines/:id", ext2Handler.DeletePipeline)

		// -- Stop Words --
		api.POST("/stop-words", ext2Handler.AddStopWord)
		api.GET("/stop-words", ext2Handler.ListStopWords)
		api.DELETE("/stop-words/:id", ext2Handler.DeleteStopWord)

		// -- Query Rewrites --
		api.POST("/query-rewrites", ext2Handler.CreateRewrite)
		api.GET("/query-rewrites", ext2Handler.ListRewrites)
		api.PUT("/query-rewrites/:id", ext2Handler.UpdateRewrite)
		api.DELETE("/query-rewrites/:id", ext2Handler.DeleteRewrite)

		// -- Index Schedules --
		api.POST("/index-schedules", ext2Handler.CreateSchedule)
		api.GET("/index-schedules", ext2Handler.ListSchedules)
		api.PUT("/index-schedules/:id", ext2Handler.UpdateSchedule)
		api.DELETE("/index-schedules/:id", ext2Handler.DeleteSchedule)
	}

	return r
}
