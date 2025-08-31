package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-catalog-api/internal/config"
	"ecommerce-catalog-api/internal/database"
	"ecommerce-catalog-api/internal/elasticsearch"
	"ecommerce-catalog-api/internal/handlers"
	"ecommerce-catalog-api/internal/repository"
	"ecommerce-catalog-api/internal/services"
	"ecommerce-catalog-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Initialize logger
	logger.Initialize(cfg.Logger.Level)
	logrus.Info("Starting E-commerce Catalog API")

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize database connection
	db, err := database.Initialize(&cfg.Database)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize database")
	}
	defer db.Close()

	// Initialize Elasticsearch client
	esClient, err := elasticsearch.Initialize(&cfg.Elasticsearch)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize Elasticsearch")
	}

	// Initialize repositories
	productRepo := repository.NewProductRepository(db.DB)

	// Initialize services (CQRS pattern)
	commandService := services.NewProductCommandService(productRepo, esClient)
	queryService := services.NewProductQueryService(esClient, productRepo)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(commandService, queryService)

	// Setup routes
	router := setupRoutes(productHandler)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logrus.WithField("port", cfg.Server.Port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("Server forced to shutdown")
	}

	logrus.Info("Server exited")
}

func setupRoutes(productHandler *handlers.ProductHandler) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(requestIDMiddleware())

	// Health check endpoint
	router.GET("/health", productHandler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Product CRUD operations (Commands)
		v1.POST("/products", productHandler.CreateProduct)
		v1.GET("/products/:id", productHandler.GetProduct)
		v1.PUT("/products/:id", productHandler.UpdateProduct)
		v1.DELETE("/products/:id", productHandler.DeleteProduct)
		v1.GET("/products", productHandler.ListProducts)

		// Search operations (Queries)
		v1.GET("/search", productHandler.SearchProducts)
		v1.GET("/search/postgresql", productHandler.SearchProductsPostgreSQL)
	}

	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// requestIDMiddleware adds a unique request ID to each request
func requestIDMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	})
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}