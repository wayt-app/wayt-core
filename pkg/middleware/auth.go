package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/wayt/wayt-core/pkg/response"
)

// TokenVersionFinder looks up the current token version for a user.
// Each user type's repository implements this interface.
type TokenVersionFinder interface {
	FindTokenVersion(id uint) (int, error)
}

// claimVersion extracts token_version from claims; returns 0 if absent (backwards compat).
func claimVersion(claims jwt.MapClaims) int {
	if v, ok := claims["token_version"].(float64); ok {
		return int(v)
	}
	return 0
}

func parseBearer(c *gin.Context, secret []byte) (jwt.MapClaims, bool) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, false
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	return claims, ok
}

// JWTAuth validates Bearer token and injects claims into context.
func JWTAuth(jwtSecret string, finder TokenVersionFinder) gin.HandlerFunc {
	secret := []byte(jwtSecret)
	return func(c *gin.Context) {
		claims, ok := parseBearer(c, secret)
		if !ok {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		rawID, hasID := claims["sub"].(float64)
		if hasID && finder != nil {
			current, err := finder.FindTokenVersion(uint(rawID))
			if err != nil || current != claimVersion(claims) {
				response.Unauthorized(c)
				c.Abort()
				return
			}
		}
		c.Set("user_id", claims["sub"])
		c.Set("username", claims["username"])
		c.Set("role", claims["role"])
		c.Set("restaurant_id", claims["restaurant_id"])
		c.Next()
	}
}

// CustomerAuth validates JWT for customers (type=customer claim).
func CustomerAuth(jwtSecret string, finder TokenVersionFinder) gin.HandlerFunc {
	secret := []byte(jwtSecret)
	return func(c *gin.Context) {
		claims, ok := parseBearer(c, secret)
		if !ok || claims["type"] != "customer" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		if rawID, hasID := claims["sub"].(float64); hasID && finder != nil {
			current, err := finder.FindTokenVersion(uint(rawID))
			if err != nil || current != claimVersion(claims) {
				response.Unauthorized(c)
				c.Abort()
				return
			}
		}
		c.Set("customer_id", claims["sub"])
		c.Set("customer_name", claims["name"])
		c.Next()
	}
}

// SuperAdminOnly allows only superadmin role. Must be placed after JWTAuth.
func SuperAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "superadmin" {
			response.Forbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// OwnerAuth validates JWT for business owners (type=owner claim).
func OwnerAuth(jwtSecret string, finder TokenVersionFinder) gin.HandlerFunc {
	secret := []byte(jwtSecret)
	return func(c *gin.Context) {
		claims, ok := parseBearer(c, secret)
		if !ok || claims["type"] != "owner" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		if rawID, hasID := claims["sub"].(float64); hasID && finder != nil {
			current, err := finder.FindTokenVersion(uint(rawID))
			if err != nil || current != claimVersion(claims) {
				response.Unauthorized(c)
				c.Abort()
				return
			}
		}
		c.Set("owner_id", claims["sub"])
		c.Set("owner_name", claims["name"])
		c.Set("owner_email", claims["email"])
		c.Set("restaurant_id", claims["restaurant_id"])
		c.Next()
	}
}

// StaffAuth validates JWT for staff (type=staff claim).
func StaffAuth(jwtSecret string, finder TokenVersionFinder) gin.HandlerFunc {
	secret := []byte(jwtSecret)
	return func(c *gin.Context) {
		claims, ok := parseBearer(c, secret)
		if !ok || claims["type"] != "staff" {
			response.Unauthorized(c)
			c.Abort()
			return
		}
		if rawID, hasID := claims["sub"].(float64); hasID && finder != nil {
			current, err := finder.FindTokenVersion(uint(rawID))
			if err != nil || current != claimVersion(claims) {
				response.Unauthorized(c)
				c.Abort()
				return
			}
		}
		c.Set("staff_id", claims["sub"])
		c.Set("staff_name", claims["name"])
		c.Set("staff_email", claims["email"])
		c.Set("branch_id", claims["branch_id"])
		c.Set("restaurant_id", claims["restaurant_id"])
		c.Next()
	}
}
