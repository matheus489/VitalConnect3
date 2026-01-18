package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRequireSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should allow access when user is super admin", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			userClaims := &UserClaims{
				UserID:       uuid.New().String(),
				Email:        "superadmin@example.com",
				Role:         "admin",
				TenantID:     uuid.New().String(),
				IsSuperAdmin: true,
			}
			c.Set("user_claims", userClaims)
			c.Next()
		})
		router.Use(RequireSuperAdmin())
		router.GET("/admin/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["message"])
	})

	t.Run("should block access when user is not super admin", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			userClaims := &UserClaims{
				UserID:       uuid.New().String(),
				Email:        "regularadmin@example.com",
				Role:         "admin",
				TenantID:     uuid.New().String(),
				IsSuperAdmin: false,
			}
			c.Set("user_claims", userClaims)
			c.Next()
		})
		router.Use(RequireSuperAdmin())
		router.GET("/admin/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "super admin access required", response["error"])
		assert.Equal(t, "SUPER_ADMIN_REQUIRED", response["code"])
	})

	t.Run("should block access when user claims are missing", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireSuperAdmin())
		router.GET("/admin/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "authentication required", response["error"])
		assert.Equal(t, "AUTH_REQUIRED", response["code"])
	})

	t.Run("should block access when claims have invalid format", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			// Set invalid claims type
			c.Set("user_claims", "invalid_claims")
			c.Next()
		})
		router.Use(RequireSuperAdmin())
		router.GET("/admin/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid claims format", response["error"])
		assert.Equal(t, "INVALID_CLAIMS", response["code"])
	})

	t.Run("should protect admin routes with both auth and super admin middleware", func(t *testing.T) {
		// Test that routes are properly protected when both middlewares are applied
		router := gin.New()

		// Apply middlewares in order: AuthRequired would set claims, then RequireSuperAdmin checks them
		// Here we simulate AuthRequired setting claims for a regular user
		router.Use(func(c *gin.Context) {
			userClaims := &UserClaims{
				UserID:       uuid.New().String(),
				Email:        "user@example.com",
				Role:         "operador",
				TenantID:     uuid.New().String(),
				IsSuperAdmin: false,
			}
			c.Set("user_claims", userClaims)
			c.Next()
		})

		adminGroup := router.Group("/api/v1/admin")
		adminGroup.Use(RequireSuperAdmin())
		{
			adminGroup.GET("/tenants", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
			adminGroup.GET("/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
		}

		// Test /admin/tenants is protected
		req1 := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusForbidden, w1.Code)

		// Test /admin/users is protected
		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusForbidden, w2.Code)
	})
}
