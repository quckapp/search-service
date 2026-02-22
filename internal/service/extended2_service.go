package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// ── Extended Models ──

type SearchTemplate struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	WorkspaceID string         `json:"workspace_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Query       string         `json:"query"`
	Filters     map[string]any `json:"filters"`
	IsPublic    bool           `json:"is_public"`
	UseCount    int            `json:"use_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type SearchBookmarkItem struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Query       string    `json:"query"`
	ResultID    string    `json:"result_id"`
	ResultType  string    `json:"result_type"`
	Title       string    `json:"title"`
	Snippet     string    `json:"snippet"`
	CreatedAt   time.Time `json:"created_at"`
}

type SearchFeedback struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Query       string    `json:"query"`
	ResultID    string    `json:"result_id"`
	Rating      int       `json:"rating"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
}

type SearchABTest struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ConfigA     map[string]any `json:"config_a"`
	ConfigB     map[string]any `json:"config_b"`
	SplitPct    int            `json:"split_pct"`
	IsActive    bool           `json:"is_active"`
	StartedAt   time.Time      `json:"started_at"`
	CreatedAt   time.Time      `json:"created_at"`
}

type SearchPipeline struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []PipelineStep `json:"steps"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type PipelineStep struct {
	Name   string         `json:"name"`
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
	Order  int            `json:"order"`
}

type StopWord struct {
	ID        string    `json:"id"`
	Word      string    `json:"word"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
}

type QueryRewrite struct {
	ID          string    `json:"id"`
	Pattern     string    `json:"pattern"`
	Replacement string    `json:"replacement"`
	IsActive    bool      `json:"is_active"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
}

type IndexSchedule struct {
	ID        string    `json:"id"`
	IndexName string    `json:"index_name"`
	Schedule  string    `json:"schedule"`
	IsActive  bool      `json:"is_active"`
	LastRun   time.Time `json:"last_run"`
	NextRun   time.Time `json:"next_run"`
	CreatedAt time.Time `json:"created_at"`
}

// ── Extended2 Service ──

type Extended2Service struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func NewExtended2Service(redis *redis.Client, logger *logrus.Logger) *Extended2Service {
	return &Extended2Service{redis: redis, logger: logger}
}

// Search Templates
func (s *Extended2Service) CreateTemplate(ctx context.Context, t *SearchTemplate) error {
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return s.set(ctx, fmt.Sprintf("search_template:%s:%s", t.UserID, t.ID), t, 0)
}

func (s *Extended2Service) GetTemplate(ctx context.Context, userID, id string) (*SearchTemplate, error) {
	var t SearchTemplate
	err := s.get(ctx, fmt.Sprintf("search_template:%s:%s", userID, id), &t)
	if err != nil { return nil, err }
	return &t, nil
}

func (s *Extended2Service) ListTemplates(ctx context.Context, userID string) ([]SearchTemplate, error) {
	return listByPattern[SearchTemplate](ctx, s, fmt.Sprintf("search_template:%s:*", userID))
}

func (s *Extended2Service) UpdateTemplate(ctx context.Context, userID, id string, updates map[string]any) error {
	t, err := s.GetTemplate(ctx, userID, id)
	if err != nil { return err }
	if name, ok := updates["name"].(string); ok { t.Name = name }
	if desc, ok := updates["description"].(string); ok { t.Description = desc }
	if q, ok := updates["query"].(string); ok { t.Query = q }
	if pub, ok := updates["is_public"].(bool); ok { t.IsPublic = pub }
	t.UpdatedAt = time.Now()
	return s.set(ctx, fmt.Sprintf("search_template:%s:%s", userID, id), t, 0)
}

func (s *Extended2Service) DeleteTemplate(ctx context.Context, userID, id string) error {
	return s.del(ctx, fmt.Sprintf("search_template:%s:%s", userID, id))
}

func (s *Extended2Service) IncrementTemplateUsage(ctx context.Context, userID, id string) error {
	t, err := s.GetTemplate(ctx, userID, id)
	if err != nil { return err }
	t.UseCount++
	return s.set(ctx, fmt.Sprintf("search_template:%s:%s", userID, id), t, 0)
}

// Search Bookmarks
func (s *Extended2Service) BookmarkResult(ctx context.Context, b *SearchBookmarkItem) error {
	b.ID = uuid.New().String()
	b.CreatedAt = time.Now()
	return s.set(ctx, fmt.Sprintf("search_bookmark:%s:%s", b.UserID, b.ID), b, 0)
}

func (s *Extended2Service) ListBookmarks(ctx context.Context, userID string) ([]SearchBookmarkItem, error) {
	return listByPattern[SearchBookmarkItem](ctx, s, fmt.Sprintf("search_bookmark:%s:*", userID))
}

func (s *Extended2Service) DeleteBookmark(ctx context.Context, userID, id string) error {
	return s.del(ctx, fmt.Sprintf("search_bookmark:%s:%s", userID, id))
}

// Search Feedback
func (s *Extended2Service) SubmitFeedback(ctx context.Context, f *SearchFeedback) error {
	f.ID = uuid.New().String()
	f.CreatedAt = time.Now()
	return s.set(ctx, fmt.Sprintf("search_feedback:%s:%s", f.UserID, f.ID), f, 0)
}

func (s *Extended2Service) ListFeedback(ctx context.Context, userID string) ([]SearchFeedback, error) {
	return listByPattern[SearchFeedback](ctx, s, fmt.Sprintf("search_feedback:%s:*", userID))
}

func (s *Extended2Service) GetFeedbackStats(ctx context.Context) (map[string]any, error) {
	return map[string]any{"total_feedback": 0, "avg_rating": 0, "positive_pct": 0}, nil
}

// A/B Tests
func (s *Extended2Service) CreateABTest(ctx context.Context, t *SearchABTest) error {
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now()
	t.StartedAt = time.Now()
	return s.set(ctx, fmt.Sprintf("search_abtest:%s", t.ID), t, 0)
}

func (s *Extended2Service) GetABTest(ctx context.Context, id string) (*SearchABTest, error) {
	var t SearchABTest
	err := s.get(ctx, fmt.Sprintf("search_abtest:%s", id), &t)
	if err != nil { return nil, err }
	return &t, nil
}

func (s *Extended2Service) ListABTests(ctx context.Context) ([]SearchABTest, error) {
	return listByPattern[SearchABTest](ctx, s, "search_abtest:*")
}

func (s *Extended2Service) DeleteABTest(ctx context.Context, id string) error {
	return s.del(ctx, fmt.Sprintf("search_abtest:%s", id))
}

// Pipelines
func (s *Extended2Service) CreatePipeline(ctx context.Context, p *SearchPipeline) error {
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return s.set(ctx, fmt.Sprintf("search_pipeline:%s", p.ID), p, 0)
}

func (s *Extended2Service) GetPipeline(ctx context.Context, id string) (*SearchPipeline, error) {
	var p SearchPipeline
	err := s.get(ctx, fmt.Sprintf("search_pipeline:%s", id), &p)
	if err != nil { return nil, err }
	return &p, nil
}

func (s *Extended2Service) ListPipelines(ctx context.Context) ([]SearchPipeline, error) {
	return listByPattern[SearchPipeline](ctx, s, "search_pipeline:*")
}

func (s *Extended2Service) UpdatePipeline(ctx context.Context, id string, updates map[string]any) error {
	p, err := s.GetPipeline(ctx, id)
	if err != nil { return err }
	if name, ok := updates["name"].(string); ok { p.Name = name }
	if desc, ok := updates["description"].(string); ok { p.Description = desc }
	if active, ok := updates["is_active"].(bool); ok { p.IsActive = active }
	p.UpdatedAt = time.Now()
	return s.set(ctx, fmt.Sprintf("search_pipeline:%s", id), p, 0)
}

func (s *Extended2Service) DeletePipeline(ctx context.Context, id string) error {
	return s.del(ctx, fmt.Sprintf("search_pipeline:%s", id))
}

// Stop Words
func (s *Extended2Service) AddStopWord(ctx context.Context, sw *StopWord) error {
	sw.ID = uuid.New().String()
	sw.CreatedAt = time.Now()
	return s.set(ctx, fmt.Sprintf("stop_word:%s:%s", sw.Language, sw.ID), sw, 0)
}

func (s *Extended2Service) ListStopWords(ctx context.Context, language string) ([]StopWord, error) {
	return listByPattern[StopWord](ctx, s, fmt.Sprintf("stop_word:%s:*", language))
}

func (s *Extended2Service) DeleteStopWord(ctx context.Context, language, id string) error {
	return s.del(ctx, fmt.Sprintf("stop_word:%s:%s", language, id))
}

// Query Rewrites
func (s *Extended2Service) CreateRewrite(ctx context.Context, r *QueryRewrite) error {
	r.ID = uuid.New().String()
	r.CreatedAt = time.Now()
	r.IsActive = true
	return s.set(ctx, fmt.Sprintf("query_rewrite:%s", r.ID), r, 0)
}

func (s *Extended2Service) ListRewrites(ctx context.Context) ([]QueryRewrite, error) {
	return listByPattern[QueryRewrite](ctx, s, "query_rewrite:*")
}

func (s *Extended2Service) UpdateRewrite(ctx context.Context, id string, updates map[string]any) error {
	var r QueryRewrite
	if err := s.get(ctx, fmt.Sprintf("query_rewrite:%s", id), &r); err != nil { return err }
	if pattern, ok := updates["pattern"].(string); ok { r.Pattern = pattern }
	if repl, ok := updates["replacement"].(string); ok { r.Replacement = repl }
	if active, ok := updates["is_active"].(bool); ok { r.IsActive = active }
	return s.set(ctx, fmt.Sprintf("query_rewrite:%s", id), &r, 0)
}

func (s *Extended2Service) DeleteRewrite(ctx context.Context, id string) error {
	return s.del(ctx, fmt.Sprintf("query_rewrite:%s", id))
}

// Index Schedules
func (s *Extended2Service) CreateSchedule(ctx context.Context, is *IndexSchedule) error {
	is.ID = uuid.New().String()
	is.CreatedAt = time.Now()
	is.IsActive = true
	return s.set(ctx, fmt.Sprintf("index_schedule:%s", is.ID), is, 0)
}

func (s *Extended2Service) ListSchedules(ctx context.Context) ([]IndexSchedule, error) {
	return listByPattern[IndexSchedule](ctx, s, "index_schedule:*")
}

func (s *Extended2Service) UpdateSchedule(ctx context.Context, id string, updates map[string]any) error {
	var is IndexSchedule
	if err := s.get(ctx, fmt.Sprintf("index_schedule:%s", id), &is); err != nil { return err }
	if schedule, ok := updates["schedule"].(string); ok { is.Schedule = schedule }
	if active, ok := updates["is_active"].(bool); ok { is.IsActive = active }
	return s.set(ctx, fmt.Sprintf("index_schedule:%s", id), &is, 0)
}

func (s *Extended2Service) DeleteSchedule(ctx context.Context, id string) error {
	return s.del(ctx, fmt.Sprintf("index_schedule:%s", id))
}

// ── Redis Helpers ──

func (s *Extended2Service) set(ctx context.Context, key string, val any, ttl time.Duration) error {
	if s.redis == nil { return fmt.Errorf("storage not available") }
	data, err := json.Marshal(val)
	if err != nil { return err }
	return s.redis.Set(ctx, key, data, ttl).Err()
}

func (s *Extended2Service) get(ctx context.Context, key string, dest any) error {
	if s.redis == nil { return fmt.Errorf("storage not available") }
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil { return err }
	return json.Unmarshal(data, dest)
}

func (s *Extended2Service) del(ctx context.Context, key string) error {
	if s.redis == nil { return fmt.Errorf("storage not available") }
	return s.redis.Del(ctx, key).Err()
}

func listByPattern[T any](ctx context.Context, s *Extended2Service, pattern string) ([]T, error) {
	if s.redis == nil { return nil, fmt.Errorf("storage not available") }
	var results []T
	iter := s.redis.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		data, err := s.redis.Get(ctx, iter.Val()).Bytes()
		if err != nil { continue }
		var item T
		if err := json.Unmarshal(data, &item); err == nil {
			results = append(results, item)
		}
	}
	return results, nil
}
