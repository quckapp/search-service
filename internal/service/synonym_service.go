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

type SynonymService struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func NewSynonymService(redis *redis.Client, logger *logrus.Logger) *SynonymService {
	return &SynonymService{redis: redis, logger: logger}
}

func (s *SynonymService) Create(ctx context.Context, userID string, req *models.CreateSynonymRequest) (*models.SynonymGroup, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("storage not available")
	}

	group := &models.SynonymGroup{
		ID:          uuid.New().String(),
		WorkspaceID: req.WorkspaceID,
		Terms:       req.Terms,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	data, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("synonym:%s:%s", req.WorkspaceID, group.ID)
	s.redis.Set(ctx, key, data, 0)

	listKey := fmt.Sprintf("synonyms:%s", req.WorkspaceID)
	s.redis.SAdd(ctx, listKey, group.ID)

	return group, nil
}

func (s *SynonymService) List(ctx context.Context, workspaceID string) ([]models.SynonymGroup, error) {
	if s.redis == nil {
		return []models.SynonymGroup{}, nil
	}

	listKey := fmt.Sprintf("synonyms:%s", workspaceID)
	ids, err := s.redis.SMembers(ctx, listKey).Result()
	if err != nil {
		return []models.SynonymGroup{}, nil
	}

	var groups []models.SynonymGroup
	for _, id := range ids {
		key := fmt.Sprintf("synonym:%s:%s", workspaceID, id)
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		var group models.SynonymGroup
		if json.Unmarshal(data, &group) == nil {
			groups = append(groups, group)
		}
	}
	return groups, nil
}

func (s *SynonymService) Update(ctx context.Context, workspaceID, synonymID string, req *models.UpdateSynonymRequest) (*models.SynonymGroup, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("storage not available")
	}

	key := fmt.Sprintf("synonym:%s:%s", workspaceID, synonymID)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("synonym group not found")
	}

	var group models.SynonymGroup
	if err := json.Unmarshal(data, &group); err != nil {
		return nil, err
	}

	group.Terms = req.Terms
	group.UpdatedAt = time.Now()

	updated, err := json.Marshal(&group)
	if err != nil {
		return nil, err
	}
	s.redis.Set(ctx, key, updated, 0)

	return &group, nil
}

func (s *SynonymService) Delete(ctx context.Context, workspaceID, synonymID string) error {
	if s.redis == nil {
		return nil
	}

	key := fmt.Sprintf("synonym:%s:%s", workspaceID, synonymID)
	s.redis.Del(ctx, key)

	listKey := fmt.Sprintf("synonyms:%s", workspaceID)
	s.redis.SRem(ctx, listKey, synonymID)

	return nil
}

func (s *SynonymService) ApplyToIndex(ctx context.Context, workspaceID string) (int, error) {
	groups, err := s.List(ctx, workspaceID)
	if err != nil {
		return 0, err
	}
	// Returns count of synonym groups that would be applied
	return len(groups), nil
}
