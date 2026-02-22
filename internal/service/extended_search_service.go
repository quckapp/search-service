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

type ExtendedSearchService struct {
	es     *elasticsearch.Client
	redis  *redis.Client
	logger *logrus.Logger
}

func NewExtendedSearchService(es *elasticsearch.Client, redis *redis.Client, logger *logrus.Logger) *ExtendedSearchService {
	return &ExtendedSearchService{es: es, redis: redis, logger: logger}
}

// ── Bookmark Search ──

func (s *ExtendedSearchService) SearchBookmarks(ctx context.Context, params *models.SearchParams) (*models.SearchResponse, error) {
	params.Validate()

	must := []map[string]interface{}{
		{"multi_match": map[string]interface{}{
			"query":     params.Query,
			"fields":    []string{"title^3", "description", "url", "tags"},
			"fuzziness": "AUTO",
		}},
	}

	filters := buildExtFilters(params)
	query := buildExtQuery(must, filters, params)

	result, err := s.executeSearch("quckapp_bookmarks", query)
	if err != nil {
		return extEmptyResponse(params), nil
	}

	return s.parseResponse(result, params), nil
}

// ── Task Search ──

func (s *ExtendedSearchService) SearchTasks(ctx context.Context, params *models.SearchParams) (*models.SearchResponse, error) {
	params.Validate()

	must := []map[string]interface{}{
		{"multi_match": map[string]interface{}{
			"query":     params.Query,
			"fields":    []string{"title^3", "description"},
			"fuzziness": "AUTO",
		}},
	}

	filters := buildExtFilters(params)
	query := buildExtQuery(must, filters, params)

	result, err := s.executeSearch("quckapp_tasks", query)
	if err != nil {
		return extEmptyResponse(params), nil
	}

	return s.parseResponse(result, params), nil
}

// ── Emoji Search ──

func (s *ExtendedSearchService) SearchEmoji(ctx context.Context, query, workspaceID string) (*models.SearchResponse, error) {
	params := &models.SearchParams{Query: query, WorkspaceID: workspaceID, Page: 1, PerPage: 50}
	params.Validate()

	must := []map[string]interface{}{
		{"multi_match": map[string]interface{}{
			"query":  query,
			"fields": []string{"name^2", "category"},
			"type":   "phrase_prefix",
		}},
	}

	var filters []map[string]interface{}
	if workspaceID != "" {
		filters = append(filters, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"workspace_id": workspaceID}},
					{"term": map[string]interface{}{"is_custom": false}},
				},
			},
		})
	}

	searchQuery := buildExtQuery(must, filters, params)
	result, err := s.executeSearch("quckapp_emoji", searchQuery)
	if err != nil {
		return extEmptyResponse(params), nil
	}

	return s.parseResponse(result, params), nil
}

// ── Advanced Search ──

func (s *ExtendedSearchService) AdvancedSearch(ctx context.Context, req *models.AdvancedSearchParams) (*models.GlobalSearchResponse, error) {
	resp := &models.GlobalSearchResponse{}

	for _, sub := range req.Queries {
		params := &models.SearchParams{
			Query:       sub.Query,
			WorkspaceID: req.WorkspaceID,
			Page:        req.Page,
			PerPage:     req.PerPage,
		}
		params.Validate()

		fields := sub.Fields
		if len(fields) == 0 {
			fields = []string{"*"}
		}

		must := []map[string]interface{}{
			{"multi_match": map[string]interface{}{
				"query":  sub.Query,
				"fields": fields,
			}},
		}

		var filters []map[string]interface{}
		if req.WorkspaceID != "" {
			filters = append(filters, map[string]interface{}{
				"term": map[string]interface{}{"workspace_id": req.WorkspaceID},
			})
		}

		query := buildExtQuery(must, filters, params)
		result, _ := s.executeSearch(sub.Index, query)
		parsed := s.parseResponse(result, params)

		switch sub.Index {
		case "quckapp_messages":
			resp.Messages = parsed
		case "quckapp_files":
			resp.Files = parsed
		case "quckapp_users":
			resp.Users = parsed
		case "quckapp_channels":
			resp.Channels = parsed
		}
	}

	return resp, nil
}

// ── Aggregation ──

func (s *ExtendedSearchService) Aggregate(ctx context.Context, req *models.AggregationRequest) (*models.AggregationResponse, error) {
	if s.es == nil {
		return &models.AggregationResponse{Buckets: []models.AggregationBucket{}}, nil
	}

	size := req.Size
	if size <= 0 {
		size = 10
	}

	query := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"field_agg": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": req.Field,
					"size":  size,
				},
			},
		},
	}

	result, err := s.executeSearch(req.Index, query)
	if err != nil {
		return &models.AggregationResponse{Buckets: []models.AggregationBucket{}}, nil
	}

	resp := &models.AggregationResponse{Buckets: []models.AggregationBucket{}}
	if aggs, ok := result["aggregations"].(map[string]interface{}); ok {
		if fieldAgg, ok := aggs["field_agg"].(map[string]interface{}); ok {
			if buckets, ok := fieldAgg["buckets"].([]interface{}); ok {
				for _, b := range buckets {
					bucket := b.(map[string]interface{})
					key := fmt.Sprintf("%v", bucket["key"])
					count := int64(0)
					if dc, ok := bucket["doc_count"].(float64); ok {
						count = int64(dc)
					}
					resp.Buckets = append(resp.Buckets, models.AggregationBucket{
						Key: key, DocCount: count,
					})
				}
			}
		}
	}

	return resp, nil
}

// ── Batch Delete ──

func (s *ExtendedSearchService) BatchDelete(ctx context.Context, req *models.BatchDeleteRequest) *models.BatchDeleteResponse {
	resp := &models.BatchDeleteResponse{}

	if s.es == nil {
		return resp
	}

	for _, id := range req.IDs {
		_, err := s.es.Delete(req.Index, id)
		if err != nil {
			resp.Failed++
			resp.Errors = append(resp.Errors, fmt.Sprintf("%s: %s", id, err.Error()))
		} else {
			resp.Deleted++
		}
	}
	return resp
}

// ── Update Document ──

func (s *ExtendedSearchService) UpdateDocument(ctx context.Context, index, id string, doc map[string]interface{}) error {
	if s.es == nil {
		return nil
	}

	body := map[string]interface{}{
		"doc": doc,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)

	res, err := s.es.Update(index, id, &buf)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

// ── Index Typed Documents ──

func (s *ExtendedSearchService) IndexUser(ctx context.Context, req *models.IndexUserRequest) error {
	if s.es == nil {
		return nil
	}
	doc := map[string]interface{}{
		"username":     req.Username,
		"display_name": req.DisplayName,
		"email":        req.Email,
		"avatar_url":   req.AvatarURL,
		"workspace_id": req.WorkspaceID,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(doc)
	_, err := s.es.Index("quckapp_users", &buf, s.es.Index.WithDocumentID(req.ID))
	return err
}

func (s *ExtendedSearchService) IndexChannel(ctx context.Context, req *models.IndexChannelRequest) error {
	if s.es == nil {
		return nil
	}
	doc := map[string]interface{}{
		"name":         req.Name,
		"description":  req.Description,
		"topic":        req.Topic,
		"type":         req.Type,
		"workspace_id": req.WorkspaceID,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(doc)
	_, err := s.es.Index("quckapp_channels", &buf, s.es.Index.WithDocumentID(req.ID))
	return err
}

func (s *ExtendedSearchService) IndexBookmark(ctx context.Context, req *models.IndexBookmarkRequest) error {
	if s.es == nil {
		return nil
	}
	doc := map[string]interface{}{
		"title":        req.Title,
		"description":  req.Description,
		"url":          req.URL,
		"tags":         req.Tags,
		"user_id":      req.UserID,
		"workspace_id": req.WorkspaceID,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(doc)
	_, err := s.es.Index("quckapp_bookmarks", &buf, s.es.Index.WithDocumentID(req.ID))
	return err
}

func (s *ExtendedSearchService) IndexTask(ctx context.Context, req *models.IndexTaskRequest) error {
	if s.es == nil {
		return nil
	}
	doc := map[string]interface{}{
		"title":        req.Title,
		"description":  req.Description,
		"status":       req.Status,
		"priority":     req.Priority,
		"assignee_id":  req.AssigneeID,
		"user_id":      req.UserID,
		"workspace_id": req.WorkspaceID,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(doc)
	_, err := s.es.Index("quckapp_tasks", &buf, s.es.Index.WithDocumentID(req.ID))
	return err
}

// ── Document Count ──

func (s *ExtendedSearchService) CountDocuments(ctx context.Context, index string) (int64, error) {
	if s.es == nil {
		return 0, nil
	}

	res, err := s.es.Count(s.es.Count.WithIndex(index))
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)

	if count, ok := result["count"].(float64); ok {
		return int64(count), nil
	}
	return 0, nil
}

// ── Helpers ──

func (s *ExtendedSearchService) executeSearch(index string, query map[string]interface{}) (map[string]interface{}, error) {
	if s.es == nil {
		return map[string]interface{}{
			"hits": map[string]interface{}{
				"total": map[string]interface{}{"value": 0},
				"hits":  []interface{}{},
			},
		}, nil
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(query)

	res, err := s.es.Search(
		s.es.Search.WithIndex(index),
		s.es.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)
	return result, nil
}

func (s *ExtendedSearchService) parseResponse(result map[string]interface{}, params *models.SearchParams) *models.SearchResponse {
	resp := &models.SearchResponse{
		Results: []models.SearchHit{},
		Page:    params.Page,
		PerPage: params.PerPage,
	}

	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return resp
	}

	if total, ok := hits["total"].(map[string]interface{}); ok {
		if val, ok := total["value"].(float64); ok {
			resp.Total = int64(val)
		}
	}

	if hitList, ok := hits["hits"].([]interface{}); ok {
		for _, hit := range hitList {
			hitMap := hit.(map[string]interface{})
			searchHit := models.SearchHit{
				ID: hitMap["_id"].(string),
			}
			if idx, ok := hitMap["_index"].(string); ok {
				searchHit.Index = idx
			}
			if score, ok := hitMap["_score"].(float64); ok {
				searchHit.Score = score
			}
			if source, ok := hitMap["_source"].(map[string]interface{}); ok {
				searchHit.Source = source
			}
			if highlights, ok := hitMap["highlight"].(map[string]interface{}); ok {
				if searchHit.Source == nil {
					searchHit.Source = map[string]interface{}{}
				}
				searchHit.Source["_highlights"] = highlights
			}
			resp.Results = append(resp.Results, searchHit)
		}
	}

	if resp.Total > 0 {
		resp.TotalPages = int((resp.Total + int64(params.PerPage) - 1) / int64(params.PerPage))
	}

	return resp
}

func buildExtFilters(params *models.SearchParams) []map[string]interface{} {
	var filters []map[string]interface{}
	if params.WorkspaceID != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{"workspace_id": params.WorkspaceID},
		})
	}
	if params.DateFrom != "" || params.DateTo != "" {
		rangeFilter := map[string]interface{}{}
		if params.DateFrom != "" {
			rangeFilter["gte"] = params.DateFrom
		}
		if params.DateTo != "" {
			rangeFilter["lte"] = params.DateTo
		}
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{"created_at": rangeFilter},
		})
	}
	return filters
}

func buildExtQuery(must []map[string]interface{}, filters []map[string]interface{}, params *models.SearchParams) map[string]interface{} {
	boolQuery := map[string]interface{}{
		"must": must,
	}
	if len(filters) > 0 {
		boolQuery["filter"] = filters
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
		"from": params.From(),
		"size": params.PerPage,
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"title":        map[string]interface{}{},
				"content":      map[string]interface{}{},
				"description":  map[string]interface{}{},
				"name":         map[string]interface{}{},
				"display_name": map[string]interface{}{},
			},
			"pre_tags":  []string{"<em>"},
			"post_tags": []string{"</em>"},
		},
	}

	switch params.Sort {
	case "newest":
		query["sort"] = []map[string]interface{}{{"created_at": "desc"}, {"_score": "desc"}}
	case "oldest":
		query["sort"] = []map[string]interface{}{{"created_at": "asc"}, {"_score": "desc"}}
	}

	return query
}

func extEmptyResponse(params *models.SearchParams) *models.SearchResponse {
	return &models.SearchResponse{
		Results:    []models.SearchHit{},
		Total:      0,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: 0,
	}
}
