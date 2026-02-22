package db

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func NewElasticsearch(url string, logger *logrus.Logger) *elasticsearch.Client {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{url},
	})
	if err != nil {
		logger.Warnf("Failed to create Elasticsearch client: %v", err)
		return nil
	}

	// Test connection
	res, err := es.Info()
	if err != nil {
		logger.Warnf("Elasticsearch not available: %v", err)
		return nil
	}
	res.Body.Close()
	logger.Info("Connected to Elasticsearch")
	return es
}

func NewRedis(host, port, password string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       8,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil
	}
	return client
}
