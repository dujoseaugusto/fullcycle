package viacep

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &Client{BaseURL: baseURL, HTTPClient: httpClient}
}

type Address struct {
	Localidade string `json:"localidade"`
	Uf         string `json:"uf"`
	Erro       bool   `json:"erro,omitempty"`
}

func (c *Client) GetAddress(cep string) (*Address, error) {
	u := fmt.Sprintf("%s/%s/json/", c.BaseURL, url.PathEscape(cep))
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("viacep returned status %d", resp.StatusCode)
	}

	var a Address
	if err := json.NewDecoder(resp.Body).Decode(&a); err != nil {
		return nil, err
	}

	if a.Erro {
		return nil, nil
	}
	return &a, nil
}
