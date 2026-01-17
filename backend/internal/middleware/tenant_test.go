package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantContextMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should set tenant context from user claims", func(t *testing.T) {
		tenantID := uuid.New().String()

		router := gin.New()
		router.Use(func(c *gin.Context) {
			// Simulate AuthRequired middleware setting user claims
			userClaims := &UserClaims{
				UserID:       uuid.New().String(),
				Email:        "test@example.com",
				Role:         "admin",
				TenantID:     tenantID,
				IsSuperAdmin: false,
			}
			c.Set("user_claims", userClaims)
			c.Next()
		})
		router.Use(TenantContextMiddleware())
		router.GET("/test", func(c *gin.Context) {
			tenantCtx, ok := GetTenantContext(c)
			assert.True(t, ok)
			assert.Equal(t, tenantID, tenantCtx.TenantID)
			assert.Equal(t, tenantID, tenantCtx.EffectiveTenantID)
			assert.False(t, tenantCtx.IsSuperAdmin)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should allow super admin to switch tenant context via header", func(t *testing.T) {
		userTenantID := uuid.New().String()
		targetTenantID := uuid.New().String()

		router := gin.New()
		router.Use(func(c *gin.Context) {
			userClaims := &UserClaims{
				UserID:       uuid.New().String(),
				Email:        "super@example.com",
				Role:         "admin",
				TenantID:     userTenantID,
				IsSuperAdmin: true,
			}
			c.Set("user_claims", userClaims)
			c.Next()
		})
		router.Use(TenantContextMiddleware())
		router.GET("/test", func(c *gin.Context) {
			tenantCtx, ok := GetTenantContext(c)
			assert.True(t, ok)
			assert.Equal(t, userTenantID, tenantCtx.TenantID)
			assert.Equal(t, targetTenantID, tenantCtx.EffectiveTenantID)
			assert.True(t, tenantCtx.IsSuperAdmin)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Tenant-Context", targetTenantID)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should deny non-super admin from switching tenant context", func(t *testing.T) {
		userTenantID := uuid.New().String()
		targetTenantID := uuid.New().String()

		router := gin.New()
		router.Use(func(c *gin.Context) {
			userClaims := &UserClaims{
				UserID:       uuid.New().String(),
				Email:        "user@example.com",
				Role:         "admin",
				TenantID:     userTenantID,
				IsSuperAdmin: false,
			}
			c.Set("user_claims", userClaims)
			c.Next()
		})
		router.Use(TenantContextMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Tenant-Context", targetTenantID)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("should reject invalid tenant ID in header", func(t *testing.T) {
		userTenantID := uuid.New().String()

		router := gin.New()
		router.Use(func(c *gin.Context) {
			userClaims := &UserClaims{
				UserID:       uuid.New().String(),
				Email:        "super@example.com",
				Role:         "admin",
				TenantID:     userTenantID,
				IsSuperAdmin: true,
			}
			c.Set("user_claims", userClaims)
			c.Next()
		})
		router.Use(TenantContextMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Tenant-Context", "invalid-uuid")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 401 when no user claims", func(t *testing.T) {
		router := gin.New()
		router.Use(TenantContextMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRequireTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should pass when tenant context is present", func(t *testing.T) {
		tenantID := uuid.New().String()

		router := gin.New()
		router.Use(func(c *gin.Context) {
			tenantCtx := &TenantContext{
				TenantID:          tenantID,
				EffectiveTenantID: tenantID,
				IsSuperAdmin:      false,
			}
			c.Set(tenantContextKey, tenantCtx)
			c.Next()
		})
		router.Use(RequireTenant())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should fail when no tenant context", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireTenant())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("should fail when tenant ID is empty", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			tenantCtx := &TenantContext{
				TenantID:          "",
				EffectiveTenantID: "",
				IsSuperAdmin:      false,
			}
			c.Set(tenantContextKey, tenantCtx)
			c.Next()
		})
		router.Use(RequireTenant())
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestWithTenantContext(t *testing.T) {
	t.Run("should add tenant info to context", func(t *testing.T) {
		tenantID := uuid.New().String()
		ctx := context.Background()

		ctx = WithTenantContext(ctx, tenantID, false)

		retrievedID, isSuperAdmin, err := GetTenantFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, tenantID, retrievedID)
		assert.False(t, isSuperAdmin)
	})

	t.Run("should mark super admin in context", func(t *testing.T) {
		tenantID := uuid.New().String()
		ctx := context.Background()

		ctx = WithTenantContext(ctx, tenantID, true)

		retrievedID, isSuperAdmin, err := GetTenantFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, tenantID, retrievedID)
		assert.True(t, isSuperAdmin)
	})

	t.Run("should return error when no tenant context", func(t *testing.T) {
		ctx := context.Background()

		_, _, err := GetTenantFromContext(ctx)
		assert.Error(t, err)
		assert.Equal(t, ErrMissingTenantContext, err)
	})
}

func TestInjectTenantContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should inject tenant into request context", func(t *testing.T) {
		tenantID := uuid.New().String()

		router := gin.New()
		router.Use(func(c *gin.Context) {
			tenantCtx := &TenantContext{
				TenantID:          tenantID,
				EffectiveTenantID: tenantID,
				IsSuperAdmin:      false,
			}
			c.Set(tenantContextKey, tenantCtx)
			c.Next()
		})
		router.Use(InjectTenantContext())
		router.GET("/test", func(c *gin.Context) {
			// Get tenant from request context (standard Go context)
			retrievedID, isSuperAdmin, err := GetTenantFromContext(c.Request.Context())
			assert.NoError(t, err)
			assert.Equal(t, tenantID, retrievedID)
			assert.False(t, isSuperAdmin)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
