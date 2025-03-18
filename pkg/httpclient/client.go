package httpclient

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func Get(url string) {

	httpReq, _ := http.NewRequest("GET", url, nil)
	httpReq.Header.Add("Content-type", "application/json")
	httpReq.Host = "www.example.com"
}

func Post(postURL string, postData map[string][]string) (string, error) {

	data := url.Values(postData)

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(
		postURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), err
}
