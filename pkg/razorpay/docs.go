package razorpay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
)

// SearchDocs returns a tool that searches Razorpay documentation
func SearchDocs(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"query",
			mcpgo.Description("Search terms to look for in documentation"),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		params := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "query")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		query := params["query"].(string)

		// Create HTTP request to the documentation search endpoint
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("https://search.razorpay.com/docs?q=%s", url.QueryEscape(query)),
			nil,
		)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to create request: %s", err.Error())), nil
		}

		// Add required headers
		req.Header.Set("sec-ch-ua-platform", "macOS")
		req.Header.Set("x-country-code", "IN")
		req.Header.Set("Referer", "https://razorpay.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
		req.Header.Set("sec-ch-ua", "\"Chromium\";v=\"136\", \"Google Chrome\";v=\"136\", \"Not.A/Brand\";v=\"99\"")
		req.Header.Set("content-type", "text/plain; charset=utf-8")
		req.Header.Set("sec-ch-ua-mobile", "?0")

		// Create HTTP client and send request
		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("search request failed: %s", err.Error())), nil
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to read response: %s", err.Error())), nil
		}

		// Check if the response status code is not 200 OK
		if resp.StatusCode != http.StatusOK {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("search failed with status code %d: %s",
					resp.StatusCode, string(body))), nil
		}

		// Parse the response as JSON
		var searchResult []map[string]interface{}
		err = json.Unmarshal(body, &searchResult)
		if err != nil {
			// Try parsing as error response
			var errorResult map[string]interface{}
			errParseErr := json.Unmarshal(body, &errorResult)
			if errParseErr == nil {
				// Check if it's an error response
				if errorMsg, hasError := errorResult["error"]; hasError {
					return mcpgo.NewToolResultError(
						fmt.Sprintf("search error: %v", errorMsg)), nil
				}
				return mcpgo.NewToolResultJSON(errorResult)
			}
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to parse search results: %s", err.Error())), nil
		}

		// Wrap results in a map to match the expected response structure
		resultMap := map[string]interface{}{
			"results": searchResult,
		}

		return mcpgo.NewToolResultJSON(resultMap)
	}

	return mcpgo.NewTool(
		"search_docs",
		"Search payments documentation for specific terms",
		parameters,
		handler,
	)
}

// extractText takes HTML content as a string and returns extracted text content
func extractText(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent // Return original content if parsing fails
	}

	var textBuilder strings.Builder
	var extractTextNode func(*html.Node)

	// Recursively extract text from HTML nodes
	extractTextNode = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				textBuilder.WriteString(text)
				textBuilder.WriteString("\n")
			}
		}
		// Skip script and style tags
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractTextNode(c)
		}
	}

	extractTextNode(doc)
	return strings.TrimSpace(textBuilder.String()) + "\n"
}

// GetDocument returns a tool that fetches Razorpay documentation HTML content by path
func GetDocument(
	log *slog.Logger,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"doc_path",
			mcpgo.Description("Path to the document relative to the razorpay"),
			mcpgo.Required(),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		params := make(map[string]interface{})

		validator := NewValidator(&r).
			ValidateAndAddRequiredString(params, "doc_path")

		if result, err := validator.HandleErrorsIfAny(); result != nil {
			return result, err
		}

		docPath := params["doc_path"].(string)

		// Ensure path doesn't start with a slash to avoid double slashes in URL
		if len(docPath) > 0 && docPath[0] == '/' {
			docPath = docPath[1:]
		}

		// Create full URL by appending the path to the base URL
		fullURL := fmt.Sprintf("https://razorpay.com/docs/%s", docPath)

		// Create HTTP request to fetch the document
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fullURL,
			nil,
		)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to create request: %s", err.Error())), nil
		}

		// Add required headers
		req.Header.Set("sec-ch-ua-platform", "macOS")
		req.Header.Set("x-country-code", "IN")
		req.Header.Set("Referer", "https://razorpay.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
		req.Header.Set("sec-ch-ua", "\"Chromium\";v=\"136\", \"Google Chrome\";v=\"136\", \"Not.A/Brand\";v=\"99\"")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		req.Header.Set("sec-ch-ua-mobile", "?0")

		// Create HTTP client and send request
		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("document request failed: %s", err.Error())), nil
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("failed to read response: %s", err.Error())), nil
		}

		// Check if the response status code is not 200 OK
		if resp.StatusCode != http.StatusOK {
			return mcpgo.NewToolResultError(
				fmt.Sprintf("document fetch failed with status code %d: %s",
					resp.StatusCode, string(body))), nil
		}

		htmlContent := string(body)

		// Extract text from HTML content
		textContent := extractText(htmlContent)

		// Return both HTML and extracted text content in a JSON structure
		resultMap := map[string]interface{}{
			"url":      fullURL,
			"status":   resp.StatusCode,
			"content":  textContent, // Return extracted text as primary content
			"doc_path": docPath,
		}

		return mcpgo.NewToolResultJSON(resultMap)
	}

	return mcpgo.NewTool(
		"get_document_content",
		"Get the content of a specific razorpay document",
		parameters,
		handler,
	)
}
