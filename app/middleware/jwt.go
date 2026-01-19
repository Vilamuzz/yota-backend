package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Vilamuzz/yota-backend/config"
	"github.com/Vilamuzz/yota-backend/pkg"
	jwt_pkg "github.com/Vilamuzz/yota-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTMiddleware struct {
	secretKey string
}

func NewJWTMiddleware() *JWTMiddleware {
	return &JWTMiddleware{
		secretKey: config.GetJWTSecretKey(),
	}
}

// AuthRequired validates JWT token and sets user claims in context
func (m *JWTMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.extractAndValidateToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
				http.StatusUnauthorized,
				err.Error(),
				nil,
				nil,
			))
			return
		}
		c.Set("user_data", *claims)
		c.Next()
	}
}

// RequireRoles creates middleware that checks if user has any of the specified roles
func (m *JWTMiddleware) RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := m.extractAndValidateToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
				http.StatusUnauthorized,
				err.Error(),
				nil,
				nil,
			))
			return
		}

		// Check if user role is in allowed roles
		hasRole := false
		for _, role := range allowedRoles {
			if claims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, pkg.NewResponse(
				http.StatusForbidden,
				"Access denied: insufficient permissions",
				nil,
				nil,
			))
			return
		}

		c.Set("user_data", *claims)
		c.Next()
	}
}

// extractAndValidateToken extracts and validates the JWT token from Authorization header
func (m *JWTMiddleware) extractAndValidateToken(c *gin.Context) (*jwt_pkg.UserJWTClaims, error) {
	reqToken := c.GetHeader("Authorization")
	if reqToken == "" {
		return nil, errors.New("Unauthorized: Missing Authorization Header")
	}

	if !strings.HasPrefix(reqToken, "Bearer ") {
		reqToken = "Bearer " + reqToken
	}

	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		return nil, errors.New("Unauthorized: Invalid Token Format")
	}

	tokenString := strings.TrimSpace(splitToken[1])
	claims := &jwt_pkg.UserJWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unauthorized: Invalid signing method")
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("Unauthorized: Token Expired")
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, errors.New("Unauthorized: Invalid Token Signature")
		}
		return nil, errors.New("Unauthorized: " + err.Error())
	}

	if !token.Valid {
		return nil, errors.New("Unauthorized: Invalid Token")
	}

	return claims, nil
}

// Role constants
