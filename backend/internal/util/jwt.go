package util

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ld/storeinventory/internal/constants"
)

// JWTClaims 自定义 JWT 声明，注入用户身份信息。
type JWTClaims struct {
	UserID   uint               `json:"user_id"`
	Username string             `json:"username"`
	Role     constants.UserRole `json:"role"`
	StoreID  *uint              `json:"store_id,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken 签发 JWT。
func GenerateToken(secret string, ttlHours int, userID uint, username string, role constants.UserRole, storeID *uint) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		StoreID:  storeID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttlHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ldstoreinventory",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 校验 JWT 并解析声明。
func ParseToken(secret, tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}
