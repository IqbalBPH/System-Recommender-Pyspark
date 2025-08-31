package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-catalog-api/internal/handlers"
	"ecommerce-catalog-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock services for testing
type mockCommandService struct{}
type mockQueryService struct{}

func (m *mockCommandService) CreateProduct(product *models.ProductCreate) (*models.Product, error) {
	return &models.Product{
		ID:           1,
		Name:         product.Name,
		Description:  product.Description,
		Brand:        product.Brand,
		Category:     product.Category,
		Price:        product.Price,
		StockQuantity: product.StockQuantity,
	}, nil
}

func (m *mockCommandService) UpdateProduct(id int, update *models.ProductUpdate) (*models.Product, error) {
	return &models.Product{
		ID:           id,
		Name:         "Updated Product",
		Description:  "Updated Description",
		Brand:        "Updated Brand",
		Category:     "Updated Category",
		Price:        99.99,
		StockQuantity: 10,
	}, nil
}

func (m *mockCommandService) DeleteProduct(id int) error {
	return nil
}

func (m *mockQueryService) GetProductByID(id int) (*models.Product, error) {
	return &models.Product{
		ID:           id,
		Name:         "Test Product",
		Description:  "Test Description",
		Brand:        "Test Brand",
		Category:     "Test Category",
		Price:        99.99,
		StockQuantity: 10,
	}, nil
}

func (m *mockQueryService) SearchProducts(req *models.SearchRequest) (*models.SearchResponse, error) {
	return &models.SearchResponse{
		Products: []models.Product{
			{
				ID:           1,
				Name:         "Test Product",
				Description:  "Test Description",
				Brand:        "Test Brand",
				Category:     "Test Category",
				Price:        99.99,
				StockQuantity: 10,
			},
		},
		Total: 1,
		Page:  1,
		Size:  10,
	}, nil
}

func (m *mockQueryService) SearchProductsByPostgreSQL(req *models.SearchRequest) (*models.SearchResponse, error) {
	return m.SearchProducts(req)
}

func (m *mockQueryService) ListProducts(page, size int) ([]models.Product, int64, error) {
	return []models.Product{
		{
			ID:           1,
			Name:         "Test Product",
			Description:  "Test Description",
			Brand:        "Test Brand",
			Category:     "Test Category",
			Price:        99.99,
			StockQuantity: 10,
		},
	}, 1, nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	
	commandService := &mockCommandService{}
	queryService := &mockQueryService{}
	handler := handlers.NewProductHandler(commandService, queryService)
	
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		v1.POST("/products", handler.CreateProduct)
		v1.GET("/products/:id", handler.GetProduct)
		v1.PUT("/products/:id", handler.UpdateProduct)
		v1.DELETE("/products/:id", handler.DeleteProduct)
		v1.GET("/products", handler.ListProducts)
		v1.GET("/search", handler.SearchProducts)
	}
	
	return router
}

func TestCreateProduct(t *testing.T) {
	router := setupTestRouter()
	
	product := models.ProductCreate{
		Name:         "Test Product",
		Description:  "Test Description",
		Brand:        "Test Brand",
		Category:     "Test Category",
		Price:        99.99,
		StockQuantity: 10,
	}
	
	jsonData, _ := json.Marshal(product)
	
	req, _ := http.NewRequest("POST", "/api/v1/products", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response models.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

func TestGetProduct(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/v1/products/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

func TestSearchProducts(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/v1/search?query=test&page=1&size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

func TestInvalidProductID(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/v1/products/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_parameter", response.Error)
}