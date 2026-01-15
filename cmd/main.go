package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var es *elasticsearch.Client

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	_ = godotenv.Load()

	var err error
	es, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{getEnv("ELASTICSEARCH_URL", "http://localhost:9200")},
	})
	if err != nil {
		log.Warnf("Failed to connect to Elasticsearch: %v", err)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "search-service"})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/search", globalSearch)
		api.GET("/search/messages", searchMessages)
		api.GET("/search/files", searchFiles)
		api.GET("/search/users", searchUsers)
		api.GET("/search/channels", searchChannels)
		api.POST("/index/message", indexMessage)
		api.POST("/index/file", indexFile)
		api.DELETE("/index/:type/:id", deleteFromIndex)
	}

	srv := &http.Server{Addr: ":" + getEnv("PORT", "3013"), Handler: r}
	go func() {
		log.Infof("Search service starting on port %s", getEnv("PORT", "3013"))
		srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func globalSearch(c *gin.Context) {
	query := c.Query("q")
	workspaceID := c.Query("workspace_id")

	if query == "" {
		c.JSON(400, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"multi_match": map[string]interface{}{
						"query":  query,
						"fields": []string{"content", "title", "name", "username"},
					}},
				},
				"filter": []map[string]interface{}{
					{"term": map[string]interface{}{"workspace_id": workspaceID}},
				},
			},
		},
		"size": 50,
	}

	results, err := executeSearch("quckapp_*", searchQuery)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, results)
}

func searchMessages(c *gin.Context) {
	query := c.Query("q")
	channelID := c.Query("channel_id")
	from := c.Query("from")

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"match": map[string]interface{}{"content": query}},
				},
			},
		},
		"sort": []map[string]interface{}{
			{"created_at": "desc"},
		},
		"size": 100,
	}

	if channelID != "" {
		searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = []map[string]interface{}{
			{"term": map[string]interface{}{"channel_id": channelID}},
		}
	}
	if from != "" {
		searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]map[string]interface{}),
			map[string]interface{}{"term": map[string]interface{}{"user_id": from}},
		)
	}

	results, err := executeSearch("quckapp_messages", searchQuery)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, results)
}

func searchFiles(c *gin.Context) {
	query := c.Query("q")
	fileType := c.Query("type")

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"multi_match": map[string]interface{}{
						"query":  query,
						"fields": []string{"filename", "content"},
					}},
				},
			},
		},
		"size": 50,
	}

	if fileType != "" {
		searchQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = []map[string]interface{}{
			{"term": map[string]interface{}{"file_type": fileType}},
		}
	}

	results, err := executeSearch("quckapp_files", searchQuery)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, results)
}

func searchUsers(c *gin.Context) {
	query := c.Query("q")

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"username", "display_name", "email"},
				"type":   "phrase_prefix",
			},
		},
		"size": 20,
	}

	results, err := executeSearch("quckapp_users", searchQuery)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, results)
}

func searchChannels(c *gin.Context) {
	query := c.Query("q")
	workspaceID := c.Query("workspace_id")

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"multi_match": map[string]interface{}{
						"query":  query,
						"fields": []string{"name", "description", "topic"},
					}},
				},
				"filter": []map[string]interface{}{
					{"term": map[string]interface{}{"workspace_id": workspaceID}},
				},
			},
		},
		"size": 20,
	}

	results, err := executeSearch("quckapp_channels", searchQuery)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, results)
}

func indexMessage(c *gin.Context) {
	var doc map[string]interface{}
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := indexDocument("quckapp_messages", doc["id"].(string), doc); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"indexed": true})
}

func indexFile(c *gin.Context) {
	var doc map[string]interface{}
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := indexDocument("quckapp_files", doc["id"].(string), doc); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"indexed": true})
}

func deleteFromIndex(c *gin.Context) {
	indexType := c.Param("type")
	id := c.Param("id")

	index := "quckapp_" + indexType
	if es != nil {
		es.Delete(index, id)
	}
	c.JSON(204, nil)
}

func executeSearch(index string, query map[string]interface{}) (map[string]interface{}, error) {
	if es == nil {
		return map[string]interface{}{"hits": []interface{}{}}, nil
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(query)

	res, err := es.Search(
		es.Search.WithIndex(index),
		es.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)
	return result, nil
}

func indexDocument(index, id string, doc map[string]interface{}) error {
	if es == nil {
		return nil
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(doc)

	_, err := es.Index(index, &buf, es.Index.WithDocumentID(id))
	return err
}
