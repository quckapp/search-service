package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
)

type RelevanceService struct {
	es     *elasticsearch.Client
	redis  *redis.Client
	logger *logrus.Logger
}

func NewRelevanceService(es *elasticsearch.Client, redis *redis.Client, logger *logrus.Logger) *RelevanceService {
	return &RelevanceService{es: es, redis: redis, logger: logger}
}

func (s *RelevanceService) defaultConfig(workspaceID string) *models.RelevanceConfig {
	return &models.RelevanceConfig{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		FieldBoosts: map[string]float64{
			"title":        3.0,
			"content":      1.0,
			"name":         2.0,
			"description":  1.5,
			"display_name": 2.0,
		},
		TitleBoost:      3.0,
		ContentBoost:    1.0,
		RecencyWeight:   0.5,
		ExactMatchBoost: 2.0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func (s *RelevanceService) GetConfig(ctx context.Context, workspaceID string) (*models.RelevanceConfig, error) {
	if s.redis == nil {
		return s.defaultConfig(workspaceID), nil
	}

	key := fmt.Sprintf("relevance_config:%s", workspaceID)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return s.defaultConfig(workspaceID), nil
	}

	var config models.RelevanceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return s.defaultConfig(workspaceID), nil
	}
	return &config, nil
}

func (s *RelevanceService) UpdateConfig(ctx context.Context, workspaceID string, req *models.UpdateRelevanceRequest) (*models.RelevanceConfig, error) {
	config, _ := s.GetConfig(ctx, workspaceID)

	if req.FieldBoosts != nil {
		config.FieldBoosts = req.FieldBoosts
	}
	if req.TitleBoost != nil {
		config.TitleBoost = *req.TitleBoost
	}
	if req.ContentBoost != nil {
		config.ContentBoost = *req.ContentBoost
	}
	if req.RecencyWeight != nil {
		config.RecencyWeight = *req.RecencyWeight
	}
	if req.ExactMatchBoost != nil {
		config.ExactMatchBoost = *req.ExactMatchBoost
	}
	config.UpdatedAt = time.Now()

	if s.redis != nil {
		data, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf("relevance_config:%s", workspaceID)
		s.redis.Set(ctx, key, data, 0)
	}

	return config, nil
}

func (s *RelevanceService) PreviewTuning(ctx context.Context, workspaceID, query, index string) (*models.RelevancePreview, error) {
	config, _ := s.GetConfig(ctx, workspaceID)

	preview := &models.RelevancePreview{
		Query:   query,
		Results: []models.SearchHit{},
		Config:  config,
	}

	if s.es == nil || query == "" {
		return preview, nil
	}

	// Build boosted fields from config
	var fields []interface{}
	for field, boost := range config.FieldBoosts {
		fields = append(fields, fmt.Sprintf("%s^%.1f", field, boost))
	}

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"multi_match": map[string]interface{}{
						"query":     query,
						"fields":    fields,
						"fuzziness": "AUTO",
					}},
				},
				"filter": []map[string]interface{}{
					{"term": map[string]interface{}{"workspace_id": workspaceID}},
				},
			},
		},
		"size": 10,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(searchQuery)

	if index == "" {
		index = "quckapp_messages"
	}

	res, err := s.es.Search(
		s.es.Search.WithIndex(index),
		s.es.Search.WithBody(&buf),
	)
	if err != nil {
		return preview, nil
	}
	defer res.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)

	if hits, ok := result["hits"].(map[string]interface{}); ok {
		if hitList, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitList {
				hitMap := hit.(map[string]interface{})
				searchHit := models.SearchHit{ID: hitMap["_id"].(string)}
				if idx, ok := hitMap["_index"].(string); ok {
					searchHit.Index = idx
				}
				if score, ok := hitMap["_score"].(float64); ok {
					searchHit.Score = score
				}
				if source, ok := hitMap["_source"].(map[string]interface{}); ok {
					searchHit.Source = source
				}
				preview.Results = append(preview.Results, searchHit)
			}
		}
	}

	return preview, nil
}

func (s *RelevanceService) ResetToDefaults(ctx context.Context, workspaceID string) (*models.RelevanceConfig, error) {
	config := s.defaultConfig(workspaceID)

	if s.redis != nil {
		data, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf("relevance_config:%s", workspaceID)
		s.redis.Set(ctx, key, data, 0)
	}

	return config, nil
}
