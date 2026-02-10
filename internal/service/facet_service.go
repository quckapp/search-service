package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
)

type FacetService struct {
	es     *elasticsearch.Client
	redis  *redis.Client
	logger *logrus.Logger
}

func NewFacetService(es *elasticsearch.Client, redis *redis.Client, logger *logrus.Logger) *FacetService {
	return &FacetService{es: es, redis: redis, logger: logger}
}

func (s *FacetService) GetFacets(ctx context.Context, req *models.FacetRequest) (*models.FacetResult, error) {
	if s.es == nil {
		return &models.FacetResult{Field: req.Field, Buckets: []models.FacetBucket{}}, nil
	}

	size := req.Size
	if size <= 0 {
		size = 10
	}

	query := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"facet_agg": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": req.Field,
					"size":  size,
				},
			},
		},
	}

	if req.Query != "" {
		query["query"] = map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": req.Query,
			},
		}
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(query)

	res, err := s.es.Search(
		s.es.Search.WithIndex(req.Index),
		s.es.Search.WithBody(&buf),
	)
	if err != nil {
		return &models.FacetResult{Field: req.Field, Buckets: []models.FacetBucket{}}, nil
	}
	defer res.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)

	facetResult := &models.FacetResult{Field: req.Field, Buckets: []models.FacetBucket{}}
	if aggs, ok := result["aggregations"].(map[string]interface{}); ok {
		if fieldAgg, ok := aggs["facet_agg"].(map[string]interface{}); ok {
			if buckets, ok := fieldAgg["buckets"].([]interface{}); ok {
				for _, b := range buckets {
					bucket := b.(map[string]interface{})
					key := fmt.Sprintf("%v", bucket["key"])
					count := int64(0)
					if dc, ok := bucket["doc_count"].(float64); ok {
						count = int64(dc)
					}
					facetResult.Buckets = append(facetResult.Buckets, models.FacetBucket{
						Key: key, DocCount: count,
					})
				}
			}
		}
	}

	return facetResult, nil
}

func (s *FacetService) GetFacetedSearch(ctx context.Context, index, query, facetField string, size int) (*models.SearchResponse, []models.FacetResult, error) {
	if s.es == nil {
		return &models.SearchResponse{Results: []models.SearchHit{}}, []models.FacetResult{}, nil
	}

	if size <= 0 {
		size = 10
	}

	searchQuery := map[string]interface{}{
		"size": 20,
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": query,
			},
		},
		"aggs": map[string]interface{}{
			"facet_agg": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": facetField,
					"size":  size,
				},
			},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(searchQuery)

	res, err := s.es.Search(
		s.es.Search.WithIndex(index),
		s.es.Search.WithBody(&buf),
	)
	if err != nil {
		return &models.SearchResponse{Results: []models.SearchHit{}}, []models.FacetResult{}, nil
	}
	defer res.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)

	// Parse search results
	searchResp := &models.SearchResponse{Results: []models.SearchHit{}, Page: 1, PerPage: 20}
	if hits, ok := result["hits"].(map[string]interface{}); ok {
		if total, ok := hits["total"].(map[string]interface{}); ok {
			if val, ok := total["value"].(float64); ok {
				searchResp.Total = int64(val)
			}
		}
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
				searchResp.Results = append(searchResp.Results, searchHit)
			}
		}
	}

	// Parse facets
	var facets []models.FacetResult
	facetResult := models.FacetResult{Field: facetField, Buckets: []models.FacetBucket{}}
	if aggs, ok := result["aggregations"].(map[string]interface{}); ok {
		if fieldAgg, ok := aggs["facet_agg"].(map[string]interface{}); ok {
			if buckets, ok := fieldAgg["buckets"].([]interface{}); ok {
				for _, b := range buckets {
					bucket := b.(map[string]interface{})
					key := fmt.Sprintf("%v", bucket["key"])
					count := int64(0)
					if dc, ok := bucket["doc_count"].(float64); ok {
						count = int64(dc)
					}
					facetResult.Buckets = append(facetResult.Buckets, models.FacetBucket{
						Key: key, DocCount: count,
					})
				}
			}
		}
	}
	facets = append(facets, facetResult)

	return searchResp, facets, nil
}

func (s *FacetService) GetFilterOptions(ctx context.Context, workspaceID string) (*models.FilterOptions, error) {
	options := &models.FilterOptions{
		Types:      []string{"messages", "files", "users", "channels", "bookmarks", "tasks"},
		Channels:   []string{},
		Users:      []string{},
		DateRanges: []string{"today", "this_week", "this_month", "this_year"},
	}

	if s.es == nil {
		return options, nil
	}

	// Get unique channels
	channelQuery := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"term": map[string]interface{}{"workspace_id": workspaceID},
		},
		"aggs": map[string]interface{}{
			"channels": map[string]interface{}{
				"terms": map[string]interface{}{"field": "channel_id", "size": 50},
			},
		},
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(channelQuery)
	res, err := s.es.Search(
		s.es.Search.WithIndex("quckapp_messages"),
		s.es.Search.WithBody(&buf),
	)
	if err == nil {
		defer res.Body.Close()
		var result map[string]interface{}
		json.NewDecoder(res.Body).Decode(&result)
		if aggs, ok := result["aggregations"].(map[string]interface{}); ok {
			if chAgg, ok := aggs["channels"].(map[string]interface{}); ok {
				if buckets, ok := chAgg["buckets"].([]interface{}); ok {
					for _, b := range buckets {
						bucket := b.(map[string]interface{})
						options.Channels = append(options.Channels, fmt.Sprintf("%v", bucket["key"]))
					}
				}
			}
		}
	}

	return options, nil
}
