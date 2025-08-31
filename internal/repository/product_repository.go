package repository

import (
	"database/sql"
	"time"

	"ecommerce-catalog-api/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// ProductRepository handles database operations for products
type ProductRepository struct {
	db *sqlx.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create inserts a new product into the database
func (r *ProductRepository) Create(product *models.ProductCreate) (*models.Product, error) {
	query := `
		INSERT INTO products (name, description, brand, category, price, stock_quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, name, description, brand, category, price, stock_quantity, created_at, updated_at
	`

	var result models.Product
	err := r.db.QueryRowx(
		query,
		product.Name,
		product.Description,
		product.Brand,
		product.Category,
		product.Price,
		product.StockQuantity,
	).StructScan(&result)

	if err != nil {
		logrus.WithError(err).Error("Failed to create product")
		return nil, err
	}

	logrus.WithField("product_id", result.ID).Info("Product created successfully")
	return &result, nil
}

// GetByID retrieves a product by its ID
func (r *ProductRepository) GetByID(id int) (*models.Product, error) {
	query := `
		SELECT id, name, description, brand, category, price, stock_quantity, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product models.Product
	err := r.db.QueryRowx(query, id).StructScan(&product)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logrus.WithError(err).WithField("product_id", id).Error("Failed to get product by ID")
		return nil, err
	}

	return &product, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(id int, update *models.ProductUpdate) (*models.Product, error) {
	// First, get the current product
	current, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}

	// Apply updates
	if update.Name != nil {
		current.Name = *update.Name
	}
	if update.Description != nil {
		current.Description = *update.Description
	}
	if update.Brand != nil {
		current.Brand = *update.Brand
	}
	if update.Category != nil {
		current.Category = *update.Category
	}
	if update.Price != nil {
		current.Price = *update.Price
	}
	if update.StockQuantity != nil {
		current.StockQuantity = *update.StockQuantity
	}
	current.UpdatedAt = time.Now()

	query := `
		UPDATE products
		SET name = $2, description = $3, brand = $4, category = $5, price = $6, stock_quantity = $7, updated_at = $8
		WHERE id = $1
		RETURNING id, name, description, brand, category, price, stock_quantity, created_at, updated_at
	`

	var result models.Product
	err = r.db.QueryRowx(
		query,
		id,
		current.Name,
		current.Description,
		current.Brand,
		current.Category,
		current.Price,
		current.StockQuantity,
		current.UpdatedAt,
	).StructScan(&result)

	if err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to update product")
		return nil, err
	}

	logrus.WithField("product_id", result.ID).Info("Product updated successfully")
	return &result, nil
}

// Delete removes a product from the database
func (r *ProductRepository) Delete(id int) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to delete product")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return nil // Product not found
	}

	logrus.WithField("product_id", id).Info("Product deleted successfully")
	return nil
}

// List retrieves products with pagination
func (r *ProductRepository) List(page, size int) ([]models.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	offset := (page - 1) * size

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM products`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Failed to count products")
		return nil, 0, err
	}

	// Get products
	query := `
		SELECT id, name, description, brand, category, price, stock_quantity, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var products []models.Product
	err = r.db.Select(&products, query, size, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list products")
		return nil, 0, err
	}

	return products, total, nil
}

// GetAll retrieves all products (for migration purposes)
func (r *ProductRepository) GetAll() ([]models.Product, error) {
	query := `
		SELECT id, name, description, brand, category, price, stock_quantity, created_at, updated_at
		FROM products
		ORDER BY id
	`

	var products []models.Product
	err := r.db.Select(&products, query)
	if err != nil {
		logrus.WithError(err).Error("Failed to get all products")
		return nil, err
	}

	return products, nil
}

// SearchByPostgreSQL performs PostgreSQL-based search (for performance comparison)
func (r *ProductRepository) SearchByPostgreSQL(searchReq *models.SearchRequest) (*models.SearchResponse, error) {
	if searchReq.Page < 1 {
		searchReq.Page = 1
	}
	if searchReq.Size < 1 || searchReq.Size > 100 {
		searchReq.Size = 20
	}

	offset := (searchReq.Page - 1) * searchReq.Size

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if searchReq.Query != "" {
		whereClause += " AND (name ILIKE $" + string(rune(argIndex)) + " OR description ILIKE $" + string(rune(argIndex)) + ")"
		args = append(args, "%"+searchReq.Query+"%")
		argIndex++
	}

	if len(searchReq.Brand) > 0 {
		whereClause += " AND brand = ANY($" + string(rune(argIndex)) + ")"
		args = append(args, searchReq.Brand)
		argIndex++
	}

	if len(searchReq.Category) > 0 {
		whereClause += " AND category = ANY($" + string(rune(argIndex)) + ")"
		args = append(args, searchReq.Category)
		argIndex++
	}

	if searchReq.Price != nil {
		if searchReq.Price.Min != nil {
			whereClause += " AND price >= $" + string(rune(argIndex))
			args = append(args, *searchReq.Price.Min)
			argIndex++
		}
		if searchReq.Price.Max != nil {
			whereClause += " AND price <= $" + string(rune(argIndex))
			args = append(args, *searchReq.Price.Max)
			argIndex++
		}
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM products " + whereClause
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Failed to count products in PostgreSQL search")
		return nil, err
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if searchReq.SortBy != "" {
		sortOrder := "ASC"
		if searchReq.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		orderBy = "ORDER BY " + searchReq.SortBy + " " + sortOrder
	}

	// Get products
	query := `
		SELECT id, name, description, brand, category, price, stock_quantity, created_at, updated_at
		FROM products ` + whereClause + ` ` + orderBy + `
		LIMIT $` + string(rune(argIndex)) + ` OFFSET $` + string(rune(argIndex+1))

	args = append(args, searchReq.Size, offset)

	var products []models.Product
	err = r.db.Select(&products, query, args...)
	if err != nil {
		logrus.WithError(err).Error("Failed to search products in PostgreSQL")
		return nil, err
	}

	// Simple aggregations (PostgreSQL doesn't have as rich aggregation as Elasticsearch)
	brandAggs := make(map[string]interface{})
	categoryAggs := make(map[string]interface{})

	aggregations := map[string]interface{}{
		"brands":     brandAggs,
		"categories": categoryAggs,
	}

	return &models.SearchResponse{
		Products:     products,
		Total:        total,
		Page:         searchReq.Page,
		Size:         searchReq.Size,
		Aggregations: aggregations,
	}, nil
}