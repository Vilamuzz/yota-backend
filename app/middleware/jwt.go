package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Vilamuzz/yota-backend/config"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
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

func (m *JWTMiddleware) AuthOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqToken := c.GetHeader("Authorization")
		if reqToken == "" {
			c.Next()
			return
		}

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

func (m *JWTMiddleware) RequireRoles(allowedRoles ...enum.RoleName) gin.HandlerFunc {
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

		hasRole := false
		for _, role := range allowedRoles {
			if claims.ActiveRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, pkg.NewResponse(
				http.StatusForbidden,
				"Akses ditolak: izin tidak memadai",
				nil,
				nil,
			))
			return
		}

		c.Set("user_data", *claims)
		c.Next()
	}
}

func (m *JWTMiddleware) extractAndValidateToken(c *gin.Context) (*jwt_pkg.UserJWTClaims, error) {
	reqToken := c.GetHeader("Authorization")
	if reqToken == "" {
		return nil, errors.New("Tidak terautorisasi: Header Otorisasi tidak ditemukan")
	}

	if !strings.HasPrefix(reqToken, "Bearer ") {
		reqToken = "Bearer " + reqToken
	}

	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		return nil, errors.New("Tidak terautorisasi: Format token tidak valid")
	}

	tokenString := strings.TrimSpace(splitToken[1])
	claims := &jwt_pkg.UserJWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Tidak terautorisasi: Metode penandatanganan tidak valid")
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("Tidak terautorisasi: Token kedaluwarsa")
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, errors.New("Tidak terautorisasi: Tanda tangan token tidak valid")
		}
		return nil, errors.New("Tidak terautorisasi: " + err.Error())
	}

	if !token.Valid {
		return nil, errors.New("Tidak terautorisasi: Token tidak valid")
	}

	return claims, nil
}
