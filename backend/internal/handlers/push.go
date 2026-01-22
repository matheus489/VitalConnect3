package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sidot/backend/internal/middleware"
	"github.com/sidot/backend/internal/repository"
	"github.com/sidot/backend/internal/services/notification"
)

var (
	pushService *notification.PushService
	pushSubRepo *repository.PushSubscriptionRepository
)

// SetPushService sets the push service for handlers
func SetPushService(service *notification.PushService) {
	pushService = service
}

// SetPushSubscriptionRepository sets the push subscription repository
func SetPushSubscriptionRepository(repo *repository.PushSubscriptionRepository) {
	pushSubRepo = repo
}

// SubscribePushInput represents the input for subscribing to push notifications
type SubscribePushInput struct {
	Token     string `json:"token" binding:"required"`
	Platform  string `json:"platform"` // web, android, ios
	UserAgent string `json:"user_agent"`
}

// SubscribePush registers a push notification subscription
// POST /api/v1/push/subscribe
func SubscribePush(c *gin.Context) {
	if pushSubRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "push service not configured"})
		return
	}

	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var input SubscribePushInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	platform := input.Platform
	if platform == "" {
		platform = "web"
	}

	userAgent := input.UserAgent
	if userAgent == "" {
		userAgent = c.GetHeader("User-Agent")
	}

	sub, err := pushSubRepo.Create(c.Request.Context(), userID, input.Token, platform, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register subscription"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "Push subscription registered successfully",
		"subscription_id": sub.ID,
	})
}

// UnsubscribePush removes a push notification subscription
// DELETE /api/v1/push/unsubscribe
func UnsubscribePush(c *gin.Context) {
	if pushSubRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "push service not configured"})
		return
	}

	var input struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := pushSubRepo.Delete(c.Request.Context(), input.Token)
	if err != nil {
		if err == repository.ErrSubscriptionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Push subscription removed successfully"})
}

// GetMySubscriptions returns push subscriptions for the current user
// GET /api/v1/push/subscriptions
func GetMySubscriptions(c *gin.Context) {
	if pushSubRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "push service not configured"})
		return
	}

	claims, ok := middleware.GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	subs, err := pushSubRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscriptions"})
		return
	}

	// Don't expose full tokens for security
	responses := make([]gin.H, len(subs))
	for i, sub := range subs {
		tokenPreview := sub.Token
		if len(tokenPreview) > 20 {
			tokenPreview = tokenPreview[:20] + "..."
		}
		responses[i] = gin.H{
			"id":         sub.ID,
			"platform":   sub.Platform,
			"token":      tokenPreview,
			"created_at": sub.CreatedAt,
			"updated_at": sub.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": responses,
		"total":         len(responses),
	})
}

// GetPushStatus returns the push notification service status
// GET /api/v1/push/status
func GetPushStatus(c *gin.Context) {
	configured := pushService != nil && pushService.IsConfigured()

	c.JSON(http.StatusOK, gin.H{
		"configured": configured,
		"message": func() string {
			if configured {
				return "Push notifications are enabled"
			}
			return "Push notifications require FCM_SERVER_KEY configuration"
		}(),
	})
}
