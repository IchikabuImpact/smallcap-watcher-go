package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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
	return &Client{
		BaseURL: "https://jpx.pinkgold.space",
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) FetchStockData(ctx context.Context, ticker string) (StockResponse, error) {
	url := fmt.Sprintf("%s/scrape?ticker=%s", c.BaseURL, ticker)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return StockResponse{}, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return StockResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return StockResponse{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var payload StockResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return StockResponse{}, err
	}

	return payload, nil
}
