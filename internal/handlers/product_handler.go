package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ecommerce-catalog-api/internal/models"
	"ecommerce-catalog-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	commandService services.ProductCommandServiceInterface
	queryService   services.ProductQueryServiceInterface
}

// NewProductHandler creates a new product handler
func NewProductHandler(commandService services.ProductCommandServiceInterface, queryService services.ProductQueryServiceInterface) *ProductHandler {
	return &ProductHandler{
		commandService: commandService,
		queryService:   queryService,
	}
}

// CreateProduct handles product creation (POST /api/v1/products)
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var productCreate models.ProductCreate
	if err := c.ShouldBindJSON(&productCreate); err != nil {
		logrus.WithError(err).Error("Invalid request body for product creation")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	product, err := h.commandService.CreateProduct(&productCreate)
	if err != nil {
		logrus.WithError(err).Error("Failed to create product")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create product",
		})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Success: true,
		Data:    product,
		Message: "Product created successfully",
	})
}

// GetProduct handles getting a product by ID (GET /api/v1/products/:id)
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_parameter",
			Message: "Invalid product ID",
		})
		return
	}

	product, err := h.queryService.GetProductByID(id)
	if err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to get product")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get product",
		})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    product,
	})
}

// UpdateProduct handles product updates (PUT /api/v1/products/:id)
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_parameter",
			Message: "Invalid product ID",
		})
		return
	}

	var productUpdate models.ProductUpdate
	if err := c.ShouldBindJSON(&productUpdate); err != nil {
		logrus.WithError(err).Error("Invalid request body for product update")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	product, err := h.commandService.UpdateProduct(id, &productUpdate)
	if err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to update product")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update product",
		})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    product,
		Message: "Product updated successfully",
	})
}

// DeleteProduct handles product deletion (DELETE /api/v1/products/:id)
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_parameter",
			Message: "Invalid product ID",
		})
		return
	}

	if err := h.commandService.DeleteProduct(id); err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to delete product")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete product",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Product deleted successfully",
	})
}

// ListProducts handles listing products with pagination (GET /api/v1/products)
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	products, total, err := h.queryService.ListProducts(page, size)
	if err != nil {
		logrus.WithError(err).Error("Failed to list products")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list products",
		})
		return
	}

	response := map[string]interface{}{
		"products": products,
		"total":    total,
		"page":     page,
		"size":     size,
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// SearchProducts handles product search via Elasticsearch (GET /api/v1/search)
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	var searchReq models.SearchRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&searchReq); err != nil {
		logrus.WithError(err).Error("Invalid search parameters")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid search parameters",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// Parse price range if provided
	if minPriceStr := c.Query("price_min"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			if searchReq.Price == nil {
				searchReq.Price = &models.PriceRangeFilter{}
			}
			searchReq.Price.Min = &minPrice
		}
	}

	if maxPriceStr := c.Query("price_max"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			if searchReq.Price == nil {
				searchReq.Price = &models.PriceRangeFilter{}
			}
			searchReq.Price.Max = &maxPrice
		}
	}

	startTime := time.Now()
	response, err := h.queryService.SearchProducts(&searchReq)
	duration := time.Since(startTime)

	if err != nil {
		logrus.WithError(err).Error("Failed to search products")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to search products",
		})
		return
	}

	// Add response time to headers for monitoring
	c.Header("X-Response-Time", duration.String())

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// SearchProductsPostgreSQL handles product search via PostgreSQL (GET /api/v1/search/postgresql)
func (h *ProductHandler) SearchProductsPostgreSQL(c *gin.Context) {
	var searchReq models.SearchRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&searchReq); err != nil {
		logrus.WithError(err).Error("Invalid search parameters")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid search parameters",
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// Parse price range if provided
	if minPriceStr := c.Query("price_min"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			if searchReq.Price == nil {
				searchReq.Price = &models.PriceRangeFilter{}
			}
			searchReq.Price.Min = &minPrice
		}
	}

	if maxPriceStr := c.Query("price_max"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			if searchReq.Price == nil {
				searchReq.Price = &models.PriceRangeFilter{}
			}
			searchReq.Price.Max = &maxPrice
		}
	}

	startTime := time.Now()
	response, err := h.queryService.SearchProductsByPostgreSQL(&searchReq)
	duration := time.Since(startTime)

	if err != nil {
		logrus.WithError(err).Error("Failed to search products via PostgreSQL")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to search products",
		})
		return
	}

	// Add response time to headers for monitoring
	c.Header("X-Response-Time", duration.String())
	c.Header("X-Search-Engine", "PostgreSQL")

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// HealthCheck handles health check endpoint (GET /health)
func (h *ProductHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "ecommerce-catalog-api",
	})
}