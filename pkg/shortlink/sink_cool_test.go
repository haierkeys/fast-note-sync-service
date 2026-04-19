package shortlink

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSinkCoolClient_Create(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/link/create", r.URL.Path)
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		// Check decoding request
		var reqBody CreateRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com/long-url", reqBody.URL)
		assert.Equal(t, "my-pass", reqBody.Password)
		assert.Equal(t, "Test Title", reqBody.Title)

		// Mock response
		w.WriteHeader(http.StatusOK)
		res := CreateResponse{
			Slug:      "test-slug",
			URL:       reqBody.URL,
			ShortLink: "https://sink.cool/test-slug",
		}
		json.NewEncoder(w).Encode(res)
	}))
	defer mockServer.Close()

	// Use mock server URL
	client := NewSinkCoolClient(mockServer.URL, "test-api-key")
	
	// Test creating link
	expiresAt := time.Now().Add(24 * time.Hour)
	shortLink, err := client.Create("https://example.com/long-url", expiresAt, "my-pass", true, "Test Title")

	assert.NoError(t, err)
	assert.Equal(t, "https://sink.cool/test-slug", shortLink)
}

func TestSinkCoolClient_ErrorHandling(t *testing.T) {
	// Create a mock HTTP server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid url"}`))
	}))
	defer mockServer.Close()

	client := NewSinkCoolClient(mockServer.URL, "test-api-key")
	shortLink, err := client.Create("invalid", time.Time{}, "", false, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sink.cool api error: status=400")
	assert.Empty(t, shortLink)
}
