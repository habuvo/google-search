package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GoogleSearchResult represents a single search result
type GoogleSearchResult struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Snippet     string `json:"snippet"`
	DisplayLink string `json:"displayLink"`
}

// GoogleSearchResponse represents the response from Google Custom Search API
type GoogleSearchResponse struct {
	Items []GoogleSearchResult `json:"items"`
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found. Using environment variables.")
	}

	// Check for required environment variables
	apiKey := os.Getenv("GOOGLE_API_KEY")
	searchEngineID := os.Getenv("GOOGLE_SEARCH_ENGINE_ID")

	if apiKey == "" || searchEngineID == "" {
		log.Fatal("Error: GOOGLE_API_KEY and GOOGLE_SEARCH_ENGINE_ID environment variables are required")
	}

	// Create MCP server
	s := server.NewMCPServer(
		"Google Search MCP Server",
		"1.0.0",
		server.WithLogging(),
	)

	// Create Google Search tool
	googleSearchTool := mcp.NewTool("google_search",
		mcp.WithDescription("Search the web using Google Custom Search"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The search query"),
		),
		mcp.WithNumber("num_results",
			mcp.Description("Number of results to return (max 10, default 5)"),
		),
	)

	// Add Google Search tool handler
	s.AddTool(googleSearchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract query parameter
		query, ok := request.Params.Arguments["query"].(string)
		if !ok || query == "" {
			return nil, fmt.Errorf("query must be a non-empty string")
		}

		// Extract num_results parameter (default to 5, max 10)
		numResults := 5
		if numResultsArg, ok := request.Params.Arguments["num_results"]; ok {
			if numResultsFloat, ok := numResultsArg.(float64); ok {
				numResults = int(numResultsFloat)
				if numResults < 1 {
					numResults = 5
				} else if numResults > 10 {
					numResults = 10
				}
			}
		}

		// Call Google Custom Search API
		results, err := performGoogleSearch(query, numResults, apiKey, searchEngineID)
		if err != nil {
			return nil, fmt.Errorf("search failed: %v", err)
		}

		// Format results
		formattedResults := formatSearchResults(results)

		return mcp.NewToolResultText(formattedResults), nil
	})

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// performGoogleSearch calls the Google Custom Search API and returns the results
func performGoogleSearch(query string, numResults int, apiKey, searchEngineID string) ([]GoogleSearchResult, error) {
	// Build the API URL
	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}
	params.Add("key", apiKey)
	params.Add("cx", searchEngineID)
	params.Add("q", query)
	params.Add("num", strconv.Itoa(numResults))

	// Make the HTTP request
	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-200 status: %d - %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var searchResponse GoogleSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %v", err)
	}

	return searchResponse.Items, nil
}

// formatSearchResults formats the search results into a readable string
func formatSearchResults(results []GoogleSearchResult) string {
	if len(results) == 0 {
		return "No results found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d results:\n\n", len(results)))

	for i, result := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Title))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", result.Link))
		sb.WriteString(fmt.Sprintf("   %s\n\n", result.Snippet))
	}

	return sb.String()
}
