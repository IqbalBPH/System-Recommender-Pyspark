package services

import (
	"fmt"

	"ecommerce-catalog-api/internal/elasticsearch"
	"ecommerce-catalog-api/internal/models"
	"ecommerce-catalog-api/internal/repository"
	"github.com/sirupsen/logrus"
)

// ProductCommandServiceInterface defines the command service interface
type ProductCommandServiceInterface interface {
	CreateProduct(productCreate *models.ProductCreate) (*models.Product, error)
	UpdateProduct(id int, productUpdate *models.ProductUpdate) (*models.Product, error)
	DeleteProduct(id int) error
}

// ProductQueryServiceInterface defines the query service interface
type ProductQueryServiceInterface interface {
	SearchProducts(searchReq *models.SearchRequest) (*models.SearchResponse, error)
	SearchProductsByPostgreSQL(searchReq *models.SearchRequest) (*models.SearchResponse, error)
	GetProductByID(id int) (*models.Product, error)
	ListProducts(page, size int) ([]models.Product, int64, error)
}

// ProductCommandService handles write operations (CQRS Command side)
type ProductCommandService struct {
	productRepo *repository.ProductRepository
	esClient    *elasticsearch.Client
}

// NewProductCommandService creates a new command service
func NewProductCommandService(productRepo *repository.ProductRepository, esClient *elasticsearch.Client) *ProductCommandService {
	return &ProductCommandService{
		productRepo: productRepo,
		esClient:    esClient,
	}
}

// CreateProduct creates a new product (Command)
func (s *ProductCommandService) CreateProduct(productCreate *models.ProductCreate) (*models.Product, error) {
	logrus.WithFields(logrus.Fields{
		"name":     productCreate.Name,
		"brand":    productCreate.Brand,
		"category": productCreate.Category,
	}).Info("Creating new product")

	// Step 1: Save to PostgreSQL (source of truth)
	product, err := s.productRepo.Create(productCreate)
	if err != nil {
		logrus.WithError(err).Error("Failed to create product in PostgreSQL")
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Step 2: Index in Elasticsearch
	if err := s.esClient.IndexProduct(product); err != nil {
		logrus.WithError(err).WithField("product_id", product.ID).Error("Failed to index product in Elasticsearch")
		// Note: In a production system, you might want to implement a retry mechanism
		// or a separate process to handle failed Elasticsearch operations
	}

	logrus.WithField("product_id", product.ID).Info("Product created and indexed successfully")
	return product, nil
}

// UpdateProduct updates an existing product (Command)
func (s *ProductCommandService) UpdateProduct(id int, productUpdate *models.ProductUpdate) (*models.Product, error) {
	logrus.WithField("product_id", id).Info("Updating product")

	// Step 1: Update in PostgreSQL (source of truth)
	product, err := s.productRepo.Update(id, productUpdate)
	if err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to update product in PostgreSQL")
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	if product == nil {
		return nil, nil // Product not found
	}

	// Step 2: Update in Elasticsearch
	if err := s.esClient.IndexProduct(product); err != nil {
		logrus.WithError(err).WithField("product_id", product.ID).Error("Failed to update product in Elasticsearch")
		// Note: In a production system, you might want to implement a retry mechanism
	}

	logrus.WithField("product_id", product.ID).Info("Product updated and reindexed successfully")
	return product, nil
}

// DeleteProduct deletes a product (Command)
func (s *ProductCommandService) DeleteProduct(id int) error {
	logrus.WithField("product_id", id).Info("Deleting product")

	// Step 1: Delete from PostgreSQL (source of truth)
	if err := s.productRepo.Delete(id); err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to delete product from PostgreSQL")
		return fmt.Errorf("failed to delete product: %w", err)
	}

	// Step 2: Delete from Elasticsearch
	if err := s.esClient.DeleteProduct(id); err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to delete product from Elasticsearch")
		// Note: In a production system, you might want to implement a retry mechanism
	}

	logrus.WithField("product_id", id).Info("Product deleted successfully")
	return nil
}

// ProductQueryService handles read operations (CQRS Query side)
type ProductQueryService struct {
	esClient    *elasticsearch.Client
	productRepo *repository.ProductRepository // For PostgreSQL performance comparison
}

// NewProductQueryService creates a new query service
func NewProductQueryService(esClient *elasticsearch.Client, productRepo *repository.ProductRepository) *ProductQueryService {
	return &ProductQueryService{
		esClient:    esClient,
		productRepo: productRepo,
	}
}

// SearchProducts searches for products using Elasticsearch (Query)
func (s *ProductQueryService) SearchProducts(searchReq *models.SearchRequest) (*models.SearchResponse, error) {
	logrus.WithFields(logrus.Fields{
		"query":    searchReq.Query,
		"brand":    searchReq.Brand,
		"category": searchReq.Category,
		"page":     searchReq.Page,
		"size":     searchReq.Size,
	}).Info("Searching products via Elasticsearch")

	response, err := s.esClient.SearchProducts(searchReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to search products in Elasticsearch")
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"total_results": response.Total,
		"returned":      len(response.Products),
	}).Info("Search completed successfully")

	return response, nil
}

// SearchProductsByPostgreSQL searches for products using PostgreSQL (for performance comparison)
func (s *ProductQueryService) SearchProductsByPostgreSQL(searchReq *models.SearchRequest) (*models.SearchResponse, error) {
	logrus.WithFields(logrus.Fields{
		"query":    searchReq.Query,
		"brand":    searchReq.Brand,
		"category": searchReq.Category,
		"page":     searchReq.Page,
		"size":     searchReq.Size,
	}).Info("Searching products via PostgreSQL")

	response, err := s.productRepo.SearchByPostgreSQL(searchReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to search products in PostgreSQL")
		return nil, fmt.Errorf("failed to search products in PostgreSQL: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"total_results": response.Total,
		"returned":      len(response.Products),
	}).Info("PostgreSQL search completed successfully")

	return response, nil
}

// GetProductByID gets a product by ID from PostgreSQL (for administrative purposes)
func (s *ProductQueryService) GetProductByID(id int) (*models.Product, error) {
	logrus.WithField("product_id", id).Info("Getting product by ID")

	product, err := s.productRepo.GetByID(id)
	if err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to get product by ID")
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// ListProducts lists all products with pagination from PostgreSQL
func (s *ProductQueryService) ListProducts(page, size int) ([]models.Product, int64, error) {
	logrus.WithFields(logrus.Fields{
		"page": page,
		"size": size,
	}).Info("Listing products")

	products, total, err := s.productRepo.List(page, size)
	if err != nil {
		logrus.WithError(err).Error("Failed to list products")
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"total":    total,
		"returned": len(products),
	}).Info("Products listed successfully")

	return products, total, nil
}