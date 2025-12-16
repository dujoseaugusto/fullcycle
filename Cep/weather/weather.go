package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &Client{BaseURL: baseURL, APIKey: apiKey, HTTPClient: httpClient}
}

type currentResp struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func (c *Client) GetCurrentTempC(query string) (float64, error) {
	u := fmt.Sprintf("%s/current.json?key=%s&q=%s", c.BaseURL, url.QueryEscape(c.APIKey), url.QueryEscape(query))
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weatherapi returned status %d", resp.StatusCode)
	}

	var r currentResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, err
	}
	return r.Current.TempC, nil
}
