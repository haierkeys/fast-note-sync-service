package shortlink

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SinkCoolClient struct {
	BaseURL string
	APIKey  string
}

type CreateRequest struct {
	URL        string `json:"url"`
	Expiration int64  `json:"expiration,omitempty"` // Unix seconds
	Password   string `json:"password,omitempty"`
	Cloaking   bool   `json:"cloaking,omitempty"`
}

type CreateResponse struct {
	Slug      string `json:"slug"`
	URL       string `json:"url"`
	ShortLink string `json:"shortLink"`
}

func NewSinkCoolClient(baseURL, apiKey string) *SinkCoolClient {
	return &SinkCoolClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
}

func (c *SinkCoolClient) Create(url string, expiresAt time.Time, password string, cloaking bool) (string, error) {
	reqBody := CreateRequest{
		URL:      url,
		Password: password,
		Cloaking: cloaking,
	}
	if !expiresAt.IsZero() {
		reqBody.Expiration = expiresAt.Unix()
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	apiURL := fmt.Sprintf("%s/api/link/create", c.BaseURL)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("sink.cool api error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var res CreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.ShortLink, nil
}
