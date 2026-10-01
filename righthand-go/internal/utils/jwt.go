package utils

import (
	"errors"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"righthand-go/internal/config"
)

// Claims JWT声明结构
type Claims struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	RealName string `json:"realName"`
	IsAdmin  bool   `json:"isAdmin"`
	Roles    string `json:"roles"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token
func GenerateToken(userID int64, username, realName string, isAdmin bool, roles string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(config.GlobalConfig.JWT.Expire) * time.Second)
	
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RealName: realName,
		IsAdmin:  isAdmin,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    config.GlobalConfig.JWT.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GlobalConfig.JWT.Secret))
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GlobalConfig.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateToken 验证Token是否有效
func ValidateToken(tokenString string) bool {
	_, err := ParseToken(tokenString)
	return err == nil
}

// GetUserIDFromToken 从Token中获取用户ID
func GetUserIDFromToken(tokenString string) (int64, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// GetUsernameFromToken 从Token中获取用户名
func GetUsernameFromToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.Username, nil
}
