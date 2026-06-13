//go:build oidc_integration

package oidc

import (
	"context"
	"encoding/json"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	internaloauth "github.com/haierkeys/fast-note-sync-service/internal/oauth"
)

func TestOIDCIntegrationProvider(t *testing.T) {
	method := getenv(t, "OIDC_INTEGRATION_METHOD")
	switch method {
	case "auth_code":
		testOIDCIntegrationAuthCode(t)
	case "password":
		testOIDCIntegrationPasswordGrant(t)
	default:
		t.Fatalf("unsupported OIDC_INTEGRATION_METHOD %q", method)
	}
}

func testOIDCIntegrationAuthCode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	issuer := getenv(t, "OIDC_INTEGRATION_ISSUER")
	clientID := getenv(t, "OIDC_INTEGRATION_CLIENT_ID")
	clientSecret := getenv(t, "OIDC_INTEGRATION_CLIENT_SECRET")
	redirectURL := getenv(t, "OIDC_INTEGRATION_REDIRECT_URL")
	login := getenv(t, "OIDC_INTEGRATION_LOGIN")
	password := getenv(t, "OIDC_INTEGRATION_PASSWORD")
	loginField := envOr("OIDC_INTEGRATION_LOGIN_FIELD", "login")

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	provider, err := NewProvider(ctx, ProviderConfig{
		Issuer:       issuer,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		HTTPClient:   httpClient,
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	nonce := "integration-nonce"
	codeVerifier := "integration-code-verifier-abcdefghijklmnopqrstuvwxyz0123456789"
	authURL := provider.AuthCodeURL("integration-state", nonce, codeVerifier)

	formHTML, formURL := followToForm(t, httpClient, authURL, redirectURL)
	action := parseFormAction(t, formHTML, formURL)
	callbackURL := submitLoginForm(t, httpClient, action, loginField, login, password, redirectURL)
	code := callbackURL.Query().Get("code")
	if code == "" {
		t.Fatalf("callback missing code: %s", callbackURL.String())
	}

	claims, err := provider.Exchange(ctx, code, codeVerifier, nonce)
	if err != nil {
		t.Fatalf("Exchange() error = %v", err)
	}
	if claims.Subject == "" {
		t.Fatalf("claims subject is empty: %#v", claims)
	}
}

func testOIDCIntegrationPasswordGrant(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	issuer := getenv(t, "OIDC_INTEGRATION_ISSUER")
	clientID := getenv(t, "OIDC_INTEGRATION_CLIENT_ID")
	clientSecret := getenv(t, "OIDC_INTEGRATION_CLIENT_SECRET")
	tokenURL := getenv(t, "OIDC_INTEGRATION_TOKEN_URL")
	username := getenv(t, "OIDC_INTEGRATION_LOGIN")
	password := getenv(t, "OIDC_INTEGRATION_PASSWORD")
	jwksURL := getenv(t, "OIDC_INTEGRATION_JWKS_URL")

	values := url.Values{}
	values.Set("grant_type", "password")
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSecret)
	values.Set("username", username)
	values.Set("password", password)
	values.Set("scope", "openid profile email")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("token endpoint status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		t.Fatal(err)
	}
	if tokenResp.IDToken == "" {
		t.Fatalf("token response missing id_token: %s", string(body))
	}

	verifier := internaloauth.NewJWTVerifier(internaloauth.JWTVerifierConfig{
		Issuer:   issuer,
		Audience: clientID,
		JWKSURL:  jwksURL,
	})
	claims, err := verifier.Verify(ctx, tokenResp.IDToken)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.Subject == "" {
		t.Fatalf("claims subject is empty: %#v", claims)
	}
}

func followToForm(t *testing.T, client *http.Client, rawURL string, redirectURL string) (string, *url.URL) {
	t.Helper()

	current := rawURL
	for i := 0; i < 10; i++ {
		resp, err := client.Get(current)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			next := resp.Header.Get("Location")
			if strings.HasPrefix(next, redirectURL) {
				t.Fatalf("unexpected callback before login: %s", next)
			}
			base, _ := url.Parse(current)
			nextURL, err := base.Parse(next)
			if err != nil {
				t.Fatal(err)
			}
			current = nextURL.String()
			continue
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected login form status 200, got %d: %s", resp.StatusCode, string(body))
		}
		parsed, _ := url.Parse(current)
		return string(body), parsed
	}
	t.Fatal("too many redirects before login form")
	return "", nil
}

var formActionPattern = regexp.MustCompile(`<form[^>]+action="([^"]+)"`)

func parseFormAction(t *testing.T, body string, formURL *url.URL) string {
	t.Helper()

	matches := formActionPattern.FindStringSubmatch(body)
	if len(matches) != 2 {
		t.Fatalf("login form action not found")
	}
	action := html.UnescapeString(matches[1])
	actionURL, err := formURL.Parse(action)
	if err != nil {
		t.Fatal(err)
	}
	return actionURL.String()
}

func submitLoginForm(t *testing.T, client *http.Client, action string, loginField string, login string, password string, redirectURL string) *url.URL {
	t.Helper()

	values := url.Values{}
	values.Set(loginField, login)
	values.Set("password", password)

	resp, err := client.PostForm(action, values)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		t.Fatalf("login submit status %d: %s", resp.StatusCode, string(body))
	}
	location := resp.Header.Get("Location")
	if !strings.HasPrefix(location, redirectURL) {
		t.Fatalf("login redirect = %q, want callback %q", location, redirectURL)
	}
	callbackURL, err := url.Parse(location)
	if err != nil {
		t.Fatal(err)
	}
	if errValue := callbackURL.Query().Get("error"); errValue != "" {
		t.Fatalf("callback error: %s", errValue)
	}
	return callbackURL
}

func getenv(t *testing.T, key string) string {
	t.Helper()

	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		t.Fatalf("%s is required", key)
	}
	return value
}

func envOr(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
