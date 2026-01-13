package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5006"
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "search-service", "version": "1.0.0"})
	})

	api := router.Group("/api/v1/search")
	{
		api.GET("/messages", searchMessages)
		api.GET("/users", searchUsers)
		api.GET("/files", searchFiles)
		api.GET("/channels", searchChannels)
		api.GET("/all", searchAll)
		api.POST("/index", indexDocument)
		api.DELETE("/index/:id", removeFromIndex)
		api.POST("/reindex", reindex)
	}

	srv := &http.Server{Addr: ":" + port, Handler: router}

	go func() {
		log.Printf("Search service starting on port %s", port)
		srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func searchMessages(c *gin.Context) {
	query := c.Query("q")
	c.JSON(200, gin.H{"success": true, "data": gin.H{"query": query, "results": []interface{}{}}})
}

func searchUsers(c *gin.Context) {
	c.JSON(200, gin.H{"success": true, "data": gin.H{"results": []interface{}{}}})
}

func searchFiles(c *gin.Context) {
	c.JSON(200, gin.H{"success": true, "data": gin.H{"results": []interface{}{}}})
}

func searchChannels(c *gin.Context) {
	c.JSON(200, gin.H{"success": true, "data": gin.H{"results": []interface{}{}}})
}

func searchAll(c *gin.Context) {
	c.JSON(200, gin.H{"success": true, "data": gin.H{"messages": []interface{}{}, "users": []interface{}{}, "files": []interface{}{}, "channels": []interface{}{}}})
}

func indexDocument(c *gin.Context) {
	c.JSON(200, gin.H{"success": true, "message": "Document indexed"})
}

func removeFromIndex(c *gin.Context) {
	c.JSON(200, gin.H{"success": true, "message": "Document removed from index"})
}

func reindex(c *gin.Context) {
	c.JSON(200, gin.H{"success": true, "message": "Reindex started"})
}
