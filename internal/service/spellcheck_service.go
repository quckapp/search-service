package service

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"

	"github.com/quckapp/search-service/internal/models"
)

type SpellCheckService struct {
	es     *elasticsearch.Client
	logger *logrus.Logger
}

func NewSpellCheckService(es *elasticsearch.Client, logger *logrus.Logger) *SpellCheckService {
	return &SpellCheckService{es: es, logger: logger}
}

func (s *SpellCheckService) GetSuggestions(ctx context.Context, text, index string) (*models.SpellCheckResponse, error) {
	resp := &models.SpellCheckResponse{
		Original:    text,
		Suggestions: []string{},
		Corrected:   text,
	}

	if s.es == nil || text == "" {
		return resp, nil
	}

	if index == "" {
		index = "quckapp_messages"
	}

	query := map[string]interface{}{
		"suggest": map[string]interface{}{
			"text": text,
			"spell_suggest": map[string]interface{}{
				"term": map[string]interface{}{
					"field":           "content",
					"suggest_mode":    "popular",
					"min_word_length": 3,
				},
			},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(query)

	res, err := s.es.Search(
		s.es.Search.WithIndex(index),
		s.es.Search.WithBody(&buf),
	)
	if err != nil {
		return resp, nil
	}
	defer res.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)

	if suggest, ok := result["suggest"].(map[string]interface{}); ok {
		if spellSuggest, ok := suggest["spell_suggest"].([]interface{}); ok {
			correctedParts := strings.Fields(text)
			for i, entry := range spellSuggest {
				entryMap := entry.(map[string]interface{})
				if options, ok := entryMap["options"].([]interface{}); ok && len(options) > 0 {
					for _, opt := range options {
						optMap := opt.(map[string]interface{})
						if suggestion, ok := optMap["text"].(string); ok {
							resp.Suggestions = append(resp.Suggestions, suggestion)
							if i < len(correctedParts) {
								correctedParts[i] = suggestion
							}
						}
					}
				}
			}
			if len(resp.Suggestions) > 0 {
				resp.Corrected = strings.Join(correctedParts, " ")
			}
		}
	}

	return resp, nil
}

func (s *SpellCheckService) DidYouMean(ctx context.Context, text, index string) (*models.DidYouMeanResponse, error) {
	resp := &models.DidYouMeanResponse{
		Original:   text,
		Suggestion: "",
		Confidence: 0,
	}

	if s.es == nil || text == "" {
		return resp, nil
	}

	if index == "" {
		index = "quckapp_messages"
	}

	query := map[string]interface{}{
		"suggest": map[string]interface{}{
			"text": text,
			"did_you_mean": map[string]interface{}{
				"phrase": map[string]interface{}{
					"field":       "content",
					"gram_size":   3,
					"confidence":  1.0,
					"max_errors":  2.0,
					"direct_generator": []map[string]interface{}{
						{
							"field":        "content",
							"suggest_mode": "popular",
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(query)

	res, err := s.es.Search(
		s.es.Search.WithIndex(index),
		s.es.Search.WithBody(&buf),
	)
	if err != nil {
		return resp, nil
	}
	defer res.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)

	if suggest, ok := result["suggest"].(map[string]interface{}); ok {
		if didYouMean, ok := suggest["did_you_mean"].([]interface{}); ok {
			for _, entry := range didYouMean {
				entryMap := entry.(map[string]interface{})
				if options, ok := entryMap["options"].([]interface{}); ok && len(options) > 0 {
					optMap := options[0].(map[string]interface{})
					if suggestion, ok := optMap["text"].(string); ok {
						resp.Suggestion = suggestion
					}
					if score, ok := optMap["score"].(float64); ok {
						resp.Confidence = score
					}
					break
				}
			}
		}
	}

	return resp, nil
}
