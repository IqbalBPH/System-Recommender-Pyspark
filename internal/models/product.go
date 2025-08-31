package models

import (
	"time"
)

// Product represents a product in the e-commerce catalog
type Product struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name" binding:"required"`
	Description  string    `json:"description" db:"description"`
	Brand        string    `json:"brand" db:"brand" binding:"required"`
	Category     string    `json:"category" db:"category" binding:"required"`
	Price        float64   `json:"price" db:"price" binding:"required"`
	StockQuantity int       `json:"stock_quantity" db:"stock_quantity" binding:"required"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ProductCreate represents the request body for creating a product
type ProductCreate struct {
	Name         string  `json:"name" binding:"required,min=1,max=255"`
	Description  string  `json:"description" binding:"max=1000"`
	Brand        string  `json:"brand" binding:"required,min=1,max=100"`
	Category     string  `json:"category" binding:"required,min=1,max=100"`
	Price        float64 `json:"price" binding:"required,gt=0"`
	StockQuantity int     `json:"stock_quantity" binding:"required,gte=0"`
}

// ProductUpdate represents the request body for updating a product
type ProductUpdate struct {
	Name         *string  `json:"name" binding:"omitempty,min=1,max=255"`
	Description  *string  `json:"description" binding:"omitempty,max=1000"`
	Brand        *string  `json:"brand" binding:"omitempty,min=1,max=100"`
	Category     *string  `json:"category" binding:"omitempty,min=1,max=100"`
	Price        *float64 `json:"price" binding:"omitempty,gt=0"`
	StockQuantity *int     `json:"stock_quantity" binding:"omitempty,gte=0"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query    string             `json:"query" form:"query"`
	Brand    []string           `json:"brand" form:"brand"`
	Category []string           `json:"category" form:"category"`
	Price    *PriceRangeFilter  `json:"price" form:"price"`
	Page     int                `json:"page" form:"page"`
	Size     int                `json:"size" form:"size"`
	SortBy   string             `json:"sort_by" form:"sort_by"`
	SortOrder string            `json:"sort_order" form:"sort_order"`
}

// PriceRangeFilter represents price filtering
type PriceRangeFilter struct {
	Min *float64 `json:"min" form:"min"`
	Max *float64 `json:"max" form:"max"`
}

// SearchResponse represents the search response
type SearchResponse struct {
	Products     []Product           `json:"products"`
	Total        int64               `json:"total"`
	Page         int                 `json:"page"`
	Size         int                 `json:"size"`
	Aggregations map[string]interface{} `json:"aggregations"`
}

// BrandAggregation represents brand aggregation result
type BrandAggregation struct {
	Brand string `json:"brand"`
	Count int64  `json:"count"`
}

// CategoryAggregation represents category aggregation result
type CategoryAggregation struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// PerformanceMetrics represents performance test results
type PerformanceMetrics struct {
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime   time.Duration `json:"min_response_time"`
	MaxResponseTime   time.Duration `json:"max_response_time"`
	RequestsPerSecond float64       `json:"requests_per_second"`
	TestDuration      time.Duration `json:"test_duration"`
}