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

type AlertService struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func NewAlertService(redis *redis.Client, logger *logrus.Logger) *AlertService {
	return &AlertService{redis: redis, logger: logger}
}

func (s *AlertService) Create(ctx context.Context, userID string, req *models.CreateAlertRequest) (*models.SearchAlert, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("storage not available")
	}

	alert := &models.SearchAlert{
		ID:          uuid.New().String(),
		UserID:      userID,
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		Query:       req.Query,
		SearchType:  req.SearchType,
		Frequency:   req.Frequency,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	data, err := json.Marshal(alert)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("search_alert:%s:%s", userID, alert.ID)
	s.redis.Set(ctx, key, data, 0)

	listKey := fmt.Sprintf("search_alerts:%s", userID)
	s.redis.SAdd(ctx, listKey, alert.ID)

	return alert, nil
}

func (s *AlertService) List(ctx context.Context, userID string) ([]models.SearchAlert, error) {
	if s.redis == nil {
		return []models.SearchAlert{}, nil
	}

	listKey := fmt.Sprintf("search_alerts:%s", userID)
	ids, err := s.redis.SMembers(ctx, listKey).Result()
	if err != nil {
		return []models.SearchAlert{}, nil
	}

	var alerts []models.SearchAlert
	for _, id := range ids {
		key := fmt.Sprintf("search_alert:%s:%s", userID, id)
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		var alert models.SearchAlert
		if json.Unmarshal(data, &alert) == nil {
			alerts = append(alerts, alert)
		}
	}
	return alerts, nil
}

func (s *AlertService) Update(ctx context.Context, userID, alertID string, req *models.UpdateAlertRequest) (*models.SearchAlert, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("storage not available")
	}

	key := fmt.Sprintf("search_alert:%s:%s", userID, alertID)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("alert not found")
	}

	var alert models.SearchAlert
	if err := json.Unmarshal(data, &alert); err != nil {
		return nil, err
	}

	if req.Name != "" {
		alert.Name = req.Name
	}
	if req.Query != "" {
		alert.Query = req.Query
	}
	if req.Frequency != "" {
		alert.Frequency = req.Frequency
	}
	if req.IsActive != nil {
		alert.IsActive = *req.IsActive
	}

	updated, err := json.Marshal(&alert)
	if err != nil {
		return nil, err
	}
	s.redis.Set(ctx, key, updated, 0)

	return &alert, nil
}

func (s *AlertService) Delete(ctx context.Context, userID, alertID string) error {
	if s.redis == nil {
		return nil
	}

	key := fmt.Sprintf("search_alert:%s:%s", userID, alertID)
	s.redis.Del(ctx, key)

	listKey := fmt.Sprintf("search_alerts:%s", userID)
	s.redis.SRem(ctx, listKey, alertID)

	// Also clean up history
	historyKey := fmt.Sprintf("alert_history:%s", alertID)
	s.redis.Del(ctx, historyKey)

	return nil
}

func (s *AlertService) GetHistory(ctx context.Context, alertID string) ([]models.AlertHistory, error) {
	if s.redis == nil {
		return []models.AlertHistory{}, nil
	}

	historyKey := fmt.Sprintf("alert_history:%s", alertID)
	results, err := s.redis.LRange(ctx, historyKey, 0, 49).Result()
	if err != nil {
		return []models.AlertHistory{}, nil
	}

	var history []models.AlertHistory
	for _, r := range results {
		var h models.AlertHistory
		if json.Unmarshal([]byte(r), &h) == nil {
			history = append(history, h)
		}
	}
	return history, nil
}
