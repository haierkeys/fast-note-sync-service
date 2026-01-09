package app

import (
	"fmt"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TokenConfig 定义 Token 管理器的配置
type TokenConfig struct {
	SecretKey string        `yaml:"secret-key"` // JWT 签名密钥
	Expiry    time.Duration `yaml:"expiry"`     // Token 过期时间，默认 7 天
	Issuer    string        `yaml:"issuer"`     // Token 签发者
}

// TokenManager 定义 Token 管理接口
type TokenManager interface {
	Generate(uid int64, nickname, ip string) (string, error)
	Parse(token string) (*UserEntity, error)
	Validate(token string) error
}

// tokenManager 实现 TokenManager 接口
type tokenManager struct {
	config TokenConfig
}

// NewTokenManager 创建一个新的 TokenManager 实例
func NewTokenManager(cfg TokenConfig) TokenManager {
	// 设置默认值
	if cfg.Expiry == 0 {
		cfg.Expiry = 7 * 24 * time.Hour // 默认 7 天
	}
	if cfg.Issuer == "" {
		cfg.Issuer = global.Name
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

// Validate 验证 Token 是否有效
func (t *tokenManager) Validate(token string) error {
	_, err := t.Parse(token)
	return err
}

// ParseToken parses a JWT token and returns the user data.
// 保留原有函数以保持向后兼容
func ParseToken(tokenString string) (*UserEntity, error) {
	// Initialize a new instance of `Claims`
	claims := &UserEntity{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set, or if the signature does not match).
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(global.Config.Security.AuthTokenKey + "_" + util.GetMachineID()), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// GenerateToken generates a new JWT token for a user.
// 保留原有函数以保持向后兼容
func GenerateToken(uid int64, nickname string, ip string, expiry int64) (string, error) {
	// Create the Claims
	expirationTime := time.Now().Add(time.Duration(expiry) * time.Second).Unix()
	claims := &UserEntity{
		UID:      uid,
		Nickname: nickname,
		IP:       ip,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(time.Unix(expirationTime, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    global.Name,
			Subject:   "user-token",
			ID:        fmt.Sprintf("%d", uid), // Use UID as unique token ID
		},
	}
	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString([]byte(global.Config.Security.AuthTokenKey + "_" + util.GetMachineID()))
	if err != nil {
		return "", err
	}

	return tokenString, nil
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

// SetTokenToContext set token to gin.Context
func SetTokenToContext(ctx *gin.Context, tokenString string) error {
	user, err := ParseToken(tokenString)
	if err != nil {
		return err
	}
	ctx.Set("user_token", user)
	return nil
}
