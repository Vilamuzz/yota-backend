package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Vilamuzz/yota-backend/pkg"
	jwt_pkg "github.com/Vilamuzz/yota-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (m *appMiddleware) AuthSuperadmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqToken := c.GetHeader("Authorization")
		if reqToken == "" {
			c.JSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "Unauthorized: Missing Authorization Header", nil, nil))
			return
		}
		if !strings.HasPrefix(reqToken, "Bearer") {
			reqToken = "Bearer " + reqToken
		}
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "Unauthorized: Invalid Token Format", nil, nil))
			return
		}
		tokenString := splitToken[1]
		token, err := jwt.ParseWithClaims(tokenString, jwt_pkg.UserJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.secretKeySuperadmin), nil
		})
		if !token.Valid {
			if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
					http.StatusUnauthorized,
					"Unauthorized: Invalid Token Signature",
					nil,
					nil,
				))
				return
			}
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
					http.StatusUnauthorized,
					"Unauthorized: Token Expired",
					nil,
					nil,
				))
				return
			}
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
				http.StatusUnauthorized,
				err.Error(),
				nil,
				nil,
			))
			return
		}
		claims, ok := token.Claims.(*jwt_pkg.UserJWTClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
				http.StatusUnauthorized,
				"Unauthorized: Invalid Token Claims",
				nil,
				nil,
			))
			return
		}
		c.Set("user_data", *claims)
		c.Next()
	}
}

func (m *appMiddleware) AuthAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqToken := c.GetHeader("Authorization")
		if reqToken == "" {
			c.JSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "Unauthorized: Missing Authorization Header", nil, nil))
			return
		}
		if !strings.HasPrefix(reqToken, "Bearer") {
			reqToken = "Bearer " + reqToken
		}
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "Unauthorized: Invalid Token Format", nil, nil))
			return
		}
		tokenString := splitToken[1]
		token, err := jwt.ParseWithClaims(tokenString, jwt_pkg.UserJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.secretKeyAdmin), nil
		})
		if !token.Valid {
			if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
					http.StatusUnauthorized,
					"Unauthorized: Invalid Token Signature",
					nil,
					nil,
				))
				return
			}
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
					http.StatusUnauthorized,
					"Unauthorized: Token Expired",
					nil,
					nil,
				))
				return
			}
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
				http.StatusUnauthorized,
				err.Error(),
				nil,
				nil,
			))
			return
		}
		claims, ok := token.Claims.(*jwt_pkg.UserJWTClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
				http.StatusUnauthorized,
				"Unauthorized: Invalid Token Claims",
				nil,
				nil,
			))
			return
		}
		c.Set("user_data", *claims)
		c.Next()
	}
}

func (m *appMiddleware) AuthUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqToken := c.GetHeader("Authorization")
		if reqToken == "" {
			c.JSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "Unauthorized: Missing Authorization Header", nil, nil))
			return
		}

		if !strings.HasPrefix(reqToken, "Bearer") {
			reqToken = "Bearer " + reqToken
		}

		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "Unauthorized: Invalid Token Format", nil, nil))
			return
		}

		tokenString := splitToken[1]
		token, err := jwt.ParseWithClaims(tokenString, jwt_pkg.UserJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.secretKeyUser), nil
		})

		if !token.Valid {
			if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
					http.StatusUnauthorized,
					"Unauthorized: Invalid Token Signature",
					nil,
					nil,
				))
				return
			}

			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
					http.StatusUnauthorized,
					"Unauthorized: Token Expired",
					nil,
					nil,
				))
				return
			}
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
				http.StatusUnauthorized,
				err.Error(),
				nil,
				nil,
			))
			return
		}
		claims, ok := token.Claims.(*jwt_pkg.UserJWTClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewResponse(
				http.StatusUnauthorized,
				"Unauthorized: Invalid Token Claims",
				nil,
				nil,
			))
			return
		}
		c.Set("user_data", *claims)
		c.Next()
	}
}
