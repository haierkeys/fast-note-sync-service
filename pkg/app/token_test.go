package app

import (
	"testing"
	"time"
)

func TestTokenManager_ShareGenerateAndParse(t *testing.T) {
	cfg := TokenConfig{
		SecretKey:     "user-secret",
		ShareTokenKey: "share-secret",
		ShareExpiry:   1 * time.Hour,
		Issuer:        "test-issuer",
	}
	tm := NewTokenManager(cfg)

	// 测试数据
	shareID := int64(12345)
	uid := int64(1001)
	resources := map[string][]string{
		"note": {"note_id_1", "note_id_2"},
		"file": {"file_id_1"},
	}

	// 1. 测试生成和解析
	token, err := tm.ShareGenerate(shareID, uid, resources)
	if err != nil {
		t.Fatalf("ShareGenerate failed: %v", err)
	}

	parsedClaims, err := tm.ShareParse(token)
	if err != nil {
		t.Fatalf("ShareParse failed: %v", err)
	}

	// 验证 SID
	if parsedClaims.SID != shareID {
		t.Errorf("Expected SID %d, got %d", shareID, parsedClaims.SID)
	}

	// 验证 ExpiresAt (由于只存了秒级 Unix 戳，允许 1 秒内的误差)
	now := time.Now()
	expectedExp := now.Add(cfg.ShareExpiry)
	if parsedClaims.ExpiresAt.Unix() < expectedExp.Unix()-1 || parsedClaims.ExpiresAt.Unix() > expectedExp.Unix()+1 {
		t.Errorf("Expected ExpiresAt around %v, got %v", expectedExp, parsedClaims.ExpiresAt)
	}

	// 3. 测试错误的密钥
	wrongKeyCfg := cfg
	wrongKeyCfg.ShareTokenKey = "wrong-secret"
	tmWrongKey := NewTokenManager(wrongKeyCfg)

	wrongToken, _ := tmWrongKey.ShareGenerate(shareID, uid, resources)
	_, err = tm.ShareParse(wrongToken)
	if err == nil {
		t.Error("Expected error when parsing token with wrong secret key, but got nil")
	}

	// 4. 测试篡改后的 Token
	tamperedToken := token + "tampered"
	_, err = tm.ShareParse(tamperedToken)
	if err == nil {
		t.Error("Expected error for tampered token, but got nil")
	}
}

func TestTokenManager_GenerateAndParse(t *testing.T) {
	cfg := TokenConfig{
		SecretKey: "user-secret",
		Expiry:    24 * time.Hour,
		Issuer:    "user-issuer",
	}
	tm := NewTokenManager(cfg)

	uid := int64(1001)
	nickname := "testuser"
	ip := "127.0.0.1"

	// 1. 测试生成和解析
	token, err := tm.Generate(uid, nickname, ip)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	parsedUser, err := tm.Parse(token)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// 验证字段
	if parsedUser.UID != uid {
		t.Errorf("Expected UID %d, got %d", uid, parsedUser.UID)
	}
	if parsedUser.Nickname != nickname {
		t.Errorf("Expected Nickname %s, got %s", nickname, parsedUser.Nickname)
	}
	if parsedUser.IP != ip {
		t.Errorf("Expected IP %s, got %s", ip, parsedUser.IP)
	}
	if parsedUser.Issuer != cfg.Issuer {
		t.Errorf("Expected Issuer %s, got %s", cfg.Issuer, parsedUser.Issuer)
	}

	// 2. 测试过期
	shortExpiryCfg := cfg
	shortExpiryCfg.Expiry = -1 * time.Second
	tmExpired := NewTokenManager(shortExpiryCfg)

	expiredToken, err := tmExpired.Generate(uid, nickname, ip)
	if err != nil {
		t.Fatalf("Generate (expired) failed: %v", err)
	}

	_, err = tm.Parse(expiredToken)
	if err == nil {
		t.Error("Expected error for expired token, but got nil")
	}

	// 3. 测试错误的密钥
	wrongKeyCfg := cfg
	wrongKeyCfg.SecretKey = "wrong-user-secret"
	tmWrongKey := NewTokenManager(wrongKeyCfg)

	wrongToken, _ := tmWrongKey.Generate(uid, nickname, ip)
	_, err = tm.Parse(wrongToken)
	if err == nil {
		t.Error("Expected error for token generated with different secret key, but got nil")
	}

	// 4. 测试篡改后的 Token
	tamperedToken := token + "xyz"
	_, err = tm.Parse(tamperedToken)
	if err == nil {
		t.Error("Expected error for tampered user token, but got nil")
	}
}
