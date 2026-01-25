package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "http://127.0.0.1:8085"
const maxRetries = 3
const retryBaseDelay = 750 * time.Millisecond
const maxErrorBody = 2048

type StockResponse struct {
	Ticker        string `json:"ticker"`
	CompanyName   string `json:"companyName"`
	CurrentPrice  string `json:"currentPrice"`
	PreviousClose string `json:"previousClose"`
	DividendYield string `json:"dividendYield"`
	PER           string `json:"per"`
	PBR           string `json:"pbr"`
	MarketCap     string `json:"marketCap"`
	Volume        string `json:"volume"`
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient() *Client {
	return NewClientWithBaseURL(defaultBaseURL)
}

func NewClientWithBaseURL(baseURL string) *Client {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		trimmed = defaultBaseURL
	}
	trimmed = strings.TrimRight(trimmed, "/")
	return &Client{
		BaseURL: trimmed,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) FetchStockData(ctx context.Context, ticker string) (StockResponse, error) {
	url := fmt.Sprintf("%s/scrape?ticker=%s", c.BaseURL, ticker)
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return StockResponse{}, err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			if !shouldRetry(attempt, 0) {
				break
			}
			time.Sleep(retryDelay(attempt))
			continue
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		resp.Body.Close()
		if readErr != nil {
			return StockResponse{}, readErr
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("unexpected status: %s for %s (body: %q)", resp.Status, url, strings.TrimSpace(string(body)))
			if !shouldRetry(attempt, resp.StatusCode) {
				break
			}
			time.Sleep(retryDelay(attempt))
			continue
		}

		var payload StockResponse
		if err := json.NewDecoder(strings.NewReader(string(body))).Decode(&payload); err != nil {
			return StockResponse{}, err
		}

		return payload, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("request failed for %s", url)
	}
	return StockResponse{}, lastErr
}

func shouldRetry(attempt int, statusCode int) bool {
	if attempt >= maxRetries {
		return false
	}
	switch statusCode {
	case 0, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func retryDelay(attempt int) time.Duration {
	return retryBaseDelay * time.Duration(1<<attempt)
}
