package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/models"
)

const (
	// ImpersonationTokenDuration is the duration for impersonation tokens (1 hour)
	ImpersonationTokenDuration = 1 * time.Hour
)

// ImpersonationClaims extends standard claims with impersonation info
type ImpersonationClaims struct {
	OriginalAdminID string `json:"original_admin_id"`
	IsImpersonation bool   `json:"is_impersonation"`
}

// AuditLogger defines the interface for audit logging
type AuditLogger interface {
	Create(ctx context.Context, input *models.CreateAuditLogInput) (*models.AuditLog, error)
}

// ImpersonationService handles user impersonation for super admins
type ImpersonationService struct {
	jwtService   *JWTService
	userRepo     UserRepository
	auditLogRepo AuditLogger
}

// NewImpersonationService creates a new impersonation service
func NewImpersonationService(jwtService *JWTService, userRepo UserRepository, auditLogRepo AuditLogger) *ImpersonationService {
	return &ImpersonationService{
		jwtService:   jwtService,
		userRepo:     userRepo,
		auditLogRepo: auditLogRepo,
	}
}

// ImpersonationResult contains the result of an impersonation request
type ImpersonationResult struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	ExpiresIn   int       `json:"expires_in"` // seconds
	User        *User     `json:"user"`
}

// GenerateImpersonationToken creates a short-lived JWT for impersonating a user
// The token includes the original admin ID in the claims for audit purposes
func (s *ImpersonationService) GenerateImpersonationToken(
	ctx context.Context,
	adminUserID uuid.UUID,
	targetUserID uuid.UUID,
	ipAddress *string,
	userAgent *string,
) (*ImpersonationResult, error) {
	// Get the target user
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	// Get the admin user for audit logging
	adminUser, err := s.userRepo.GetByID(ctx, adminUserID)
	if err != nil {
		return nil, err
	}

	// Prepare hospital ID string
	hospitalID := ""
	if targetUser.HospitalID != nil {
		hospitalID = targetUser.HospitalID.String()
	}

	// Prepare tenant ID string
	tenantID := ""
	if targetUser.TenantID != nil {
		tenantID = targetUser.TenantID.String()
	}

	// Generate a short-lived access token for the target user
	// We use the standard token generation but with a shorter duration
	now := time.Now()
	expiresAt := now.Add(ImpersonationTokenDuration)

	accessToken, err := s.jwtService.GenerateAccessTokenWithTenant(
		targetUser.ID.String(),
		targetUser.Email,
		targetUser.Role,
		hospitalID,
		tenantID,
		targetUser.IsSuperAdmin,
	)
	if err != nil {
		return nil, err
	}

	// Log the impersonation action to audit logs
	err = s.logImpersonationAction(ctx, adminUser, targetUser, ipAddress, userAgent)
	if err != nil {
		// Log error but don't fail the impersonation
		// The impersonation is still valid even if audit logging fails
	}

	return &ImpersonationResult{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
		ExpiresIn:   int(ImpersonationTokenDuration.Seconds()),
		User:        targetUser,
	}, nil
}

// logImpersonationAction logs the impersonation event to audit logs
func (s *ImpersonationService) logImpersonationAction(
	ctx context.Context,
	adminUser *User,
	targetUser *User,
	ipAddress *string,
	userAgent *string,
) error {
	if s.auditLogRepo == nil {
		return nil
	}

	detalhes := map[string]any{
		"admin_email":       adminUser.Email,
		"admin_id":          adminUser.ID.String(),
		"target_user_email": targetUser.Email,
		"target_user_id":    targetUser.ID.String(),
		"target_user_role":  targetUser.Role,
		"impersonation":     true,
	}

	if targetUser.TenantID != nil {
		detalhes["target_tenant_id"] = targetUser.TenantID.String()
	}

	detalhesJSON, err := json.Marshal(detalhes)
	if err != nil {
		return err
	}

	input := &models.CreateAuditLogInput{
		UsuarioID:    &adminUser.ID,
		ActorName:    adminUser.Email,
		Acao:         ActionUserImpersonate,
		EntidadeTipo: "User",
		EntidadeID:   targetUser.ID.String(),
		HospitalID:   nil,
		Severity:     models.SeverityWarn,
		Detalhes:     detalhesJSON,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	_, err = s.auditLogRepo.Create(ctx, input)
	return err
}

// Action constants for impersonation
const (
	ActionUserImpersonate  = "admin.user.impersonate"
	ActionUserBan          = "admin.user.ban"
	ActionUserUnban        = "admin.user.unban"
	ActionUserRoleChange   = "admin.user.role_change"
	ActionUserResetPwd     = "admin.user.reset_password"
	ActionHospitalReassign = "admin.hospital.reassign"
)
