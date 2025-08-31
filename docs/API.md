# API Documentation

## Overview

The E-commerce Product Catalog API is a RESTful service implementing CQRS (Command Query Responsibility Segregation) architecture. All write operations go to PostgreSQL (source of truth), while all read operations are served by Elasticsearch for optimal search performance.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Currently, the API does not require authentication. In a production environment, you would typically implement JWT or API key authentication.

## Response Format

All API responses follow a consistent format:

### Success Response
```json
{
  "success": true,
  "data": {...},
  "message": "Optional success message"
}
```

### Error Response
```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "details": {
    "additional": "error details"
  }
}
```

## Product Management Endpoints (Commands)

### Create Product

Creates a new product in the catalog.

**Endpoint:** `POST /api/v1/products`

**Request Body:**
```json
{
  "name": "Apple MacBook Pro 16-inch",
  "description": "Professional laptop with M3 Pro chip",
  "brand": "Apple",
  "category": "Computers",
  "price": 2499.99,
  "stock_quantity": 15
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": 123,
    "name": "Apple MacBook Pro 16-inch",
    "description": "Professional laptop with M3 Pro chip",
    "brand": "Apple",
    "category": "Computers",
    "price": 2499.99,
    "stock_quantity": 15,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "message": "Product created successfully"
}
```

### Get Product by ID

Retrieves a specific product by its ID.

**Endpoint:** `GET /api/v1/products/{id}`

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": 123,
    "name": "Apple MacBook Pro 16-inch",
    "description": "Professional laptop with M3 Pro chip",
    "brand": "Apple",
    "category": "Computers",
    "price": 2499.99,
    "stock_quantity": 15,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### Update Product

Updates an existing product. All fields are optional.

**Endpoint:** `PUT /api/v1/products/{id}`

**Request Body:**
```json
{
  "name": "Apple MacBook Pro 16-inch (Updated)",
  "price": 2399.99,
  "stock_quantity": 20
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": 123,
    "name": "Apple MacBook Pro 16-inch (Updated)",
    "description": "Professional laptop with M3 Pro chip",
    "brand": "Apple",
    "category": "Computers",
    "price": 2399.99,
    "stock_quantity": 20,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T11:45:00Z"
  },
  "message": "Product updated successfully"
}
```

### Delete Product

Removes a product from the catalog.

**Endpoint:** `DELETE /api/v1/products/{id}`

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Product deleted successfully"
}
```

### List Products

Retrieves a paginated list of all products.

**Endpoint:** `GET /api/v1/products`

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `size` (optional): Number of items per page (default: 20, max: 100)

**Example:** `GET /api/v1/products?page=2&size=10`

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "products": [
      {
        "id": 123,
        "name": "Apple MacBook Pro 16-inch",
        "description": "Professional laptop with M3 Pro chip",
        "brand": "Apple",
        "category": "Computers",
        "price": 2499.99,
        "stock_quantity": 15,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 1250,
    "page": 2,
    "size": 10
  }
}
```

## Search Endpoints (Queries)

### Search Products (Elasticsearch)

Performs advanced search using Elasticsearch with full-text search, filtering, and aggregations.

**Endpoint:** `GET /api/v1/search`

**Query Parameters:**
- `query` (optional): Text search query
- `brand` (optional): Filter by brand(s) - can be specified multiple times
- `category` (optional): Filter by category(s) - can be specified multiple times
- `price_min` (optional): Minimum price filter
- `price_max` (optional): Maximum price filter
- `sort_by` (optional): Sort field (name, price, created_at, etc.)
- `sort_order` (optional): Sort order (asc, desc)
- `page` (optional): Page number (default: 1)
- `size` (optional): Number of items per page (default: 20, max: 100)

**Examples:**

Simple text search:
```
GET /api/v1/search?query=laptop
```

Filtered search:
```
GET /api/v1/search?query=laptop&brand=Apple&brand=Dell&price_min=1000&price_max=3000
```

Complex search with sorting:
```
GET /api/v1/search?query=wireless+headphones&category=Audio&sort_by=price&sort_order=asc&page=1&size=20
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "products": [
      {
        "id": 123,
        "name": "Apple MacBook Pro 16-inch",
        "description": "Professional laptop with M3 Pro chip",
        "brand": "Apple",
        "category": "Computers",
        "price": 2499.99,
        "stock_quantity": 15,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 45,
    "page": 1,
    "size": 20,
    "aggregations": {
      "brands": {
        "buckets": [
          {
            "key": "Apple",
            "doc_count": 12
          },
          {
            "key": "Dell",
            "doc_count": 8
          }
        ]
      },
      "categories": {
        "buckets": [
          {
            "key": "Computers",
            "doc_count": 15
          },
          {
            "key": "Electronics",
            "doc_count": 10
          }
        ]
      }
    }
  }
}
```

### Search Products (PostgreSQL)

Performs search using PostgreSQL for performance comparison purposes.

**Endpoint:** `GET /api/v1/search/postgresql`

**Query Parameters:** Same as Elasticsearch search endpoint

**Response:** Same format as Elasticsearch search, but with limited aggregation capabilities

**Headers:**
- `X-Search-Engine: PostgreSQL` - Indicates the search was performed using PostgreSQL
- `X-Response-Time: 45ms` - Response time for performance monitoring

## System Endpoints

### Health Check

Checks if the API service is running and healthy.

**Endpoint:** `GET /health`

**Response:** `200 OK`
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "ecommerce-catalog-api"
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `validation_error` | 400 | Request validation failed |
| `invalid_parameter` | 400 | Invalid parameter provided |
| `not_found` | 404 | Resource not found |
| `internal_error` | 500 | Internal server error |

## Rate Limiting

Currently, no rate limiting is implemented. In production, you should implement rate limiting to prevent abuse.

## Response Headers

The API includes several helpful headers:

- `X-Request-ID` - Unique identifier for request tracking
- `X-Response-Time` - Response time in milliseconds
- `X-Search-Engine` - Search engine used (Elasticsearch/PostgreSQL)

## Elasticsearch Query Examples

### 1. Full-text Search
```
GET /api/v1/search?query=wireless+bluetooth+headphones
```
Uses Elasticsearch's `multi_match` query on name and description fields.

### 2. Exact Brand Filter
```
GET /api/v1/search?brand=Apple&brand=Samsung
```
Uses Elasticsearch's `terms` query for exact matching.

### 3. Price Range Filter
```
GET /api/v1/search?price_min=100&price_max=500
```
Uses Elasticsearch's `range` query.

### 4. Complex Search with Aggregations
```
GET /api/v1/search?query=laptop&category=Computers&sort_by=price&sort_order=desc
```
Returns results with brand and category aggregations for faceted search.

## Performance Monitoring

All search endpoints include response time in headers. Use the performance test tool to compare Elasticsearch vs PostgreSQL:

```bash
./performance-test -users=50 -duration=60s -type=both
```

This will show you the performance benefits of using Elasticsearch for search operations.

## CQRS Implementation

The API strictly follows CQRS principles:

**Commands (Writes):**
- Product creation, updates, and deletions go to PostgreSQL first
- After successful PostgreSQL operation, data is synchronized to Elasticsearch
- Ensures PostgreSQL remains the authoritative source of truth

**Queries (Reads):**
- All search operations are served exclusively by Elasticsearch
- Provides fast full-text search, filtering, and aggregations
- No read operations touch PostgreSQL in the query path

This architecture ensures optimal performance for both write consistency and read scalability.