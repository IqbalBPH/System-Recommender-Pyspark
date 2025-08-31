#!/bin/bash

# E-commerce Catalog API Demo Script
# This script demonstrates the full functionality of the API

echo "🚀 E-commerce Catalog API Demo"
echo "================================="

API_BASE="http://localhost:8080"

echo ""
echo "📊 1. Health Check"
curl -s "$API_BASE/health" | jq .

echo ""
echo "📦 2. Creating Sample Products"

# Create Product 1
echo "Creating Apple MacBook Pro..."
PRODUCT1=$(curl -s -X POST "$API_BASE/api/v1/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Apple MacBook Pro 16-inch",
    "description": "Professional laptop with M3 Pro chip, perfect for developers and content creators",
    "brand": "Apple",
    "category": "Computers",
    "price": 2499.99,
    "stock_quantity": 15
  }')

PRODUCT1_ID=$(echo $PRODUCT1 | jq -r '.data.id')
echo "✅ Created product with ID: $PRODUCT1_ID"

# Create Product 2
echo "Creating Sony Headphones..."
PRODUCT2=$(curl -s -X POST "$API_BASE/api/v1/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Sony WH-1000XM5 Wireless Headphones",
    "description": "Industry-leading noise canceling headphones with 30-hour battery life",
    "brand": "Sony",
    "category": "Audio",
    "price": 399.99,
    "stock_quantity": 50
  }')

PRODUCT2_ID=$(echo $PRODUCT2 | jq -r '.data.id')
echo "✅ Created product with ID: $PRODUCT2_ID"

# Create Product 3
echo "Creating Gaming Mouse..."
PRODUCT3=$(curl -s -X POST "$API_BASE/api/v1/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Logitech G Pro X Gaming Mouse",
    "description": "Professional gaming mouse with HERO sensor and customizable buttons",
    "brand": "Logitech",
    "category": "Gaming",
    "price": 129.99,
    "stock_quantity": 25
  }')

PRODUCT3_ID=$(echo $PRODUCT3 | jq -r '.data.id')
echo "✅ Created product with ID: $PRODUCT3_ID"

echo ""
echo "🔍 3. Search Operations"

echo ""
echo "Simple text search for 'laptop':"
curl -s "$API_BASE/api/v1/search?query=laptop&size=5" | jq '.data.products[] | {id, name, brand, price}'

echo ""
echo "Brand filter search for 'Sony':"
curl -s "$API_BASE/api/v1/search?brand=Sony" | jq '.data.products[] | {id, name, brand, price}'

echo ""
echo "Category filter for 'Gaming' products:"
curl -s "$API_BASE/api/v1/search?category=Gaming" | jq '.data.products[] | {id, name, category, price}'

echo ""
echo "Price range search ($100-$500):"
curl -s "$API_BASE/api/v1/search?price_min=100&price_max=500" | jq '.data.products[] | {id, name, price}'

echo ""
echo "Complex search with sorting by price (ascending):"
curl -s "$API_BASE/api/v1/search?query=Pro&sort_by=price&sort_order=asc" | jq '.data.products[] | {id, name, price}'

echo ""
echo "📈 4. Performance Comparison"

echo ""
echo "Elasticsearch search with timing:"
time curl -s "$API_BASE/api/v1/search?query=wireless&brand=Sony" -w "\nResponse Time: %{time_total}s\n" | jq '.data.total'

echo ""
echo "PostgreSQL search with timing:"
time curl -s "$API_BASE/api/v1/search/postgresql?query=wireless&brand=Sony" -w "\nResponse Time: %{time_total}s\n" | jq '.data.total'

echo ""
echo "📝 5. Product Management"

echo ""
echo "Get product by ID:"
curl -s "$API_BASE/api/v1/products/$PRODUCT1_ID" | jq '.data | {id, name, brand, price, stock_quantity}'

echo ""
echo "Update product price:"
curl -s -X PUT "$API_BASE/api/v1/products/$PRODUCT1_ID" \
  -H "Content-Type: application/json" \
  -d '{"price": 2299.99, "stock_quantity": 20}' | jq '.data | {id, name, price, stock_quantity}'

echo ""
echo "📋 6. List All Products"
curl -s "$API_BASE/api/v1/products?page=1&size=5" | jq '.data | {total, page, size, products: .products[] | {id, name, brand, price}}'

echo ""
echo "🔮 7. Aggregations Demo"
echo ""
echo "Search with aggregations (shows product count by brand and category):"
curl -s "$API_BASE/api/v1/search?query=*&size=0" | jq '.data.aggregations'

echo ""
echo "🧹 8. Cleanup (Optional)"
echo "To clean up created products, run:"
echo "curl -X DELETE '$API_BASE/api/v1/products/$PRODUCT1_ID'"
echo "curl -X DELETE '$API_BASE/api/v1/products/$PRODUCT2_ID'"
echo "curl -X DELETE '$API_BASE/api/v1/products/$PRODUCT3_ID'"

echo ""
echo "✨ Demo Complete!"
echo ""
echo "🚀 Next Steps:"
echo "1. Run the migrator to seed more data: ./migrator -seed -count=1000"
echo "2. Run performance tests: ./performance-test -users=50 -duration=60s -type=both"
echo "3. Explore the API documentation in docs/API.md"
echo "4. Check the detailed README.md for deployment instructions"