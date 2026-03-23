package auth

import (
	"errors"
	"strings"
	"time"

	"gochat/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int64, username string) (string, error) {
	cfg := config.GetConfig()
	expireHours := cfg.JwtConfig.ExpireHours
	if expireHours <= 0 {
		expireHours = 72
	}
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JwtConfig.Secret))
}

func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.GetConfig()
	if strings.TrimSpace(cfg.JwtConfig.Secret) == "" {
		return nil, errors.New("jwt secret missing")
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JwtConfig.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ExtractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
			return
		}
		if IsTokenRevoked(token) {
			c.AbortWithStatusJSON(401, gin.H{"error": "token revoked"})
			return
		}
		claims, err := ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}
		if claims.UserID <= 0 {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}
		uid := uint64(claims.UserID)
		// Keep all legacy keys for backward compatibility across handlers/services/ws.
		c.Set("user_id", uid)
		c.Set("userID", int64(claims.UserID))
		c.Set("userId", uid)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func ExtractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}
	if token := c.Query("token"); token != "" {
		return token
	}
	return ""
}
