package models

import "time"

// ── Index Documents ──

type MessageDocument struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	ChannelID   string    `json:"channel_id"`
	WorkspaceID string    `json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type FileDocument struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	Content     string    `json:"content"`
	FileType    string    `json:"file_type"`
	MimeType    string    `json:"mime_type"`
	Size        int64     `json:"size"`
	UserID      string    `json:"user_id"`
	ChannelID   string    `json:"channel_id"`
	WorkspaceID string    `json:"workspace_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserDocument struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	WorkspaceID string `json:"workspace_id"`
}

type ChannelDocument struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Topic       string `json:"topic"`
	Type        string `json:"type"`
	WorkspaceID string `json:"workspace_id"`
}

// ── Search Parameters ──

type SearchParams struct {
	Query       string `form:"q"`
	WorkspaceID string `form:"workspace_id"`
	ChannelID   string `form:"channel_id"`
	UserID      string `form:"user_id"`
	FileType    string `form:"type"`
	DateFrom    string `form:"date_from"`
	DateTo      string `form:"date_to"`
	Page        int    `form:"page,default=1"`
	PerPage     int    `form:"per_page,default=20"`
	Sort        string `form:"sort,default=relevance"` // relevance, newest, oldest
}

func (p *SearchParams) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
}

func (p *SearchParams) From() int {
	return (p.Page - 1) * p.PerPage
}

// ── Search Response ──

type SearchResponse struct {
	Results    []SearchHit `json:"results"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

type SearchHit struct {
	Index  string                 `json:"index"`
	ID     string                 `json:"id"`
	Score  float64                `json:"score"`
	Source map[string]interface{} `json:"source"`
}

type GlobalSearchResponse struct {
	Messages *SearchResponse `json:"messages"`
	Files    *SearchResponse `json:"files"`
	Users    *SearchResponse `json:"users"`
	Channels *SearchResponse `json:"channels"`
}

// ── Index Requests ──

type IndexRequest struct {
	Index    string                 `json:"index" binding:"required"`
	ID       string                 `json:"id" binding:"required"`
	Document map[string]interface{} `json:"document" binding:"required"`
}

type BulkIndexRequest struct {
	Documents []IndexRequest `json:"documents" binding:"required,min=1,max=100"`
}

type BulkIndexResponse struct {
	Indexed int      `json:"indexed"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

type ReindexRequest struct {
	Index string `json:"index" binding:"required"`
}

// ── Suggestion ──

type SuggestionResponse struct {
	Suggestions []string `json:"suggestions"`
}

// ── Additional Index Documents ──

type BookmarkDocument struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Tags        []string  `json:"tags"`
	UserID      string    `json:"user_id"`
	WorkspaceID string    `json:"workspace_id"`
	FolderID    string    `json:"folder_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type TaskDocument struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	AssigneeID  string    `json:"assignee_id"`
	UserID      string    `json:"user_id"`
	WorkspaceID string    `json:"workspace_id"`
	ChannelID   string    `json:"channel_id"`
	DueDate     string    `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
}

type ReactionDocument struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	ChannelID string    `json:"channel_id"`
	UserID    string    `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type EmojiDocument struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	WorkspaceID string `json:"workspace_id"`
	IsCustom    bool   `json:"is_custom"`
}

// ── Search History ──

type SearchHistory struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Query       string    `json:"query"`
	SearchType  string    `json:"search_type"` // global, messages, files, users, channels
	WorkspaceID string    `json:"workspace_id"`
	ResultCount int64     `json:"result_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// ── Saved Searches ──

type SavedSearch struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id"`
	Name        string            `json:"name"`
	Query       string            `json:"query"`
	SearchType  string            `json:"search_type"`
	Filters     map[string]string `json:"filters,omitempty"`
	WorkspaceID string            `json:"workspace_id"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type CreateSavedSearchRequest struct {
	Name        string            `json:"name" binding:"required"`
	Query       string            `json:"query" binding:"required"`
	SearchType  string            `json:"search_type" binding:"required"`
	Filters     map[string]string `json:"filters"`
	WorkspaceID string            `json:"workspace_id"`
}

type UpdateSavedSearchRequest struct {
	Name    string            `json:"name"`
	Query   string            `json:"query"`
	Filters map[string]string `json:"filters"`
}

// ── Index Management ──

type IndexMapping struct {
	Index    string                 `json:"index" binding:"required"`
	Mappings map[string]interface{} `json:"mappings" binding:"required"`
}

type IndexSettings struct {
	Index    string                 `json:"index" binding:"required"`
	Settings map[string]interface{} `json:"settings" binding:"required"`
}

type IndexAliasRequest struct {
	Index string `json:"index" binding:"required"`
	Alias string `json:"alias" binding:"required"`
}

type IndexInfo struct {
	Name      string                 `json:"name"`
	Health    string                 `json:"health"`
	Status    string                 `json:"status"`
	DocCount  string                 `json:"doc_count"`
	StoreSize string                 `json:"store_size"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
	Mappings  map[string]interface{} `json:"mappings,omitempty"`
}

// ── Analytics ──

type SearchAnalytics struct {
	TotalSearches     int64            `json:"total_searches"`
	UniqueUsers       int64            `json:"unique_users"`
	TopQueries        []QueryCount     `json:"top_queries"`
	SearchesByType    map[string]int64 `json:"searches_by_type"`
	AvgResultCount    float64          `json:"avg_result_count"`
	ZeroResultQueries []string         `json:"zero_result_queries"`
}

type QueryCount struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
}

// ── Advanced Search ──

type AdvancedSearchParams struct {
	Queries     []SubQuery `json:"queries" binding:"required,min=1"`
	WorkspaceID string     `json:"workspace_id"`
	Page        int        `json:"page"`
	PerPage     int        `json:"per_page"`
}

type SubQuery struct {
	Index  string                 `json:"index" binding:"required"`
	Query  string                 `json:"query" binding:"required"`
	Fields []string               `json:"fields"`
	Filter map[string]interface{} `json:"filter"`
	Boost  float64                `json:"boost"`
}

// ── Scroll / Cursor ──

type ScrollRequest struct {
	Index   string `json:"index" binding:"required"`
	Query   string `json:"query"`
	Size    int    `json:"size"`
	ScrollID string `json:"scroll_id"`
}

type ScrollResponse struct {
	ScrollID string      `json:"scroll_id"`
	Results  []SearchHit `json:"results"`
	Total    int64       `json:"total"`
	HasMore  bool        `json:"has_more"`
}

// ── Aggregation ──

type AggregationRequest struct {
	Index   string `json:"index" binding:"required"`
	Field   string `json:"field" binding:"required"`
	Size    int    `json:"size"`
}

type AggregationResponse struct {
	Buckets []AggregationBucket `json:"buckets"`
}

type AggregationBucket struct {
	Key      string `json:"key"`
	DocCount int64  `json:"doc_count"`
}

// ── Batch Operations ──

type BatchDeleteRequest struct {
	Index string   `json:"index" binding:"required"`
	IDs   []string `json:"ids" binding:"required,min=1,max=100"`
}

type BatchDeleteResponse struct {
	Deleted int      `json:"deleted"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

type UpdateDocumentRequest struct {
	Document map[string]interface{} `json:"document" binding:"required"`
}

// ── More Index Types ──

type IndexUserRequest struct {
	ID          string `json:"id" binding:"required"`
	Username    string `json:"username" binding:"required"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	WorkspaceID string `json:"workspace_id" binding:"required"`
}

type IndexChannelRequest struct {
	ID          string `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Topic       string `json:"topic"`
	Type        string `json:"type"`
	WorkspaceID string `json:"workspace_id" binding:"required"`
}

type IndexBookmarkRequest struct {
	ID          string   `json:"id" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Tags        []string `json:"tags"`
	UserID      string   `json:"user_id" binding:"required"`
	WorkspaceID string   `json:"workspace_id" binding:"required"`
}

type IndexTaskRequest struct {
	ID          string `json:"id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	AssigneeID  string `json:"assignee_id"`
	UserID      string `json:"user_id" binding:"required"`
	WorkspaceID string `json:"workspace_id" binding:"required"`
}

// -- Search Facets/Filters --

type FacetRequest struct {
	Index string `json:"index" binding:"required"`
	Field string `json:"field" binding:"required"`
	Query string `json:"query"`
	Size  int    `json:"size"`
}

type FacetResult struct {
	Field   string        `json:"field"`
	Buckets []FacetBucket `json:"buckets"`
}

type FacetBucket struct {
	Key      string `json:"key"`
	DocCount int64  `json:"doc_count"`
}

type FilterOptions struct {
	Types      []string `json:"types"`
	Channels   []string `json:"channels"`
	Users      []string `json:"users"`
	DateRanges []string `json:"date_ranges"`
}

// -- Synonym Management --

type SynonymGroup struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Terms       []string  `json:"terms"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateSynonymRequest struct {
	Terms       []string `json:"terms" binding:"required,min=2"`
	WorkspaceID string   `json:"workspace_id" binding:"required"`
}

type UpdateSynonymRequest struct {
	Terms []string `json:"terms" binding:"required,min=2"`
}

// -- Relevance Tuning --

type RelevanceConfig struct {
	ID              string             `json:"id"`
	WorkspaceID     string             `json:"workspace_id"`
	FieldBoosts     map[string]float64 `json:"field_boosts"`
	TitleBoost      float64            `json:"title_boost"`
	ContentBoost    float64            `json:"content_boost"`
	RecencyWeight   float64            `json:"recency_weight"`
	ExactMatchBoost float64            `json:"exact_match_boost"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

type UpdateRelevanceRequest struct {
	FieldBoosts     map[string]float64 `json:"field_boosts"`
	TitleBoost      *float64           `json:"title_boost"`
	ContentBoost    *float64           `json:"content_boost"`
	RecencyWeight   *float64           `json:"recency_weight"`
	ExactMatchBoost *float64           `json:"exact_match_boost"`
}

type RelevancePreview struct {
	Query   string           `json:"query"`
	Results []SearchHit      `json:"results"`
	Config  *RelevanceConfig `json:"config"`
}

// -- Search Alerts --

type SearchAlert struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	WorkspaceID   string    `json:"workspace_id"`
	Name          string    `json:"name"`
	Query         string    `json:"query"`
	SearchType    string    `json:"search_type"`
	Frequency     string    `json:"frequency"`
	LastTriggered time.Time `json:"last_triggered"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateAlertRequest struct {
	Name        string `json:"name" binding:"required"`
	Query       string `json:"query" binding:"required"`
	SearchType  string `json:"search_type" binding:"required"`
	Frequency   string `json:"frequency" binding:"required"`
	WorkspaceID string `json:"workspace_id" binding:"required"`
}

type UpdateAlertRequest struct {
	Name      string `json:"name"`
	Query     string `json:"query"`
	Frequency string `json:"frequency"`
	IsActive  *bool  `json:"is_active"`
}

type AlertHistory struct {
	ID          string    `json:"id"`
	AlertID     string    `json:"alert_id"`
	ResultCount int64     `json:"result_count"`
	TriggeredAt time.Time `json:"triggered_at"`
}

// -- Spell Check / Did You Mean --

type SpellCheckResponse struct {
	Original    string   `json:"original"`
	Suggestions []string `json:"suggestions"`
	Corrected   string   `json:"corrected"`
}

type DidYouMeanResponse struct {
	Original   string  `json:"original"`
	Suggestion string  `json:"suggestion"`
	Confidence float64 `json:"confidence"`
}

// -- Search Permissions/Scoping --

type SearchScope struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	WorkspaceID    string    `json:"workspace_id"`
	AllowedIndices []string  `json:"allowed_indices"`
	DeniedChannels []string  `json:"denied_channels"`
	CreatedAt      time.Time `json:"created_at"`
}

type SetSearchScopeRequest struct {
	AllowedIndices []string `json:"allowed_indices"`
	DeniedChannels []string `json:"denied_channels"`
	WorkspaceID    string   `json:"workspace_id" binding:"required"`
}
