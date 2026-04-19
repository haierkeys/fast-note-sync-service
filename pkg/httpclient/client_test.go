package httpclient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	// Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "foo=bar", string(body))

		w.Write([]byte("success"))
	}))
	defer ts.Close()

	data := map[string][]string{
		"foo": {"bar"},
	}

	resp, err := Post(ts.URL, data)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp)
}
