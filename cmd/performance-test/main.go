package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type TestConfig struct {
	BaseURL         string
	ConcurrentUsers int
	Duration        time.Duration
	TestType        string
}

type TestResult struct {
	Endpoint            string
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	AverageResponseTime time.Duration
	MinResponseTime     time.Duration
	MaxResponseTime     time.Duration
	RequestsPerSecond   float64
	TestDuration        time.Duration
}

var searchQueries = []string{
	"laptop",
	"phone",
	"headphones",
	"shoes",
	"coffee",
	"camera",
	"gaming",
	"book",
	"watch",
	"speaker",
}

var brands = []string{
	"Apple",
	"Samsung",
	"Sony",
	"Nike",
	"Dell",
	"HP",
	"Canon",
	"Microsoft",
}

var categories = []string{
	"Electronics",
	"Computers",
	"Sports",
	"Audio",
	"Gaming",
	"Fashion",
}

func main() {
	var (
		baseURL     = flag.String("url", "http://localhost:8080", "Base URL of the API")
		users       = flag.Int("users", 10, "Number of concurrent users")
		duration    = flag.Duration("duration", 30*time.Second, "Test duration")
		testType    = flag.String("type", "both", "Test type: elasticsearch, postgresql, or both")
	)
	flag.Parse()

	logrus.Info("Starting Performance Test")

	config := TestConfig{
		BaseURL:         *baseURL,
		ConcurrentUsers: *users,
		Duration:        *duration,
		TestType:        *testType,
	}

	var results []TestResult

	if config.TestType == "elasticsearch" || config.TestType == "both" {
		logrus.Info("Testing Elasticsearch endpoint")
		result := runPerformanceTest(config, "/api/v1/search", "Elasticsearch")
		results = append(results, result)
	}

	if config.TestType == "postgresql" || config.TestType == "both" {
		logrus.Info("Testing PostgreSQL endpoint")
		result := runPerformanceTest(config, "/api/v1/search/postgresql", "PostgreSQL")
		results = append(results, result)
	}

	// Print results
	printResults(results)

	// Compare results if both were tested
	if len(results) > 1 {
		compareResults(results)
	}
}

func runPerformanceTest(config TestConfig, endpoint, engineName string) TestResult {
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	totalRequests := int64(0)
	successfulRequests := int64(0)
	failedRequests := int64(0)
	responseTimes := []time.Duration{}
	
	startTime := time.Now()
	stopTime := startTime.Add(config.Duration)
	
	// Start concurrent users
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			
			client := &http.Client{
				Timeout: 10 * time.Second,
			}
			
			for time.Now().Before(stopTime) {
				// Create random search request
				searchReq := generateRandomSearchRequest()
				
				// Make request
				requestStart := time.Now()
				success := makeSearchRequest(client, config.BaseURL+endpoint, searchReq)
				responseTime := time.Since(requestStart)
				
				// Update statistics
				mu.Lock()
				totalRequests++
				responseTimes = append(responseTimes, responseTime)
				if success {
					successfulRequests++
				} else {
					failedRequests++
				}
				mu.Unlock()
				
				// Small delay to prevent overwhelming the server
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}
	
	wg.Wait()
	actualDuration := time.Since(startTime)
	
	// Calculate statistics
	var totalResponseTime time.Duration
	minResponseTime := time.Hour
	maxResponseTime := time.Duration(0)
	
	for _, rt := range responseTimes {
		totalResponseTime += rt
		if rt < minResponseTime {
			minResponseTime = rt
		}
		if rt > maxResponseTime {
			maxResponseTime = rt
		}
	}
	
	var averageResponseTime time.Duration
	if len(responseTimes) > 0 {
		averageResponseTime = totalResponseTime / time.Duration(len(responseTimes))
	}
	
	requestsPerSecond := float64(totalRequests) / actualDuration.Seconds()
	
	return TestResult{
		Endpoint:            engineName,
		TotalRequests:       totalRequests,
		SuccessfulRequests:  successfulRequests,
		FailedRequests:      failedRequests,
		AverageResponseTime: averageResponseTime,
		MinResponseTime:     minResponseTime,
		MaxResponseTime:     maxResponseTime,
		RequestsPerSecond:   requestsPerSecond,
		TestDuration:        actualDuration,
	}
}

func generateRandomSearchRequest() map[string]interface{} {
	requests := []map[string]interface{}{
		// Simple text search
		{
			"query": searchQueries[time.Now().UnixNano()%int64(len(searchQueries))],
			"page":  1,
			"size":  20,
		},
		// Brand filter
		{
			"brand": []string{brands[time.Now().UnixNano()%int64(len(brands))]},
			"page":  1,
			"size":  20,
		},
		// Category filter
		{
			"category": []string{categories[time.Now().UnixNano()%int64(len(categories))]},
			"page":     1,
			"size":     20,
		},
		// Price range
		{
			"price_min": 50.0,
			"price_max": 200.0,
			"page":      1,
			"size":      20,
		},
		// Complex search
		{
			"query":     searchQueries[time.Now().UnixNano()%int64(len(searchQueries))],
			"brand":     []string{brands[time.Now().UnixNano()%int64(len(brands))]},
			"price_min": 20.0,
			"price_max": 500.0,
			"sort_by":   "price",
			"sort_order": "asc",
			"page":      1,
			"size":      20,
		},
	}
	
	return requests[time.Now().UnixNano()%int64(len(requests))]
}

func makeSearchRequest(client *http.Client, url string, searchReq map[string]interface{}) bool {
	// Convert to query parameters
	queryParams := "?"
	first := true
	
	for key, value := range searchReq {
		if !first {
			queryParams += "&"
		}
		queryParams += fmt.Sprintf("%s=%v", key, value)
		first = false
	}
	
	resp, err := client.Get(url + queryParams)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

func printResults(results []TestResult) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("PERFORMANCE TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))
	
	for _, result := range results {
		fmt.Printf("\n%s Results:\n", result.Endpoint)
		fmt.Println(strings.Repeat("-", 40))
		fmt.Printf("Total Requests:       %d\n", result.TotalRequests)
		fmt.Printf("Successful Requests:  %d\n", result.SuccessfulRequests)
		fmt.Printf("Failed Requests:      %d\n", result.FailedRequests)
		fmt.Printf("Success Rate:         %.2f%%\n", float64(result.SuccessfulRequests)/float64(result.TotalRequests)*100)
		fmt.Printf("Average Response Time: %v\n", result.AverageResponseTime)
		fmt.Printf("Min Response Time:    %v\n", result.MinResponseTime)
		fmt.Printf("Max Response Time:    %v\n", result.MaxResponseTime)
		fmt.Printf("Requests per Second:  %.2f\n", result.RequestsPerSecond)
		fmt.Printf("Test Duration:        %v\n", result.TestDuration)
	}
}

func compareResults(results []TestResult) {
	if len(results) != 2 {
		return
	}
	
	elasticsearch := results[0]
	postgresql := results[1]
	
	// Ensure correct order
	if elasticsearch.Endpoint != "Elasticsearch" {
		elasticsearch, postgresql = postgresql, elasticsearch
	}
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("COMPARISON: ELASTICSEARCH vs POSTGRESQL")
	fmt.Println(strings.Repeat("=", 80))
	
	fmt.Printf("\nRequests per Second:\n")
	fmt.Printf("  Elasticsearch: %.2f req/s\n", elasticsearch.RequestsPerSecond)
	fmt.Printf("  PostgreSQL:    %.2f req/s\n", postgresql.RequestsPerSecond)
	improvement := ((elasticsearch.RequestsPerSecond - postgresql.RequestsPerSecond) / postgresql.RequestsPerSecond) * 100
	fmt.Printf("  Improvement:   %.2f%%\n", improvement)
	
	fmt.Printf("\nAverage Response Time:\n")
	fmt.Printf("  Elasticsearch: %v\n", elasticsearch.AverageResponseTime)
	fmt.Printf("  PostgreSQL:    %v\n", postgresql.AverageResponseTime)
	timeImprovement := float64(postgresql.AverageResponseTime-elasticsearch.AverageResponseTime) / float64(postgresql.AverageResponseTime) * 100
	fmt.Printf("  Improvement:   %.2f%%\n", timeImprovement)
	
	fmt.Printf("\nSuccess Rate:\n")
	esSuccess := float64(elasticsearch.SuccessfulRequests) / float64(elasticsearch.TotalRequests) * 100
	pgSuccess := float64(postgresql.SuccessfulRequests) / float64(postgresql.TotalRequests) * 100
	fmt.Printf("  Elasticsearch: %.2f%%\n", esSuccess)
	fmt.Printf("  PostgreSQL:    %.2f%%\n", pgSuccess)
	
	fmt.Println("\n" + strings.Repeat("=", 80))
	
	if improvement > 0 && timeImprovement > 0 {
		fmt.Printf("🎉 Elasticsearch outperformed PostgreSQL by %.2f%% in throughput and %.2f%% in response time!\n", improvement, timeImprovement)
	} else if improvement > 0 {
		fmt.Printf("⚡ Elasticsearch achieved %.2f%% higher throughput than PostgreSQL\n", improvement)
	} else if timeImprovement > 0 {
		fmt.Printf("⚡ Elasticsearch achieved %.2f%% faster response times than PostgreSQL\n", timeImprovement)
	} else {
		fmt.Println("📊 PostgreSQL performed competitively with Elasticsearch")
	}
}