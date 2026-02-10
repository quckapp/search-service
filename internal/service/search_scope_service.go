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

type SearchScopeService struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func NewSearchScopeService(redis *redis.Client, logger *logrus.Logger) *SearchScopeService {
	return &SearchScopeService{redis: redis, logger: logger}
}

func (s *SearchScopeService) SetScope(ctx context.Context, userID string, req *models.SetSearchScopeRequest) (*models.SearchScope, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("storage not available")
	}

	scope := &models.SearchScope{
		ID:             uuid.New().String(),
		UserID:         userID,
		WorkspaceID:    req.WorkspaceID,
		AllowedIndices: req.AllowedIndices,
		DeniedChannels: req.DeniedChannels,
		CreatedAt:      time.Now(),
	}

	if scope.AllowedIndices == nil {
		scope.AllowedIndices = []string{}
	}
	if scope.DeniedChannels == nil {
		scope.DeniedChannels = []string{}
	}

	data, err := json.Marshal(scope)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("search_scope:%s:%s", userID, req.WorkspaceID)
	s.redis.Set(ctx, key, data, 0)

	return scope, nil
}

func (s *SearchScopeService) GetScope(ctx context.Context, userID, workspaceID string) (*models.SearchScope, error) {
	if s.redis == nil {
		return &models.SearchScope{
			UserID:         userID,
			WorkspaceID:    workspaceID,
			AllowedIndices: []string{},
			DeniedChannels: []string{},
		}, nil
	}

	key := fmt.Sprintf("search_scope:%s:%s", userID, workspaceID)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		// Return default scope with full access
		return &models.SearchScope{
			UserID:         userID,
			WorkspaceID:    workspaceID,
			AllowedIndices: []string{},
			DeniedChannels: []string{},
		}, nil
	}

	var scope models.SearchScope
	if err := json.Unmarshal(data, &scope); err != nil {
		return nil, err
	}
	return &scope, nil
}

func (s *SearchScopeService) ListAvailableScopes(ctx context.Context, workspaceID string) ([]string, error) {
	return []string{
		"quckapp_messages",
		"quckapp_files",
		"quckapp_users",
		"quckapp_channels",
		"quckapp_bookmarks",
		"quckapp_tasks",
		"quckapp_emoji",
	}, nil
}
