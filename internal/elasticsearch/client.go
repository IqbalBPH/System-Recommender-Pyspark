package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ecommerce-catalog-api/internal/config"
	"ecommerce-catalog-api/internal/models"
	
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

const ProductsIndex = "products"

// Client wraps the Elasticsearch client
type Client struct {
	es *elasticsearch.Client
}

// Initialize creates a new Elasticsearch client
func Initialize(cfg *config.ElasticsearchConfig) (*Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.URL},
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	client := &Client{es: es}

	// Test connection
	if err := client.Health(); err != nil {
		return nil, fmt.Errorf("elasticsearch health check failed: %w", err)
	}

	// Create index and mapping
	if err := client.CreateProductsIndex(); err != nil {
		return nil, fmt.Errorf("failed to create products index: %w", err)
	}

	logrus.Info("Elasticsearch connection established successfully")

	return client, nil
}

// Health checks if Elasticsearch is healthy
func (c *Client) Health() error {
	res, err := c.es.Cluster.Health()
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch health check failed: %s", res.Status())
	}

	return nil
}

// CreateProductsIndex creates the products index with proper mapping
func (c *Client) CreateProductsIndex() error {
	mapping := `{
		"mappings": {
			"properties": {
				"id": {"type": "integer"},
				"name": {
					"type": "text",
					"analyzer": "standard",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"description": {
					"type": "text",
					"analyzer": "standard"
				},
				"brand": {
					"type": "keyword"
				},
				"category": {
					"type": "keyword"
				},
				"price": {
					"type": "float"
				},
				"stock_quantity": {
					"type": "integer"
				},
				"created_at": {
					"type": "date",
					"format": "strict_date_optional_time||epoch_millis"
				},
				"updated_at": {
					"type": "date",
					"format": "strict_date_optional_time||epoch_millis"
				}
			}
		},
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"standard": {
						"type": "standard",
						"stopwords": "_english_"
					}
				}
			}
		}
	}`

	req := esapi.IndicesCreateRequest{
		Index: ProductsIndex,
		Body:  strings.NewReader(mapping),
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 400 { // 400 means index already exists
		return fmt.Errorf("failed to create index: %s", res.Status())
	}

	logrus.Info("Products index created or already exists")
	return nil
}

// IndexProduct indexes a product in Elasticsearch
func (c *Client) IndexProduct(product *models.Product) error {
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      ProductsIndex,
		DocumentID: fmt.Sprintf("%d", product.ID),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index product: %s", res.Status())
	}

	return nil
}

// DeleteProduct removes a product from Elasticsearch
func (c *Client) DeleteProduct(productID int) error {
	req := esapi.DeleteRequest{
		Index:      ProductsIndex,
		DocumentID: fmt.Sprintf("%d", productID),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("failed to delete product: %s", res.Status())
	}

	return nil
}

// SearchProducts performs a complex search with filters and aggregations
func (c *Client) SearchProducts(searchReq *models.SearchRequest) (*models.SearchResponse, error) {
	query := c.buildSearchQuery(searchReq)
	
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	logrus.WithField("query", string(queryBytes)).Debug("Elasticsearch query")

	req := esapi.SearchRequest{
		Index: []string{ProductsIndex},
		Body:  bytes.NewReader(queryBytes),
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search failed: %s", res.Status())
	}

	return c.parseSearchResponse(res, searchReq)
}

// buildSearchQuery builds the Elasticsearch query from search request
func (c *Client) buildSearchQuery(searchReq *models.SearchRequest) map[string]interface{} {
	// Set defaults
	if searchReq.Page < 1 {
		searchReq.Page = 1
	}
	if searchReq.Size < 1 || searchReq.Size > 100 {
		searchReq.Size = 20
	}

	from := (searchReq.Page - 1) * searchReq.Size

	query := map[string]interface{}{
		"from": from,
		"size": searchReq.Size,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{},
				"filter": []interface{}{},
			},
		},
		"aggs": map[string]interface{}{
			"brands": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "brand",
					"size":  20,
				},
			},
			"categories": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "category",
					"size":  20,
				},
			},
		},
	}

	// Add text search
	if searchReq.Query != "" {
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]interface{}),
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  searchReq.Query,
					"fields": []string{"name^2", "description"},
					"type":   "best_fields",
				},
			},
		)
	} else {
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
			query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]interface{}),
			map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		)
	}

	// Add brand filter
	if len(searchReq.Brand) > 0 {
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			query["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]interface{}),
			map[string]interface{}{
				"terms": map[string]interface{}{
					"brand": searchReq.Brand,
				},
			},
		)
	}

	// Add category filter
	if len(searchReq.Category) > 0 {
		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			query["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]interface{}),
			map[string]interface{}{
				"terms": map[string]interface{}{
					"category": searchReq.Category,
				},
			},
		)
	}

	// Add price range filter
	if searchReq.Price != nil && (searchReq.Price.Min != nil || searchReq.Price.Max != nil) {
		priceRange := map[string]interface{}{}
		if searchReq.Price.Min != nil {
			priceRange["gte"] = *searchReq.Price.Min
		}
		if searchReq.Price.Max != nil {
			priceRange["lte"] = *searchReq.Price.Max
		}

		query["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"] = append(
			query["query"].(map[string]interface{})["bool"].(map[string]interface{})["filter"].([]interface{}),
			map[string]interface{}{
				"range": map[string]interface{}{
					"price": priceRange,
				},
			},
		)
	}

	// Add sorting
	if searchReq.SortBy != "" {
		sortOrder := "asc"
		if searchReq.SortOrder == "desc" {
			sortOrder = "desc"
		}

		sortField := searchReq.SortBy
		if searchReq.SortBy == "name" {
			sortField = "name.keyword"
		}

		query["sort"] = []interface{}{
			map[string]interface{}{
				sortField: map[string]interface{}{
					"order": sortOrder,
				},
			},
		}
	}

	return query
}

// parseSearchResponse parses the Elasticsearch response
func (c *Client) parseSearchResponse(res *esapi.Response, searchReq *models.SearchRequest) (*models.SearchResponse, error) {
	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	hits := response["hits"].(map[string]interface{})
	total := int64(hits["total"].(map[string]interface{})["value"].(float64))

	var products []models.Product
	for _, hit := range hits["hits"].([]interface{}) {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})

		var product models.Product
		productBytes, _ := json.Marshal(source)
		if err := json.Unmarshal(productBytes, &product); err != nil {
			logrus.WithError(err).Warn("Failed to parse product from search result")
			continue
		}

		products = append(products, product)
	}

	// Parse aggregations
	aggregations := make(map[string]interface{})
	if aggs, exists := response["aggregations"]; exists {
		aggregations = aggs.(map[string]interface{})
	}

	return &models.SearchResponse{
		Products:     products,
		Total:        total,
		Page:         searchReq.Page,
		Size:         searchReq.Size,
		Aggregations: aggregations,
	}, nil
}

// BulkIndexProducts indexes multiple products at once
func (c *Client) BulkIndexProducts(products []models.Product) error {
	if len(products) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, product := range products {
		// Index operation metadata
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": ProductsIndex,
				"_id":    fmt.Sprintf("%d", product.ID),
			},
		}
		metaBytes, _ := json.Marshal(meta)
		buf.Write(metaBytes)
		buf.WriteByte('\n')

		// Document data
		productBytes, _ := json.Marshal(product)
		buf.Write(productBytes)
		buf.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index failed: %s", res.Status())
	}

	logrus.WithField("count", len(products)).Info("Bulk indexed products")
	return nil
}