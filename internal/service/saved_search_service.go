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

type SavedSearchService struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func NewSavedSearchService(redis *redis.Client, logger *logrus.Logger) *SavedSearchService {
	return &SavedSearchService{redis: redis, logger: logger}
}

func (s *SavedSearchService) Create(ctx context.Context, userID string, req *models.CreateSavedSearchRequest) (*models.SavedSearch, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("storage not available")
	}

	saved := &models.SavedSearch{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		Query:       req.Query,
		SearchType:  req.SearchType,
		Filters:     req.Filters,
		WorkspaceID: req.WorkspaceID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	data, err := json.Marshal(saved)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("saved_search:%s:%s", userID, saved.ID)
	s.redis.Set(ctx, key, data, 0) // No expiry

	// Add to user's saved search list
	listKey := fmt.Sprintf("saved_searches:%s", userID)
	s.redis.SAdd(ctx, listKey, saved.ID)

	return saved, nil
}

func (s *SavedSearchService) GetByID(ctx context.Context, userID, searchID string) (*models.SavedSearch, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("storage not available")
	}

	key := fmt.Sprintf("saved_search:%s:%s", userID, searchID)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("saved search not found")
	}

	var saved models.SavedSearch
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, err
	}
	return &saved, nil
}

func (s *SavedSearchService) GetByUser(ctx context.Context, userID string) ([]models.SavedSearch, error) {
	if s.redis == nil {
		return []models.SavedSearch{}, nil
	}

	listKey := fmt.Sprintf("saved_searches:%s", userID)
	ids, err := s.redis.SMembers(ctx, listKey).Result()
	if err != nil {
		return []models.SavedSearch{}, nil
	}

	var searches []models.SavedSearch
	for _, id := range ids {
		key := fmt.Sprintf("saved_search:%s:%s", userID, id)
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		var saved models.SavedSearch
		if json.Unmarshal(data, &saved) == nil {
			searches = append(searches, saved)
		}
	}
	return searches, nil
}

func (s *SavedSearchService) Update(ctx context.Context, userID, searchID string, req *models.UpdateSavedSearchRequest) (*models.SavedSearch, error) {
	saved, err := s.GetByID(ctx, userID, searchID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		saved.Name = req.Name
	}
	if req.Query != "" {
		saved.Query = req.Query
	}
	if req.Filters != nil {
		saved.Filters = req.Filters
	}
	saved.UpdatedAt = time.Now()

	data, err := json.Marshal(saved)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("saved_search:%s:%s", userID, searchID)
	s.redis.Set(ctx, key, data, 0)

	return saved, nil
}

func (s *SavedSearchService) Delete(ctx context.Context, userID, searchID string) error {
	if s.redis == nil {
		return nil
	}

	key := fmt.Sprintf("saved_search:%s:%s", userID, searchID)
	s.redis.Del(ctx, key)

	listKey := fmt.Sprintf("saved_searches:%s", userID)
	s.redis.SRem(ctx, listKey, searchID)

	return nil
}
