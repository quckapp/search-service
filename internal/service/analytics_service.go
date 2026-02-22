package service

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
)

type AnalyticsService struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func NewAnalyticsService(redis *redis.Client, logger *logrus.Logger) *AnalyticsService {
	return &AnalyticsService{redis: redis, logger: logger}
}

func (s *AnalyticsService) GetAnalytics(ctx context.Context, workspaceID string) (*models.SearchAnalytics, error) {
	if s.redis == nil {
		return &models.SearchAnalytics{
			TopQueries:     []models.QueryCount{},
			SearchesByType: map[string]int64{},
		}, nil
	}

	analytics := &models.SearchAnalytics{
		TopQueries:     []models.QueryCount{},
		SearchesByType: map[string]int64{},
	}

	// Top queries from sorted set
	queriesKey := fmt.Sprintf("search_analytics:%s:queries", workspaceID)
	topQueries, err := s.redis.ZRevRangeWithScores(ctx, queriesKey, 0, 9).Result()
	if err == nil {
		for _, q := range topQueries {
			analytics.TopQueries = append(analytics.TopQueries, models.QueryCount{
				Query: q.Member.(string),
				Count: int64(q.Score),
			})
			analytics.TotalSearches += int64(q.Score)
		}
	}

	// Searches by type
	for _, t := range []string{"global", "messages", "files", "users", "channels", "bookmarks", "tasks"} {
		typeKey := fmt.Sprintf("search_analytics:%s:type:%s", workspaceID, t)
		count, err := s.redis.Get(ctx, typeKey).Int64()
		if err == nil {
			analytics.SearchesByType[t] = count
		}
	}

	return analytics, nil
}

func (s *AnalyticsService) RecordSearchType(ctx context.Context, workspaceID, searchType string) {
	if s.redis == nil {
		return
	}
	typeKey := fmt.Sprintf("search_analytics:%s:type:%s", workspaceID, searchType)
	s.redis.Incr(ctx, typeKey)
}

func (s *AnalyticsService) GetPopularQueries(ctx context.Context, workspaceID string, limit int64) ([]models.QueryCount, error) {
	if s.redis == nil {
		return []models.QueryCount{}, nil
	}

	queriesKey := fmt.Sprintf("search_analytics:%s:queries", workspaceID)
	results, err := s.redis.ZRevRangeWithScores(ctx, queriesKey, 0, limit-1).Result()
	if err != nil {
		return []models.QueryCount{}, nil
	}

	var queries []models.QueryCount
	for _, r := range results {
		queries = append(queries, models.QueryCount{
			Query: r.Member.(string),
			Count: int64(r.Score),
		})
	}
	return queries, nil
}

func (s *AnalyticsService) ClearAnalytics(ctx context.Context, workspaceID string) error {
	if s.redis == nil {
		return nil
	}

	pattern := fmt.Sprintf("search_analytics:%s:*", workspaceID)
	keys, _ := s.redis.Keys(ctx, pattern).Result()
	if len(keys) > 0 {
		s.redis.Del(ctx, keys...)
	}
	return nil
}
