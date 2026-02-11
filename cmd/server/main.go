package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/api"
	"github.com/quckapp/search-service/internal/config"
	"github.com/quckapp/search-service/internal/db"
	"github.com/quckapp/search-service/internal/handler"
	"github.com/quckapp/search-service/internal/service"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	cfg := config.Load()

	// Initialize Elasticsearch
	esClient := db.NewElasticsearch(cfg.ElasticsearchURL, logger)

	// Initialize Redis (optional)
	redisClient := db.NewRedis(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword)
	if redisClient != nil {
		logger.Info("Connected to Redis")
		defer redisClient.Close()
	}

	// -- Initialize Services --
	searchService := service.NewSearchService(esClient, redisClient, logger)
	historyService := service.NewHistoryService(redisClient, logger)
	savedSearchService := service.NewSavedSearchService(redisClient, logger)
	indexMgmtService := service.NewIndexManagementService(esClient, logger)
	extSearchService := service.NewExtendedSearchService(esClient, redisClient, logger)
	analyticsService := service.NewAnalyticsService(redisClient, logger)
	facetService := service.NewFacetService(esClient, redisClient, logger)
	synonymService := service.NewSynonymService(redisClient, logger)
	relevanceService := service.NewRelevanceService(esClient, redisClient, logger)
	alertService := service.NewAlertService(redisClient, logger)
	spellCheckService := service.NewSpellCheckService(esClient, logger)
	searchScopeService := service.NewSearchScopeService(redisClient, logger)
	extended2Service := service.NewExtended2Service(redisClient, logger)

	// -- Initialize Handlers --
	searchHandler := handler.NewSearchHandler(searchService, logger)
	historyHandler := handler.NewHistoryHandler(historyService, logger)
	savedSearchHandler := handler.NewSavedSearchHandler(savedSearchService, logger)
	indexMgmtHandler := handler.NewIndexManagementHandler(indexMgmtService, logger)
	extSearchHandler := handler.NewExtendedSearchHandler(extSearchService, logger)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService, logger)
	facetHandler := handler.NewFacetHandler(facetService, logger)
	synonymHandler := handler.NewSynonymHandler(synonymService, logger)
	relevanceHandler := handler.NewRelevanceHandler(relevanceService, logger)
	alertHandler := handler.NewAlertHandler(alertService, logger)
	spellCheckHandler := handler.NewSpellCheckHandler(spellCheckService, logger)
	searchScopeHandler := handler.NewSearchScopeHandler(searchScopeService, logger)
	ext2Handler := handler.NewExtended2Handler(extended2Service, logger)

	// Setup router
	router := api.NewRouter(
		searchHandler,
		historyHandler,
		savedSearchHandler,
		indexMgmtHandler,
		extSearchHandler,
		analyticsHandler,
		facetHandler,
		synonymHandler,
		relevanceHandler,
		alertHandler,
		spellCheckHandler,
		searchScopeHandler,
		ext2Handler,
		cfg,
		logger,
	)

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Search service starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Search service stopped")
}
