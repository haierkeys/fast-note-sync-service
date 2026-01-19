package app

import (
	"bytes"
	"fmt"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"crypto/aes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 默认 Token 签发者
const DefaultTokenIssuer = "fast-note-sync-service"

// TokenConfig 定义 Token 管理器的配置
type TokenConfig struct {
	SecretKey     string        `yaml:"secret-key"`         // JWT 签名密钥
	Expiry        time.Duration `yaml:"expiry"`             // Token 过期时间，默认 365 天
	ShareTokenKey string        `yaml:"share-token-key"`    // 分享专用签名密钥
	ShareExpiry   time.Duration `yaml:"share-token-expiry"` // 分享专用过期时间
	Issuer        string        `yaml:"issuer"`             // Token 签发者
}

// TokenManager 定义 Token 管理接口
type TokenManager interface {
	// 用户认证相关
	Generate(uid int64, nickname, ip string) (string, error)
	Parse(token string) (*UserEntity, error)

	// 资源分享相关
	ShareGenerate(shareID int64, uid int64, resources map[string][]string) (string, error)
	ShareParse(token string) (*ShareEntity, error)

	Validate(token string) error
	GetSecretKey() string
}

// tokenManager 实现 TokenManager 接口
type tokenManager struct {
	config TokenConfig
}

// NewTokenManager 创建一个新的 TokenManager 实例
func NewTokenManager(cfg TokenConfig) TokenManager {
	// 设置默认值
	if cfg.Expiry == 0 {
		cfg.Expiry = 365 * 24 * time.Hour // 默认 365 天
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
	IP       string `json:"ip"`
	jwt.RegisteredClaims
}

// ShareEntity 资源分享 Claims
type ShareEntity struct {
	SID       int64               `json:"sid"`       // 数据库中的分享记录 ID (Share ID)
	UID       int64               `json:"uid"`       // 数据库中的用户 ID (User ID)
	Resources map[string][]string `json:"resources"` // 资源列表
	ExpiresAt time.Time           `json:"exp"`
}

// Generate 生成一个新的 JWT Token
func (t *tokenManager) Generate(uid int64, nickname, ip string) (string, error) {
	expirationTime := time.Now().Add(t.config.Expiry)
	claims := &UserEntity{
		UID:      uid,
		Nickname: nickname,
		IP:       ip,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
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

// ShareGenerate 构建分享 Token (极致缩短版: 单块 AES + Checksum)
func (t *tokenManager) ShareGenerate(shareID int64, uid int64, resources map[string][]string) (string, error) {
	expirationTime := time.Unix(time.Now().Add(t.config.ShareExpiry).Unix(), 0)

	// 准备数据 (刚好 16 字节): SID (8) + ExpiresAt (4) + Checksum (4)
	data := make([]byte, 16)
	binary.BigEndian.PutUint64(data[0:8], uint64(shareID))
	binary.BigEndian.PutUint32(data[8:12], uint32(expirationTime.Unix()))

	// 生成校验和: 使用 Key + SID + Exp 生成摘要，取前 4 字节
	key := sha256.Sum256([]byte(t.config.ShareTokenKey + "_" + util.GetMachineID()))
	h := sha256.New()
	h.Write(key[:])
	h.Write(data[0:12])
	sum := h.Sum(nil)
	copy(data[12:16], sum[:4])

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	// 执行单块加密 (16 字节)
	ciphertext := make([]byte, 16)
	block.Encrypt(ciphertext, data)

	// 使用 RawURLEncoding 得到固定的 22 字符长度
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// ShareParse 解析分享 Token
func (t *tokenManager) ShareParse(tokenString string) (*ShareEntity, error) {
	ciphertext, err := base64.RawURLEncoding.DecodeString(tokenString)
	if err != nil || len(ciphertext) != 16 {
		return nil, fmt.Errorf("invalid token format")
	}

	// 使用 ShareTokenKey + MachineID 生成 AES Key
	key := sha256.Sum256([]byte(t.config.ShareTokenKey + "_" + util.GetMachineID()))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	// 执行单块解密
	data := make([]byte, 16)
	block.Decrypt(data, ciphertext)

	// 验证校验和
	h := sha256.New()
	h.Write(key[:])
	h.Write(data[0:12])
	sum := h.Sum(nil)

	if !bytes.Equal(data[12:16], sum[:4]) {
		return nil, fmt.Errorf("invalid token checksum")
	}

	shareID := int64(binary.BigEndian.Uint64(data[0:8]))
	expUnix := int64(binary.BigEndian.Uint32(data[8:12]))

	return &ShareEntity{
		SID:       shareID,
		ExpiresAt: time.Unix(expUnix, 0),
	}, nil
}

// Validate 验证 Token 是否有效
func (t *tokenManager) Validate(token string) error {
	_, err := t.Parse(token)
	return err
}

// GetSecretKey 获取密钥
func (t *tokenManager) GetSecretKey() string {
	return t.config.SecretKey
}

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
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// GetUid extracts the user ID from the request context.
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
func GetIP(ctx *gin.Context) (out string) {
	user, exist := ctx.Get("user_token")
	if exist {
		if userEntity, ok := user.(*UserEntity); ok {
			out = userEntity.IP
		}
	}
	return
}

// SetTokenToContextWithKey 使用指定密钥设置 Token 到 Context
func SetTokenToContextWithKey(ctx *gin.Context, tokenString string, secretKey string) error {
	user, err := ParseTokenWithKey(tokenString, secretKey)
	if err != nil {
		return err
	}
	ctx.Set("user_token", user)
	return nil
}
