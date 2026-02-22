package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
)

type HistoryService struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func NewHistoryService(redis *redis.Client, logger *logrus.Logger) *HistoryService {
	return &HistoryService{redis: redis, logger: logger}
}

func (s *HistoryService) RecordSearch(ctx context.Context, userID, query, searchType, workspaceID string, resultCount int64) error {
	if s.redis == nil {
		return nil
	}

	history := models.SearchHistory{
		ID:          uuid.New().String(),
		UserID:      userID,
		Query:       query,
		SearchType:  searchType,
		WorkspaceID: workspaceID,
		ResultCount: resultCount,
		CreatedAt:   time.Now(),
	}

	data, err := json.Marshal(history)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("search_history:%s", userID)
	s.redis.LPush(ctx, key, data)
	s.redis.LTrim(ctx, key, 0, 99) // Keep last 100

	// Track query frequency for analytics
	analyticsKey := fmt.Sprintf("search_analytics:%s:queries", workspaceID)
	s.redis.ZIncrBy(ctx, analyticsKey, 1, query)

	return nil
}

func (s *HistoryService) GetHistory(ctx context.Context, userID string, limit int64) ([]models.SearchHistory, error) {
	if s.redis == nil {
		return []models.SearchHistory{}, nil
	}

	key := fmt.Sprintf("search_history:%s", userID)
	results, err := s.redis.LRange(ctx, key, 0, limit-1).Result()
	if err != nil {
		return []models.SearchHistory{}, nil
	}

	var history []models.SearchHistory
	for _, r := range results {
		var h models.SearchHistory
		if json.Unmarshal([]byte(r), &h) == nil {
			history = append(history, h)
		}
	}
	return history, nil
}

func (s *HistoryService) ClearHistory(ctx context.Context, userID string) error {
	if s.redis == nil {
		return nil
	}
	key := fmt.Sprintf("search_history:%s", userID)
	return s.redis.Del(ctx, key).Err()
}

func (s *HistoryService) DeleteHistoryItem(ctx context.Context, userID, historyID string) error {
	if s.redis == nil {
		return nil
	}

	key := fmt.Sprintf("search_history:%s", userID)
	results, err := s.redis.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, r := range results {
		var h models.SearchHistory
		if json.Unmarshal([]byte(r), &h) == nil && h.ID == historyID {
			s.redis.LRem(ctx, key, 1, r)
			return nil
		}
	}
	return nil
}
