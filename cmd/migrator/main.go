package main

import (
	"flag"
	"fmt"
	"os"

	"ecommerce-catalog-api/internal/config"
	"ecommerce-catalog-api/internal/database"
	"ecommerce-catalog-api/internal/elasticsearch"
	"ecommerce-catalog-api/internal/repository"
	"ecommerce-catalog-api/internal/services"
	"ecommerce-catalog-api/pkg/logger"
	"ecommerce-catalog-api/pkg/utils"

	"github.com/sirupsen/logrus"
)

func main() {
	var (
		seedFlag     = flag.Bool("seed", false, "Seed the database with sample data")
		migrateFlag  = flag.Bool("migrate", false, "Migrate data from PostgreSQL to Elasticsearch")
		seedCount    = flag.Int("count", 1000, "Number of products to seed (default: 1000)")
		resetFlag    = flag.Bool("reset", false, "Reset Elasticsearch index before migration")
	)
	flag.Parse()

	if !*seedFlag && !*migrateFlag {
		fmt.Println("Usage:")
		fmt.Println("  migrator -seed -count=1000    # Seed database with 1000 products")
		fmt.Println("  migrator -migrate             # Migrate data from PostgreSQL to Elasticsearch")
		fmt.Println("  migrator -migrate -reset      # Reset Elasticsearch index and migrate")
		fmt.Println("  migrator -seed -migrate       # Seed and then migrate")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Initialize logger
	logger.Initialize(cfg.Logger.Level)
	logrus.Info("Starting Migrator Tool")

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

	// Initialize repositories and services
	productRepo := repository.NewProductRepository(db.DB)
	commandService := services.NewProductCommandService(productRepo, esClient)

	if *seedFlag {
		if err := seedDatabase(commandService, *seedCount); err != nil {
			logrus.WithError(err).Fatal("Failed to seed database")
		}
	}

	if *migrateFlag {
		if err := migrateData(productRepo, esClient, *resetFlag); err != nil {
			logrus.WithError(err).Fatal("Failed to migrate data")
		}
	}

	logrus.Info("Migrator completed successfully")
}

func seedDatabase(commandService *services.ProductCommandService, count int) error {
	logrus.WithField("count", count).Info("Starting database seeding")

	seeder := utils.NewProductSeeder()
	products := seeder.GenerateProducts(count)

	successCount := 0
	errorCount := 0

	for i, productCreate := range products {
		if i%100 == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": fmt.Sprintf("%d/%d", i, count),
				"percent":  fmt.Sprintf("%.1f%%", float64(i)/float64(count)*100),
			}).Info("Seeding progress")
		}

		_, err := commandService.CreateProduct(&productCreate)
		if err != nil {
			logrus.WithError(err).WithField("product_name", productCreate.Name).Warn("Failed to create product")
			errorCount++
		} else {
			successCount++
		}
	}

	logrus.WithFields(logrus.Fields{
		"total":   count,
		"success": successCount,
		"errors":  errorCount,
	}).Info("Database seeding completed")

	if errorCount > 0 {
		return fmt.Errorf("seeding completed with %d errors", errorCount)
	}

	return nil
}

func migrateData(productRepo *repository.ProductRepository, esClient *elasticsearch.Client, reset bool) error {
	logrus.Info("Starting data migration from PostgreSQL to Elasticsearch")

	if reset {
		logrus.Info("Resetting Elasticsearch index")
		if err := esClient.CreateProductsIndex(); err != nil {
			return fmt.Errorf("failed to reset Elasticsearch index: %w", err)
		}
	}

	// Get all products from PostgreSQL
	products, err := productRepo.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get products from PostgreSQL: %w", err)
	}

	if len(products) == 0 {
		logrus.Warn("No products found in PostgreSQL")
		return nil
	}

	logrus.WithField("count", len(products)).Info("Products retrieved from PostgreSQL")

	// Bulk index to Elasticsearch in batches
	batchSize := 100
	successCount := 0
	errorCount := 0

	for i := 0; i < len(products); i += batchSize {
		end := i + batchSize
		if end > len(products) {
			end = len(products)
		}

		batch := products[i:end]
		
		logrus.WithFields(logrus.Fields{
			"batch":    fmt.Sprintf("%d-%d", i+1, end),
			"progress": fmt.Sprintf("%.1f%%", float64(end)/float64(len(products))*100),
		}).Info("Processing batch")

		if err := esClient.BulkIndexProducts(batch); err != nil {
			logrus.WithError(err).WithField("batch", fmt.Sprintf("%d-%d", i+1, end)).Error("Failed to index batch")
			errorCount += len(batch)
		} else {
			successCount += len(batch)
		}
	}

	logrus.WithFields(logrus.Fields{
		"total":   len(products),
		"success": successCount,
		"errors":  errorCount,
	}).Info("Data migration completed")

	if errorCount > 0 {
		return fmt.Errorf("migration completed with %d errors", errorCount)
	}

	return nil
}