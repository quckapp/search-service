package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
)

type IndexManagementService struct {
	es     *elasticsearch.Client
	logger *logrus.Logger
}

func NewIndexManagementService(es *elasticsearch.Client, logger *logrus.Logger) *IndexManagementService {
	return &IndexManagementService{es: es, logger: logger}
}

func (s *IndexManagementService) ListIndices(ctx context.Context) ([]models.IndexInfo, error) {
	if s.es == nil {
		return []models.IndexInfo{}, nil
	}

	res, err := s.es.Cat.Indices(
		s.es.Cat.Indices.WithIndex("quckapp_*"),
		s.es.Cat.Indices.WithFormat("json"),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var indices []map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return []models.IndexInfo{}, nil
	}

	var result []models.IndexInfo
	for _, idx := range indices {
		info := models.IndexInfo{
			Name:   getString(idx, "index"),
			Health: getString(idx, "health"),
			Status: getString(idx, "status"),
		}
		if dc, ok := idx["docs.count"].(string); ok {
			info.DocCount = dc
		}
		if ss, ok := idx["store.size"].(string); ok {
			info.StoreSize = ss
		}
		result = append(result, info)
	}
	return result, nil
}

func (s *IndexManagementService) GetIndexInfo(ctx context.Context, index string) (*models.IndexInfo, error) {
	if s.es == nil {
		return nil, fmt.Errorf("elasticsearch not available")
	}

	// Get mappings
	mappingsRes, err := s.es.Indices.GetMapping(s.es.Indices.GetMapping.WithIndex(index))
	if err != nil {
		return nil, err
	}
	defer mappingsRes.Body.Close()

	var mappings map[string]interface{}
	json.NewDecoder(mappingsRes.Body).Decode(&mappings)

	// Get settings
	settingsRes, err := s.es.Indices.GetSettings(s.es.Indices.GetSettings.WithIndex(index))
	if err != nil {
		return nil, err
	}
	defer settingsRes.Body.Close()

	var settings map[string]interface{}
	json.NewDecoder(settingsRes.Body).Decode(&settings)

	return &models.IndexInfo{
		Name:     index,
		Mappings: mappings,
		Settings: settings,
	}, nil
}

func (s *IndexManagementService) CreateIndex(ctx context.Context, index string, mappings map[string]interface{}) error {
	if s.es == nil {
		return nil
	}

	body := map[string]interface{}{
		"mappings": mappings,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)

	res, err := s.es.Indices.Create(index, s.es.Indices.Create.WithBody(&buf))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to create index: %s", string(bodyBytes))
	}
	return nil
}

func (s *IndexManagementService) DeleteIndex(ctx context.Context, index string) error {
	if s.es == nil {
		return nil
	}

	// Safety: only allow deleting quckapp_ prefixed indices
	if !strings.HasPrefix(index, "quckapp_") {
		return fmt.Errorf("can only delete quckapp_* indices")
	}

	res, err := s.es.Indices.Delete([]string{index})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (s *IndexManagementService) PutMapping(ctx context.Context, req *models.IndexMapping) error {
	if s.es == nil {
		return nil
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(req.Mappings)

	res, err := s.es.Indices.PutMapping([]string{req.Index}, &buf)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (s *IndexManagementService) UpdateSettings(ctx context.Context, req *models.IndexSettings) error {
	if s.es == nil {
		return nil
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(req.Settings)

	res, err := s.es.Indices.PutSettings(&buf, s.es.Indices.PutSettings.WithIndex(req.Index))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (s *IndexManagementService) CreateAlias(ctx context.Context, req *models.IndexAliasRequest) error {
	if s.es == nil {
		return nil
	}

	body := map[string]interface{}{
		"actions": []map[string]interface{}{
			{"add": map[string]interface{}{
				"index": req.Index,
				"alias": req.Alias,
			}},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(body)

	res, err := s.es.Indices.UpdateAliases(&buf)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (s *IndexManagementService) DeleteAlias(ctx context.Context, index, alias string) error {
	if s.es == nil {
		return nil
	}

	res, err := s.es.Indices.DeleteAlias([]string{index}, []string{alias})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (s *IndexManagementService) RefreshIndex(ctx context.Context, index string) error {
	if s.es == nil {
		return nil
	}

	res, err := s.es.Indices.Refresh(s.es.Indices.Refresh.WithIndex(index))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (s *IndexManagementService) FlushIndex(ctx context.Context, index string) error {
	if s.es == nil {
		return nil
	}

	res, err := s.es.Indices.Flush(s.es.Indices.Flush.WithIndex(index))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
