package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
)

const (
	indexMessages = "quckapp_messages"
	indexFiles    = "quckapp_files"
	indexUsers    = "quckapp_users"
	indexChannels = "quckapp_channels"
	cacheTTL     = 5 * time.Minute
)

type SearchService struct {
	es     *elasticsearch.Client
	redis  *redis.Client
	logger *logrus.Logger
}

func NewSearchService(es *elasticsearch.Client, redis *redis.Client, logger *logrus.Logger) *SearchService {
	return &SearchService{es: es, redis: redis, logger: logger}
}

// ── Global Search ──

func (s *SearchService) GlobalSearch(ctx context.Context, params *models.SearchParams) (*models.GlobalSearchResponse, error) {
	params.Validate()

	// Limit per-type results in global search
	perType := 5
	if params.PerPage > 5 {
		perType = params.PerPage / 4
	}

	msgParams := *params
	msgParams.PerPage = perType
	filesParams := *params
	filesParams.PerPage = perType
	usersParams := *params
	usersParams.PerPage = perType
	channelsParams := *params
	channelsParams.PerPage = perType

	messages, _ := s.SearchMessages(ctx, &msgParams)
	files, _ := s.SearchFiles(ctx, &filesParams)
	users, _ := s.SearchUsers(ctx, &usersParams)
	channels, _ := s.SearchChannels(ctx, &channelsParams)

	return &models.GlobalSearchResponse{
		Messages: messages,
		Files:    files,
		Users:    users,
		Channels: channels,
	}, nil
}

// ── Message Search ──

func (s *SearchService) SearchMessages(ctx context.Context, params *models.SearchParams) (*models.SearchResponse, error) {
	params.Validate()

	cacheKey := s.buildCacheKey("msg", params)
	if cached := s.getFromCache(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	must := []map[string]interface{}{
		{"match": map[string]interface{}{
			"content": map[string]interface{}{
				"query":     params.Query,
				"fuzziness": "AUTO",
			},
		}},
	}

	filters := s.buildFilters(params)
	if params.ChannelID != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{"channel_id": params.ChannelID},
		})
	}
	if params.UserID != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{"user_id": params.UserID},
		})
	}

	query := s.buildQuery(must, filters, params)
	result, err := s.executeSearch(indexMessages, query)
	if err != nil {
		return emptyResponse(params), nil
	}

	resp := s.parseResponse(result, params)
	s.setCache(ctx, cacheKey, resp)
	return resp, nil
}

// ── File Search ──

func (s *SearchService) SearchFiles(ctx context.Context, params *models.SearchParams) (*models.SearchResponse, error) {
	params.Validate()

	cacheKey := s.buildCacheKey("file", params)
	if cached := s.getFromCache(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	must := []map[string]interface{}{
		{"multi_match": map[string]interface{}{
			"query":     params.Query,
			"fields":    []string{"filename^2", "content"},
			"fuzziness": "AUTO",
		}},
	}

	filters := s.buildFilters(params)
	if params.FileType != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{"file_type": params.FileType},
		})
	}
	if params.ChannelID != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{"channel_id": params.ChannelID},
		})
	}

	query := s.buildQuery(must, filters, params)
	result, err := s.executeSearch(indexFiles, query)
	if err != nil {
		return emptyResponse(params), nil
	}

	resp := s.parseResponse(result, params)
	s.setCache(ctx, cacheKey, resp)
	return resp, nil
}

// ── User Search ──

func (s *SearchService) SearchUsers(ctx context.Context, params *models.SearchParams) (*models.SearchResponse, error) {
	params.Validate()

	cacheKey := s.buildCacheKey("user", params)
	if cached := s.getFromCache(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	must := []map[string]interface{}{
		{"multi_match": map[string]interface{}{
			"query":  params.Query,
			"fields": []string{"username^3", "display_name^2", "email"},
			"type":   "phrase_prefix",
		}},
	}

	filters := s.buildFilters(params)
	query := s.buildQuery(must, filters, params)

	result, err := s.executeSearch(indexUsers, query)
	if err != nil {
		return emptyResponse(params), nil
	}

	resp := s.parseResponse(result, params)
	s.setCache(ctx, cacheKey, resp)
	return resp, nil
}

// ── Channel Search ──

func (s *SearchService) SearchChannels(ctx context.Context, params *models.SearchParams) (*models.SearchResponse, error) {
	params.Validate()

	cacheKey := s.buildCacheKey("ch", params)
	if cached := s.getFromCache(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	must := []map[string]interface{}{
		{"multi_match": map[string]interface{}{
			"query":     params.Query,
			"fields":    []string{"name^3", "description", "topic"},
			"fuzziness": "AUTO",
		}},
	}

	filters := s.buildFilters(params)
	query := s.buildQuery(must, filters, params)

	result, err := s.executeSearch(indexChannels, query)
	if err != nil {
		return emptyResponse(params), nil
	}

	resp := s.parseResponse(result, params)
	s.setCache(ctx, cacheKey, resp)
	return resp, nil
}

// ── Suggest / Autocomplete ──

func (s *SearchService) Suggest(ctx context.Context, query, workspaceID string) (*models.SuggestionResponse, error) {
	if s.es == nil || query == "" {
		return &models.SuggestionResponse{Suggestions: []string{}}, nil
	}

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"multi_match": map[string]interface{}{
						"query":  query,
						"fields": []string{"name", "username", "display_name", "filename"},
						"type":   "phrase_prefix",
					}},
				},
				"filter": []map[string]interface{}{
					{"term": map[string]interface{}{"workspace_id": workspaceID}},
				},
			},
		},
		"size":    10,
		"_source": []string{"name", "username", "display_name", "filename"},
	}

	result, err := s.executeSearch("quckapp_*", searchQuery)
	if err != nil {
		return &models.SuggestionResponse{Suggestions: []string{}}, nil
	}

	suggestions := []string{}
	seen := map[string]bool{}
	if hits, ok := result["hits"].(map[string]interface{}); ok {
		if hitList, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitList {
				hitMap := hit.(map[string]interface{})
				if source, ok := hitMap["_source"].(map[string]interface{}); ok {
					for _, field := range []string{"name", "username", "display_name", "filename"} {
						if val, ok := source[field].(string); ok && val != "" && !seen[val] {
							suggestions = append(suggestions, val)
							seen[val] = true
						}
					}
				}
			}
		}
	}

	return &models.SuggestionResponse{Suggestions: suggestions}, nil
}

// ── Index Operations ──

func (s *SearchService) IndexDocument(ctx context.Context, index, id string, doc map[string]interface{}) error {
	if s.es == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(doc); err != nil {
		return err
	}

	_, err := s.es.Index(index, &buf, s.es.Index.WithDocumentID(id))
	if err != nil {
		return err
	}

	// Invalidate related caches
	s.invalidateCache(ctx, index)
	return nil
}

func (s *SearchService) BulkIndex(ctx context.Context, docs []models.IndexRequest) *models.BulkIndexResponse {
	resp := &models.BulkIndexResponse{}

	for _, doc := range docs {
		err := s.IndexDocument(ctx, doc.Index, doc.ID, doc.Document)
		if err != nil {
			resp.Failed++
			resp.Errors = append(resp.Errors, fmt.Sprintf("%s/%s: %s", doc.Index, doc.ID, err.Error()))
		} else {
			resp.Indexed++
		}
	}

	return resp
}

func (s *SearchService) DeleteDocument(ctx context.Context, index, id string) error {
	if s.es == nil {
		return nil
	}

	_, err := s.es.Delete(index, id)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx, index)
	return nil
}

func (s *SearchService) Reindex(ctx context.Context, index string) error {
	if s.es == nil {
		return nil
	}

	// Create a reindex request (source and dest are the same, which refreshes)
	body := map[string]interface{}{
		"source": map[string]interface{}{"index": index},
		"dest":   map[string]interface{}{"index": index + "_reindexed"},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)

	res, err := s.es.Reindex(&buf)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	s.invalidateCache(ctx, index)
	return nil
}

// ── Helpers ──

func (s *SearchService) buildFilters(params *models.SearchParams) []map[string]interface{} {
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

func (s *SearchService) buildQuery(must []map[string]interface{}, filters []map[string]interface{}, params *models.SearchParams) map[string]interface{} {
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
				"content":      map[string]interface{}{},
				"filename":     map[string]interface{}{},
				"name":         map[string]interface{}{},
				"display_name": map[string]interface{}{},
				"description":  map[string]interface{}{},
			},
			"pre_tags":  []string{"<em>"},
			"post_tags": []string{"</em>"},
		},
	}

	// Sort
	switch params.Sort {
	case "newest":
		query["sort"] = []map[string]interface{}{{"created_at": "desc"}, {"_score": "desc"}}
	case "oldest":
		query["sort"] = []map[string]interface{}{{"created_at": "asc"}, {"_score": "desc"}}
	default:
		// relevance - default ES scoring
	}

	return query
}

func (s *SearchService) executeSearch(index string, query map[string]interface{}) (map[string]interface{}, error) {
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

func (s *SearchService) parseResponse(result map[string]interface{}, params *models.SearchParams) *models.SearchResponse {
	resp := &models.SearchResponse{
		Results: []models.SearchHit{},
		Page:    params.Page,
		PerPage: params.PerPage,
	}

	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return resp
	}

	// Total
	if total, ok := hits["total"].(map[string]interface{}); ok {
		if val, ok := total["value"].(float64); ok {
			resp.Total = int64(val)
		}
	}

	// Hits
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
			// Include highlights in source
			if highlights, ok := hitMap["highlight"].(map[string]interface{}); ok {
				if searchHit.Source == nil {
					searchHit.Source = map[string]interface{}{}
				}
				searchHit.Source["_highlights"] = highlights
			}
			resp.Results = append(resp.Results, searchHit)
		}
	}

	// Total pages
	if resp.Total > 0 {
		resp.TotalPages = int((resp.Total + int64(params.PerPage) - 1) / int64(params.PerPage))
	}

	return resp
}

func emptyResponse(params *models.SearchParams) *models.SearchResponse {
	return &models.SearchResponse{
		Results:    []models.SearchHit{},
		Total:      0,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: 0,
	}
}

// ── Cache ──

func (s *SearchService) buildCacheKey(prefix string, params *models.SearchParams) string {
	return fmt.Sprintf("search:%s:%s:%s:%d:%d:%s",
		prefix, params.Query, params.WorkspaceID, params.Page, params.PerPage, params.Sort)
}

func (s *SearchService) getFromCache(ctx context.Context, key string) *models.SearchResponse {
	if s.redis == nil {
		return nil
	}
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var resp models.SearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil
	}
	return &resp
}

func (s *SearchService) setCache(ctx context.Context, key string, resp *models.SearchResponse) {
	if s.redis == nil {
		return
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	s.redis.Set(ctx, key, data, cacheTTL)
}

func (s *SearchService) invalidateCache(ctx context.Context, index string) {
	if s.redis == nil {
		return
	}
	// Extract type from index name
	parts := strings.Split(index, "_")
	if len(parts) < 2 {
		return
	}
	pattern := "search:" + parts[len(parts)-1][:3] + ":*"
	keys, _ := s.redis.Keys(ctx, pattern).Result()
	if len(keys) > 0 {
		s.redis.Del(ctx, keys...)
	}
}

// ── Health ──

func (s *SearchService) HealthCheck() map[string]interface{} {
	health := map[string]interface{}{
		"service": "search-service",
		"status":  "healthy",
	}

	if s.es != nil {
		res, err := s.es.Info()
		if err != nil {
			health["elasticsearch"] = "disconnected"
		} else {
			res.Body.Close()
			health["elasticsearch"] = "connected"
		}
	} else {
		health["elasticsearch"] = "not configured"
	}

	if s.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := s.redis.Ping(ctx).Err(); err != nil {
			health["redis"] = "disconnected"
		} else {
			health["redis"] = "connected"
		}
	} else {
		health["redis"] = "not configured"
	}

	return health
}
