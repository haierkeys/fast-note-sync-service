package app

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"crypto/aes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// DefaultTokenIssuer default Token issuer // 默认 Token 签发者
const DefaultTokenIssuer = "fast-note-sync-service"

// TokenConfig defines Token manager configuration // TokenConfig 定义 Token 管理器的配置
type TokenConfig struct {
	SecretKey     string        `yaml:"secret-key"`         // JWT signing key // JWT 签名密钥
	Expiry        time.Duration `yaml:"expiry"`             // Token expiration time, defaults to 365 days // Token 过期时间，默认 365 天
	ShareTokenKey string        `yaml:"share-token-key"`    // Dedicated signing key for sharing // 分享专用签名密钥
	ShareExpiry   time.Duration `yaml:"share-token-expiry"` // Dedicated expiration time for sharing // 分享专用过期时间
	Issuer        string        `yaml:"issuer"`             // Token issuer // Token 签发者
}

// TokenManager defines Token management interface // TokenManager 定义 Token 管理接口
type TokenManager interface {
	// User authentication related // 用户认证相关
	Generate(uid int64, nickname, ip string, tokenID int64, nonce string) (string, error)
	Parse(token string) (*UserEntity, error)

	// Resource sharing related // 资源分享相关
	ShareGenerate(shareID int64, uid int64, resources map[string][]string) (string, error)
	ShareParse(token string) (*ShareEntity, error)

	Validate(token string) error
	GetSecretKey() string
}

// tokenManager implementation of TokenManager interface // tokenManager 实现 TokenManager 接口
type tokenManager struct {
	config TokenConfig
}

// NewTokenManager creates a new TokenManager instance
// NewTokenManager 创建一个新的 TokenManager 实例
func NewTokenManager(cfg TokenConfig) TokenManager {
	// Set default values
	// 设置默认值
	if cfg.Expiry == 0 {
		cfg.Expiry = 365 * 24 * time.Hour // Default 365 days // 默认 365 天
	}
	if cfg.Issuer == "" {
		cfg.Issuer = DefaultTokenIssuer
	}
	return &tokenManager{config: cfg}
}

// UserSelectEntity represents the user data stored in the JWT.
type UserSelectEntity struct {
	UID      int64  `json:"uid"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type UserEntity struct {
	UID      int64  `json:"uid"`
	Nickname string `json:"nickname"`
	TokenID  int64  `json:"tokenId"` // 数据库中的 auth_token.id
	Nonce    string `json:"nonce"`   // 令牌标识符，用于轮换校验
	jwt.RegisteredClaims
}

// ShareEntity resource sharing Claims // ShareEntity 资源分享 Claims
type ShareEntity struct {
	SID       int64               `json:"sid"`       // Share record ID in database // 数据库中的分享记录 ID (Share ID)
	UID       int64               `json:"uid"`       // User ID in database // 数据库中的用户 ID (User ID)
	Resources map[string][]string `json:"resources"` // Resource list // 资源列表
	ExpiresAt time.Time           `json:"exp"`
}

// Generate generates a new JWT Token
func (t *tokenManager) Generate(uid int64, nickname, _ string, tokenID int64, nonce string) (string, error) {
	claims := &UserEntity{
		UID:      uid,
		Nickname: nickname,
		TokenID:  tokenID,
		Nonce:    nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.config.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    t.config.Issuer,
			Subject:   "user-token",
			ID:        fmt.Sprintf("%d", uid),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.config.SecretKey + "_" + util.GetMachineID()))
}

// Parse parses JWT Token and returns user info
// Parse 解析 JWT Token 并返回用户信息
func (t *tokenManager) Parse(token string) (*UserEntity, error) {
	claims := &UserEntity{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.config.SecretKey + "_" + util.GetMachineID()), nil
	})

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ShareGenerate builds share Token using HMAC-SHA256 (54 characters)
// ShareGenerate 构建分享 Token (使用 HMAC-SHA256 算法产生 54 字符)
func (t *tokenManager) ShareGenerate(shareID int64, uid int64, resources map[string][]string) (string, error) {
	expiresAt := time.Now().Add(t.config.ShareExpiry).Unix()

	// Prepare payload (24 bytes): SID (8 bytes) + UID (8 bytes) + ExpiresAt (8 bytes)
	// 准备 payload (24 字节): SID (8 字节) + UID (8 字节) + ExpiresAt (8 字节)
	payload := make([]byte, 24)
	binary.BigEndian.PutUint64(payload[0:8], uint64(shareID))
	binary.BigEndian.PutUint64(payload[8:16], uint64(uid))
	binary.BigEndian.PutUint64(payload[16:24], uint64(expiresAt))

	// Generate HMAC-SHA256 tag and truncate to first 16 bytes (128 bit tag)
	// 生成 HMAC-SHA256 摘要并取前 16 字节作为签名 (128 bit 标签)
	key := sha256.Sum256([]byte(t.config.ShareTokenKey + "_" + util.GetMachineID()))
	mac := hmac.New(sha256.New, key[:])
	mac.Write(payload)
	tag := mac.Sum(nil)[:16]

	// Combine payload and tag, encode using base64 RawURLEncoding (resulting in 54 chars)
	// 组合 payload 和 tag 并使用 Base64 RawURLEncoding 编码 (得到 54 字符)
	combined := append(payload, tag...)
	return base64.RawURLEncoding.EncodeToString(combined), nil
}

// ShareParse parses share Token with compatibility fallback
// ShareParse 解析分享 Token，支持兼容旧版
func (t *tokenManager) ShareParse(tokenString string) (*ShareEntity, error) {
	data, err := base64.RawURLEncoding.DecodeString(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token format")
	}

	// 1. If length matches 40 bytes, try parsing with HMAC-SHA256 (new version)
	// 1. 如果长度为 40 字节，尝试用新版 HMAC-SHA256 校验和解析
	if len(data) == 40 {
		payload, tag := data[:24], data[24:]
		key := sha256.Sum256([]byte(t.config.ShareTokenKey + "_" + util.GetMachineID()))
		mac := hmac.New(sha256.New, key[:])
		mac.Write(payload)
		expected := mac.Sum(nil)[:16]

		if hmac.Equal(tag, expected) {
			shareID := int64(binary.BigEndian.Uint64(payload[0:8]))
			uid := int64(binary.BigEndian.Uint64(payload[8:16]))
			expiresAt := int64(binary.BigEndian.Uint64(payload[16:24]))

			if time.Now().Unix() > expiresAt {
				return nil, fmt.Errorf("token expired")
			}
			return &ShareEntity{
				SID:       shareID,
				UID:       uid,
				ExpiresAt: time.Unix(expiresAt, 0),
			}, nil
		}
	}

	// 2. If length matches 16 bytes, fallback to parsing with AES-ECB (old version)
	// 2. 如果长度为 16 字节，回退用旧版 AES-ECB 算法解密与校验
	if len(data) == 16 {
		key := sha256.Sum256([]byte(t.config.ShareTokenKey + "_" + util.GetMachineID()))
		block, err := aes.NewCipher(key[:])
		if err == nil {
			decrypted := make([]byte, 16)
			block.Decrypt(decrypted, data)

			// Verify old checksum
			// 校验旧校验和
			h := sha256.New()
			h.Write(key[:])
			h.Write(decrypted[0:13])
			sum := h.Sum(nil)

			if bytes.Equal(decrypted[13:16], sum[:3]) {
				// Parse SID (6 bytes)
				// 解析 SID (6 字节)
				sidBytes := make([]byte, 8)
				copy(sidBytes[2:8], decrypted[0:6])
				shareID := int64(binary.BigEndian.Uint64(sidBytes))

				// Parse UID (3 bytes)
				// 解析 UID (3 字节)
				uidBytes := make([]byte, 8)
				copy(uidBytes[5:8], decrypted[6:9])
				uid := int64(binary.BigEndian.Uint64(uidBytes))

				// Parse ExpiresAt (4 bytes)
				// 解析 ExpiresAt (4 字节)
				expUnix := int64(binary.BigEndian.Uint32(decrypted[9:13]))

				if time.Now().Unix() > expUnix {
					return nil, fmt.Errorf("token expired")
				}
				return &ShareEntity{
					SID:       shareID,
					UID:       uid,
					ExpiresAt: time.Unix(expUnix, 0),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("invalid token signature or fallback failed")
}

// Validate validates if Token is valid
// Validate 验证 Token 是否有效
func (t *tokenManager) Validate(token string) error {
	_, err := t.Parse(token)
	return err
}

// GetSecretKey gets secret key
// GetSecretKey 获取密钥
func (t *tokenManager) GetSecretKey() string {
	return t.config.SecretKey
}

// ParseTokenWithKey parses Token with specified key
// ParseTokenWithKey 使用指定密钥解析 Token
func ParseTokenWithKey(tokenString string, secretKey string) (*UserEntity, error) {
	claims := &UserEntity{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey + "_" + util.GetMachineID()), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, code.ErrorTokenExpired
		}
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// GetTokenID extracts the token ID from the request context.
func GetTokenID(ctx *gin.Context) (out int64) {
	user, exist := ctx.Get("user_token")
	if exist {
		if userEntity, ok := user.(*UserEntity); ok {
			out = userEntity.TokenID
		}
	}
	return
}

// GetUID extracts the user ID from the request context.
func GetUID(ctx *gin.Context) (out int64) {
	user, exist := ctx.Get("user_token")
	if exist {
		if userEntity, ok := user.(*UserEntity); ok {
			out = userEntity.UID
		}
	}
	return
}

// GetShareEntity extracts the share entity from the request context.
func GetShareEntity(ctx *gin.Context) (out *ShareEntity) {
	user, exist := ctx.Get("share_entity")
	if exist {
		if shareEntity, ok := user.(*ShareEntity); ok {
			out = shareEntity
		}
	}
	return
}

// GetIP extracts the user IP from the request context.
// Deprecated: IP is now managed statefully in the database.
func GetIP(ctx *gin.Context) (out string) {
	return ""
}

// SetTokenToContextWithKey sets Token to Context with specified key
// SetTokenToContextWithKey 使用指定密钥设置 Token 到 Context
func SetTokenToContextWithKey(ctx *gin.Context, tokenString string, secretKey string) error {
	user, err := ParseTokenWithKey(tokenString, secretKey)
	if err != nil {
		return err
	}
	ctx.Set("user_token", user)
	return nil
}
