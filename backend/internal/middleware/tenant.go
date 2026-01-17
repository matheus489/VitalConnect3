package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TenantContext holds tenant information for the current request
type TenantContext struct {
	TenantID          string // Tenant ID from JWT claims (user's assigned tenant)
	IsSuperAdmin      bool   // Whether the user is a super admin
	EffectiveTenantID string // The tenant ID to use for queries (may differ for super-admin context switch)
}

// Context keys for tenant context
const (
	tenantContextKey       = "tenant_context"
	tenantContextHeaderKey = "X-Tenant-Context"
)

// Errors related to tenant context
var (
	ErrMissingTenantContext  = errors.New("tenant context not found")
	ErrInvalidTenantID       = errors.New("invalid tenant ID format")
	ErrTenantContextDenied   = errors.New("tenant context switch not allowed")
	ErrMissingTenantID       = errors.New("tenant ID is required")
)

// TenantContextMiddleware extracts tenant information from JWT claims and optional header
// Must be used AFTER AuthRequired middleware
func TenantContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context (set by AuthRequired middleware)
		userClaims, ok := GetUserClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required for tenant context",
			})
			return
		}

		// Create tenant context from user claims
		tenantCtx := &TenantContext{
			TenantID:          userClaims.TenantID,
			IsSuperAdmin:      userClaims.IsSuperAdmin,
			EffectiveTenantID: userClaims.TenantID, // Default to user's tenant
		}

		// Check for X-Tenant-Context header (super-admin only)
		headerTenantID := c.GetHeader(tenantContextHeaderKey)
		if headerTenantID != "" {
			// Only super-admins can use X-Tenant-Context header
			if !userClaims.IsSuperAdmin {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "tenant context switch requires super admin privileges",
					"code":  "TENANT_CONTEXT_DENIED",
				})
				return
			}

			// Validate the header tenant ID format
			if _, err := uuid.Parse(headerTenantID); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error":   "invalid X-Tenant-Context header",
					"details": "tenant ID must be a valid UUID",
				})
				return
			}

			// Set effective tenant ID to the one from header
			tenantCtx.EffectiveTenantID = headerTenantID
		}

		// Store tenant context in Gin context
		c.Set(tenantContextKey, tenantCtx)

		c.Next()
	}
}

// RequireTenant is a middleware that ensures tenant context is present and valid
// Must be used AFTER TenantContextMiddleware
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, ok := GetTenantContext(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "tenant context required",
				"code":  "TENANT_REQUIRED",
			})
			return
		}

		// Ensure effective tenant ID is present
		if tenantCtx.EffectiveTenantID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "tenant ID is required for this operation",
				"code":  "TENANT_ID_MISSING",
			})
			return
		}

		// Validate tenant ID format
		if _, err := uuid.Parse(tenantCtx.EffectiveTenantID); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid tenant ID format",
				"code":  "INVALID_TENANT_ID",
			})
			return
		}

		c.Next()
	}
}

// GetTenantContext retrieves the tenant context from Gin context
func GetTenantContext(c *gin.Context) (*TenantContext, bool) {
	ctx, exists := c.Get(tenantContextKey)
	if !exists {
		return nil, false
	}

	tenantCtx, ok := ctx.(*TenantContext)
	return tenantCtx, ok
}

// GetTenantIDFromGinContext extracts the effective tenant ID from Gin context
func GetTenantIDFromGinContext(c *gin.Context) (string, error) {
	tenantCtx, ok := GetTenantContext(c)
	if !ok {
		return "", ErrMissingTenantContext
	}

	if tenantCtx.EffectiveTenantID == "" {
		return "", ErrMissingTenantID
	}

	return tenantCtx.EffectiveTenantID, nil
}

// IsSuperAdminFromGinContext checks if the current user is a super admin
func IsSuperAdminFromGinContext(c *gin.Context) bool {
	tenantCtx, ok := GetTenantContext(c)
	if !ok {
		return false
	}
	return tenantCtx.IsSuperAdmin
}

// TenantContextKey is the context key type for standard context.Context
type tenantCtxKeyType string

const tenantCtxKey tenantCtxKeyType = "tenant"

// TenantInfo holds tenant information for standard Go context
type TenantInfo struct {
	TenantID     string
	IsSuperAdmin bool
}

// WithTenantContext creates a new context with tenant information
// This is used to pass tenant context to repository layer
func WithTenantContext(ctx context.Context, tenantID string, isSuperAdmin bool) context.Context {
	return context.WithValue(ctx, tenantCtxKey, &TenantInfo{
		TenantID:     tenantID,
		IsSuperAdmin: isSuperAdmin,
	})
}

// GetTenantFromContext extracts tenant information from standard Go context
// Returns tenant ID, isSuperAdmin flag, and error if not found
func GetTenantFromContext(ctx context.Context) (string, bool, error) {
	info, ok := ctx.Value(tenantCtxKey).(*TenantInfo)
	if !ok || info == nil {
		return "", false, ErrMissingTenantContext
	}
	return info.TenantID, info.IsSuperAdmin, nil
}

// GetTenantIDFromContext extracts only the tenant ID from standard Go context
func GetTenantIDFromContext(ctx context.Context) (string, error) {
	tenantID, _, err := GetTenantFromContext(ctx)
	return tenantID, err
}

// InjectTenantContext is a middleware that injects tenant context into the request's context.Context
// This makes tenant info available to repository layer via context
// Must be used AFTER TenantContextMiddleware
func InjectTenantContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, ok := GetTenantContext(c)
		if ok && tenantCtx.EffectiveTenantID != "" {
			// Create new context with tenant info and replace the request context
			ctx := WithTenantContext(c.Request.Context(), tenantCtx.EffectiveTenantID, tenantCtx.IsSuperAdmin)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}
