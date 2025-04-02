package app

import (
	"fmt"
	"time"

	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// UserEntity represents the user data stored in the JWT.
type UserEntity struct {
	UID      int64  `json:"uid"`
	Nickname string `json:"nickname"`
	IP       string `json:"ip"`
	jwt.RegisteredClaims
}

// ParseToken parses a JWT token and returns the user data.
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
