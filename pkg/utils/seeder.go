package utils

import (
	"math/rand"
	"time"

	"ecommerce-catalog-api/internal/models"
)

// ProductSeeder provides methods to generate seed data
type ProductSeeder struct {
	brands     []string
	categories []string
	products   []ProductSeed
}

type ProductSeed struct {
	Name         string
	Description  string
	Brand        string
	Category     string
	PriceMin     float64
	PriceMax     float64
	StockMin     int
	StockMax     int
}

// NewProductSeeder creates a new product seeder
func NewProductSeeder() *ProductSeeder {
	return &ProductSeeder{
		brands: []string{
			"Apple", "Samsung", "Sony", "LG", "Dell", "HP", "Lenovo", "ASUS",
			"Nike", "Adidas", "Puma", "Under Armour", "New Balance",
			"Canon", "Nikon", "Panasonic", "JBL", "Bose", "Beats",
			"Microsoft", "Google", "Amazon", "Tesla", "BMW", "Mercedes",
			"Coca-Cola", "Pepsi", "Nestlé", "Unilever", "P&G",
		},
		categories: []string{
			"Electronics", "Computers", "Smartphones", "Audio", "Cameras",
			"Gaming", "Sports", "Fashion", "Home & Garden", "Books",
			"Health & Beauty", "Automotive", "Toys", "Kitchen", "Office",
		},
		products: []ProductSeed{
			// Electronics
			{
				Name: "Wireless Bluetooth Headphones", Description: "Premium over-ear wireless headphones with noise cancellation and 30-hour battery life",
				Category: "Audio", PriceMin: 99.99, PriceMax: 399.99, StockMin: 10, StockMax: 100,
			},
			{
				Name: "4K Smart TV", Description: "Ultra HD 4K Smart TV with HDR support and built-in streaming apps",
				Category: "Electronics", PriceMin: 299.99, PriceMax: 1999.99, StockMin: 5, StockMax: 50,
			},
			{
				Name: "Gaming Laptop", Description: "High-performance gaming laptop with dedicated graphics card and RGB keyboard",
				Category: "Computers", PriceMin: 799.99, PriceMax: 2999.99, StockMin: 3, StockMax: 25,
			},
			{
				Name: "Wireless Mouse", Description: "Ergonomic wireless mouse with precision tracking and long battery life",
				Category: "Computers", PriceMin: 19.99, PriceMax: 79.99, StockMin: 50, StockMax: 200,
			},
			{
				Name: "Smartphone", Description: "Latest flagship smartphone with advanced camera system and 5G connectivity",
				Category: "Smartphones", PriceMin: 199.99, PriceMax: 1299.99, StockMin: 20, StockMax: 100,
			},
			// Sports & Fashion
			{
				Name: "Running Shoes", Description: "Lightweight running shoes with advanced cushioning technology",
				Category: "Sports", PriceMin: 79.99, PriceMax: 249.99, StockMin: 30, StockMax: 150,
			},
			{
				Name: "Fitness Tracker", Description: "Advanced fitness tracker with heart rate monitoring and GPS",
				Category: "Sports", PriceMin: 49.99, PriceMax: 299.99, StockMin: 25, StockMax: 120,
			},
			{
				Name: "Yoga Mat", Description: "Premium non-slip yoga mat with excellent grip and cushioning",
				Category: "Sports", PriceMin: 29.99, PriceMax: 89.99, StockMin: 40, StockMax: 200,
			},
			{
				Name: "Designer Sunglasses", Description: "Stylish designer sunglasses with UV protection",
				Category: "Fashion", PriceMin: 99.99, PriceMax: 599.99, StockMin: 15, StockMax: 80,
			},
			{
				Name: "Leather Wallet", Description: "Genuine leather wallet with RFID blocking technology",
				Category: "Fashion", PriceMin: 39.99, PriceMax: 149.99, StockMin: 50, StockMax: 200,
			},
			// Home & Kitchen
			{
				Name: "Coffee Maker", Description: "Programmable coffee maker with thermal carafe and auto-shutoff",
				Category: "Kitchen", PriceMin: 49.99, PriceMax: 299.99, StockMin: 20, StockMax: 80,
			},
			{
				Name: "Blender", Description: "High-speed blender perfect for smoothies and food preparation",
				Category: "Kitchen", PriceMin: 39.99, PriceMax: 399.99, StockMin: 25, StockMax: 100,
			},
			{
				Name: "Air Fryer", Description: "Digital air fryer for healthy cooking with little to no oil",
				Category: "Kitchen", PriceMin: 79.99, PriceMax: 249.99, StockMin: 15, StockMax: 60,
			},
			{
				Name: "Smart Thermostat", Description: "WiFi-enabled smart thermostat with energy-saving features",
				Category: "Home & Garden", PriceMin: 149.99, PriceMax: 299.99, StockMin: 10, StockMax: 40,
			},
			{
				Name: "LED Light Bulbs", Description: "Energy-efficient LED light bulbs with dimming capability",
				Category: "Home & Garden", PriceMin: 9.99, PriceMax: 49.99, StockMin: 100, StockMax: 500,
			},
			// Books & Education
			{
				Name: "Programming Book", Description: "Comprehensive guide to modern programming languages and best practices",
				Category: "Books", PriceMin: 29.99, PriceMax: 79.99, StockMin: 50, StockMax: 200,
			},
			{
				Name: "E-reader", Description: "Waterproof e-reader with adjustable lighting and weeks of battery life",
				Category: "Electronics", PriceMin: 99.99, PriceMax: 299.99, StockMin: 20, StockMax: 80,
			},
			// Health & Beauty
			{
				Name: "Electric Toothbrush", Description: "Rechargeable electric toothbrush with multiple cleaning modes",
				Category: "Health & Beauty", PriceMin: 49.99, PriceMax: 199.99, StockMin: 30, StockMax: 120,
			},
			{
				Name: "Face Moisturizer", Description: "Anti-aging face moisturizer with SPF protection",
				Category: "Health & Beauty", PriceMin: 19.99, PriceMax: 89.99, StockMin: 40, StockMax: 150,
			},
			{
				Name: "Hair Dryer", Description: "Professional hair dryer with ionic technology and multiple heat settings",
				Category: "Health & Beauty", PriceMin: 39.99, PriceMax: 299.99, StockMin: 25, StockMax: 100,
			},
			// Gaming
			{
				Name: "Gaming Console", Description: "Next-generation gaming console with 4K graphics and ray tracing",
				Category: "Gaming", PriceMin: 299.99, PriceMax: 599.99, StockMin: 5, StockMax: 30,
			},
			{
				Name: "Gaming Headset", Description: "Surround sound gaming headset with noise-canceling microphone",
				Category: "Gaming", PriceMin: 59.99, PriceMax: 249.99, StockMin: 20, StockMax: 80,
			},
			{
				Name: "Mechanical Keyboard", Description: "RGB mechanical gaming keyboard with customizable keys",
				Category: "Gaming", PriceMin: 89.99, PriceMax: 299.99, StockMin: 15, StockMax: 60,
			},
			// Automotive
			{
				Name: "Car Phone Mount", Description: "Magnetic car phone mount with 360-degree rotation",
				Category: "Automotive", PriceMin: 19.99, PriceMax: 49.99, StockMin: 50, StockMax: 200,
			},
			{
				Name: "Car Charger", Description: "Fast-charging USB car charger with multiple ports",
				Category: "Automotive", PriceMin: 14.99, PriceMax: 39.99, StockMin: 75, StockMax: 300,
			},
			{
				Name: "Dash Camera", Description: "Full HD dash camera with loop recording and G-sensor",
				Category: "Automotive", PriceMin: 59.99, PriceMax: 199.99, StockMin: 20, StockMax: 80,
			},
			// Office
			{
				Name: "Office Chair", Description: "Ergonomic office chair with lumbar support and adjustable height",
				Category: "Office", PriceMin: 149.99, PriceMax: 599.99, StockMin: 10, StockMax: 40,
			},
			{
				Name: "Monitor", Description: "4K UHD monitor with IPS panel and USB-C connectivity",
				Category: "Office", PriceMin: 249.99, PriceMax: 899.99, StockMin: 15, StockMax: 50,
			},
			{
				Name: "Desk Organizer", Description: "Bamboo desk organizer with multiple compartments",
				Category: "Office", PriceMin: 24.99, PriceMax: 79.99, StockMin: 40, StockMax: 150,
			},
		},
	}
}

// GenerateProducts generates a specified number of realistic product data
func (s *ProductSeeder) GenerateProducts(count int) []models.ProductCreate {
	rand.Seed(time.Now().UnixNano())
	
	var products []models.ProductCreate
	
	for i := 0; i < count; i++ {
		seed := s.products[rand.Intn(len(s.products))]
		brand := s.brands[rand.Intn(len(s.brands))]
		
		// Generate random price within the seed's range
		price := seed.PriceMin + rand.Float64()*(seed.PriceMax-seed.PriceMin)
		
		// Generate random stock quantity within the seed's range
		stockQuantity := seed.StockMin + rand.Intn(seed.StockMax-seed.StockMin+1)
		
		// Add variation to the product name
		variations := []string{"Pro", "Max", "Plus", "Elite", "Premium", "Standard", "Basic", "Advanced", "Ultimate", "Deluxe"}
		variation := variations[rand.Intn(len(variations))]
		
		productName := seed.Name
		if rand.Float32() < 0.3 { // 30% chance to add variation
			productName = brand + " " + seed.Name + " " + variation
		} else {
			productName = brand + " " + seed.Name
		}
		
		// Add some variation to description
		descriptions := []string{
			seed.Description,
			seed.Description + " Perfect for everyday use.",
			seed.Description + " Highly rated by customers.",
			seed.Description + " Best-seller in its category.",
			seed.Description + " Limited time offer.",
		}
		
		product := models.ProductCreate{
			Name:         productName,
			Description:  descriptions[rand.Intn(len(descriptions))],
			Brand:        brand,
			Category:     seed.Category,
			Price:        float64(int(price*100)) / 100, // Round to 2 decimal places
			StockQuantity: stockQuantity,
		}
		
		products = append(products, product)
	}
	
	return products
}