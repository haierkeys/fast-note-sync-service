package oauth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTVerifier_Verify_validRSAJWKSAccessToken(t *testing.T) {
	key := newTestRSAKey(t)
	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testJWKS("test-key", &key.PublicKey)))
	}))
	defer jwks.Close()

	verifier := NewJWTVerifier(JWTVerifierConfig{
		Issuer:         "https://issuer.example.com",
		Audience:       "fns-mcp",
		JWKSURL:        jwks.URL,
		RequiredScopes: []string{"notes:read"},
	})

	claims, err := verifier.Verify(context.Background(), newTestJWT(t, key, "test-key", jwt.MapClaims{
		"iss":   "https://issuer.example.com",
		"aud":   "fns-mcp",
		"sub":   "subject-1",
		"email": "user@example.com",
		"scope": "notes:read files:read",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"nbf":   time.Now().Add(-time.Minute).Unix(),
		"iat":   time.Now().Add(-time.Minute).Unix(),
	}))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.Subject != "subject-1" {
		t.Fatalf("Verify() subject = %q, want subject-1", claims.Subject)
	}
	if claims.Raw["email"] != "user@example.com" {
		t.Fatalf("Verify() email claim = %v, want user@example.com", claims.Raw["email"])
	}
}

func TestJWTVerifier_Verify_acceptsAuth0PermissionsClaim(t *testing.T) {
	key := newTestRSAKey(t)
	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testJWKS("test-key", &key.PublicKey)))
	}))
	defer jwks.Close()

	verifier := NewJWTVerifier(JWTVerifierConfig{
		Issuer:         "https://issuer.example.com/",
		Audience:       "https://example.test/api/mcp",
		JWKSURL:        jwks.URL,
		RequiredScopes: []string{"notes:read"},
	})

	claims, err := verifier.Verify(context.Background(), newTestJWT(t, key, "test-key", jwt.MapClaims{
		"iss":         "https://issuer.example.com/",
		"aud":         "https://example.test/api/mcp",
		"sub":         "subject-1",
		"permissions": []interface{}{"notes:read", "files:read"},
		"exp":         time.Now().Add(time.Hour).Unix(),
		"nbf":         time.Now().Add(-time.Minute).Unix(),
		"iat":         time.Now().Add(-time.Minute).Unix(),
	}))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(claims.Scopes) != 2 {
		t.Fatalf("Verify() scopes = %v, want Auth0 permissions", claims.Scopes)
	}
}

func TestJWTVerifier_Verify_unknownKidReturnsInvalidToken(t *testing.T) {
	key := newTestRSAKey(t)
	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testJWKS("different-key", &key.PublicKey)))
	}))
	defer jwks.Close()

	verifier := NewJWTVerifier(JWTVerifierConfig{
		Issuer:   "https://issuer.example.com",
		Audience: "fns-mcp",
		JWKSURL:  jwks.URL,
	})

	_, err := verifier.Verify(context.Background(), newTestJWT(t, key, "test-key", jwt.MapClaims{
		"iss": "https://issuer.example.com",
		"aud": "fns-mcp",
		"sub": "subject-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	}))
	if !IsInvalidToken(err) {
		t.Fatalf("Verify() error = %v, want invalid token", err)
	}
}

func TestJWTVerifier_Verify_missingRequiredScopeReturnsInsufficientScope(t *testing.T) {
	key := newTestRSAKey(t)
	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testJWKS("test-key", &key.PublicKey)))
	}))
	defer jwks.Close()

	verifier := NewJWTVerifier(JWTVerifierConfig{
		Issuer:         "https://issuer.example.com",
		Audience:       "fns-mcp",
		JWKSURL:        jwks.URL,
		RequiredScopes: []string{"notes:write"},
	})

	_, err := verifier.Verify(context.Background(), newTestJWT(t, key, "test-key", jwt.MapClaims{
		"iss":   "https://issuer.example.com",
		"aud":   "fns-mcp",
		"sub":   "subject-1",
		"scope": "notes:read",
		"exp":   time.Now().Add(time.Hour).Unix(),
	}))
	if !errors.Is(err, ErrInsufficientScope) {
		t.Fatalf("Verify() error = %v, want ErrInsufficientScope", err)
	}
}

func newTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	return key
}

func newTestJWT(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return signed
}

func testJWKS(kid string, key *rsa.PublicKey) string {
	n := base64.RawURLEncoding.EncodeToString(key.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())
	return `{"keys":[{"kty":"RSA","use":"sig","kid":"` + kid + `","alg":"RS256","n":"` + n + `","e":"` + e + `"}]}`
}
