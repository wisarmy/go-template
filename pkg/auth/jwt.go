package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	Issuer        string        `mapstructure:"issuer"`
	Expiration    time.Duration `mapstructure:"expiration"`
	RefreshExpire time.Duration `mapstructure:"refresh_expire"`
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID int, username, role string, config JWTConfig) (string, error) {
	now := time.Now()
	expireAt := now.Add(config.Expiration)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Secret))
}

// GenerateRefreshToken creates a new refresh token for a user
func GenerateRefreshToken(userID int, username string, config JWTConfig) (string, error) {
	now := time.Now()
	expireAt := now.Add(config.RefreshExpire)

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expireAt),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    config.Issuer,
		Subject:   fmt.Sprintf("%d", userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Secret))
}

// ParseToken parses and validates a JWT token
func ParseToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ParseRefreshToken parses and validates a refresh token
func ParseRefreshToken(tokenString string, secret string) (int, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrExpiredToken
		}
		return 0, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return 0, ErrInvalidToken
	}

	// Extract user ID from subject
	var userID int
	if _, err := fmt.Sscanf(claims.Subject, "%d", &userID); err != nil {
		return 0, ErrInvalidToken
	}

	return userID, nil
}
