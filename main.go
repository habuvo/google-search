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

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GoogleSearchResult represents a single search result.
type GoogleSearchResult struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Snippet     string `json:"snippet"`
	DisplayLink string `json:"displayLink"`
}

// GoogleSearchResponse represents the response from Google Custom Search API.
type GoogleSearchResponse struct {
	Items []GoogleSearchResult `json:"items"`
}

// Config holds the application configuration.
type Config struct {
	APIKey         string
	SearchEngineID string
}

const (
	maxNumResults     = 10
	defaultNumResults = 5
	baseURL           = "https://www.googleapis.com/customsearch/v1"
)

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create MCP server
	s := createServer()

	// Create and register Google Search tool
	registerGoogleSearchTool(s, config)

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// loadConfig loads and validates the application configuration.
func loadConfig() (*Config, error) {
	// Check for required environment variables
	apiKey := os.Getenv("GOOGLE_API_KEY")
	searchEngineID := os.Getenv("GOOGLE_SEARCH_ENGINE_ID")

	if apiKey == "" || searchEngineID == "" {
		return nil, fmt.Errorf("GOOGLE_API_KEY and GOOGLE_SEARCH_ENGINE_ID environment variables are required")
	}

	return &Config{
		APIKey:         apiKey,
		SearchEngineID: searchEngineID,
	}, nil
}

// createServer creates and configures the MCP server.
func createServer() *server.MCPServer {
	return server.NewMCPServer(
		"Google Search MCP Server",
		"1.0.0",
		server.WithLogging(),
	)
}

// registerGoogleSearchTool creates and registers the Google Search tool with the server.
func registerGoogleSearchTool(s *server.MCPServer, config *Config) {
	// Create Google Search tool
	googleSearchTool := createGoogleSearchTool()

	// Add Google Search tool handler
	s.AddTool(googleSearchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleGoogleSearchRequest(ctx, request, config)
	})
}

// createGoogleSearchTool creates and configures the Google Search tool.
func createGoogleSearchTool() mcp.Tool {
	return mcp.NewTool("google_search",
		mcp.WithDescription("Search the web using Google Custom Search"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The search query"),
		),
		mcp.WithNumber("num_results",
			mcp.Description(fmt.Sprintf("Number of results to return (max %d, default %d)", maxNumResults, defaultNumResults)),
		),
	)
}

// handleGoogleSearchRequest processes a Google Search tool request.
func handleGoogleSearchRequest(_ context.Context,
	request mcp.CallToolRequest,
	config *Config,
) (*mcp.CallToolResult, error) {
	// Extract and validate query parameter
	query, ok := request.Params.Arguments["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query must be a non-empty string")
	}

	// Extract and validate num_results parameter
	numResults := extractNumResults(request.Params.Arguments)

	// Call Google Custom Search API
	results, err := performGoogleSearch(query, numResults, config.APIKey, config.SearchEngineID)
	if err != nil {
		return nil, fmt.Errorf("search failed: %v", err)
	}

	// Format results
	formattedResults := formatSearchResults(results)

	return mcp.NewToolResultText(formattedResults), nil
}

// extractNumResults extracts and validates the num_results parameter.
func extractNumResults(arguments map[string]interface{}) int {
	numResults := defaultNumResults

	if numResultsArg, ok := arguments["num_results"]; ok {
		if numResultsFloat, ok := numResultsArg.(float64); ok {
			numResults = int(numResultsFloat)
			if numResults < 1 || numResults > maxNumResults {
				numResults = maxNumResults
			}
		}
	}

	return numResults
}

// performGoogleSearch calls the Google Custom Search API and returns the results.
func performGoogleSearch(query string, numResults int, apiKey, searchEngineID string) ([]GoogleSearchResult, error) {
	// Build the request parameters
	params := buildSearchParams(query, numResults, apiKey, searchEngineID)

	// Make the HTTP request
	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	return parseSearchResponse(resp)
}

// buildSearchParams creates the URL parameters for the Google Search API request.
func buildSearchParams(query string, numResults int, apiKey, searchEngineID string) url.Values {
	params := url.Values{}
	params.Add("key", apiKey)
	params.Add("cx", searchEngineID)
	params.Add("q", query)
	params.Add("num", strconv.Itoa(numResults))

	return params
}

// parseSearchResponse processes the HTTP response from the Google Search API.
func parseSearchResponse(resp *http.Response) ([]GoogleSearchResult, error) {
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

// formatSearchResults formats the search results into a readable string.
func formatSearchResults(results []GoogleSearchResult) string {
	if len(results) == 0 {
		return "No results found."
	}

	var sb *strings.Builder

	fmt.Fprintf(sb, "Found %d results:\n\n", len(results))

	for i, result := range results {
		formatSingleResult(sb, i, result)
	}

	return sb.String()
}

// formatSingleResult formats a single search result and appends it to the string builder.
func formatSingleResult(sb *strings.Builder, index int, result GoogleSearchResult) {
	fmt.Fprintf(sb, "%d. %s\n", index+1, result.Title)
	fmt.Fprintf(sb, "   URL: %s\n", result.Link)
	fmt.Fprintf(sb, "   %s\n\n", result.Snippet)
}
