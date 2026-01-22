package audit

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sidot/backend/internal/middleware"
	"github.com/sidot/backend/internal/models"
	"github.com/sidot/backend/internal/repository"
)

// AuditService provides audit logging functionality
type AuditService struct {
	repo *repository.AuditLogRepository
}

// NewAuditService creates a new audit service
func NewAuditService(repo *repository.AuditLogRepository) *AuditService {
	return &AuditService{repo: repo}
}

// LogEvent logs an audit event to the database
func (s *AuditService) LogEvent(
	ctx context.Context,
	acao string,
	entidadeTipo string,
	entidadeID string,
	hospitalID *uuid.UUID,
	severity models.Severity,
	detalhes map[string]interface{},
) error {
	// Try to get user information from context
	var usuarioID *uuid.UUID
	actorName := models.SIDOTBotActor

	// Extract HTTP context information
	var ipAddress, userAgent *string

	// Check if context contains Gin context
	if ginCtx, ok := ctx.(*gin.Context); ok {
		claims, exists := middleware.GetUserClaims(ginCtx)
		if exists && claims != nil {
			uid, err := uuid.Parse(claims.UserID)
			if err == nil {
				usuarioID = &uid
			}
			// Use email as actor name, could also use a "Nome" field if available
			actorName = claims.Email
		}

		// Extract IP address
		ip := ginCtx.ClientIP()
		if ip != "" {
			ipAddress = &ip
		}

		// Extract User-Agent
		ua := ginCtx.GetHeader("User-Agent")
		if ua != "" {
			userAgent = &ua
		}
	}

	// Convert detalhes to JSON
	var detalhesJSON json.RawMessage
	if detalhes != nil {
		jsonBytes, err := json.Marshal(detalhes)
		if err == nil {
			detalhesJSON = jsonBytes
		}
	}

	input := &models.CreateAuditLogInput{
		UsuarioID:    usuarioID,
		ActorName:    actorName,
		Acao:         acao,
		EntidadeTipo: entidadeTipo,
		EntidadeID:   entidadeID,
		HospitalID:   hospitalID,
		Severity:     severity,
		Detalhes:     detalhesJSON,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	_, err := s.repo.Create(ctx, input)
	if err != nil {
		log.Printf("Warning: Failed to create audit log: %v", err)
		return err
	}

	return nil
}

// LogEventWithUser logs an audit event with explicit user information
func (s *AuditService) LogEventWithUser(
	ctx context.Context,
	usuarioID *uuid.UUID,
	actorName string,
	acao string,
	entidadeTipo string,
	entidadeID string,
	hospitalID *uuid.UUID,
	severity models.Severity,
	detalhes map[string]interface{},
	ipAddress *string,
	userAgent *string,
) error {
	// Convert detalhes to JSON
	var detalhesJSON json.RawMessage
	if detalhes != nil {
		jsonBytes, err := json.Marshal(detalhes)
		if err == nil {
			detalhesJSON = jsonBytes
		}
	}

	input := &models.CreateAuditLogInput{
		UsuarioID:    usuarioID,
		ActorName:    actorName,
		Acao:         acao,
		EntidadeTipo: entidadeTipo,
		EntidadeID:   entidadeID,
		HospitalID:   hospitalID,
		Severity:     severity,
		Detalhes:     detalhesJSON,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	_, err := s.repo.Create(ctx, input)
	if err != nil {
		log.Printf("Warning: Failed to create audit log: %v", err)
		return err
	}

	return nil
}

// LogSystemEvent logs an event performed by the system (SIDOT Bot)
func (s *AuditService) LogSystemEvent(
	ctx context.Context,
	acao string,
	entidadeTipo string,
	entidadeID string,
	hospitalID *uuid.UUID,
	severity models.Severity,
	detalhes map[string]interface{},
) error {
	return s.LogEventWithUser(
		ctx,
		nil, // no user ID for system events
		models.SIDOTBotActor,
		acao,
		entidadeTipo,
		entidadeID,
		hospitalID,
		severity,
		detalhes,
		nil,
		nil,
	)
}

// LogAuthEvent logs authentication-related events
func (s *AuditService) LogAuthEvent(
	ctx context.Context,
	acao string,
	usuarioID *uuid.UUID,
	actorName string,
	severity models.Severity,
	ipAddress *string,
	userAgent *string,
	detalhes map[string]interface{},
) error {
	entidadeID := "system"
	if usuarioID != nil {
		entidadeID = usuarioID.String()
	}

	return s.LogEventWithUser(
		ctx,
		usuarioID,
		actorName,
		acao,
		"Auth",
		entidadeID,
		nil, // no hospital_id for auth events
		severity,
		detalhes,
		ipAddress,
		userAgent,
	)
}

// ExtractRequestInfo extracts IP address and User-Agent from Gin context
func ExtractRequestInfo(c *gin.Context) (ipAddress *string, userAgent *string) {
	ip := c.ClientIP()
	if ip != "" {
		ipAddress = &ip
	}

	ua := c.GetHeader("User-Agent")
	if ua != "" {
		userAgent = &ua
	}

	return
}

// GetUserInfoFromContext extracts user ID and name from Gin context
func GetUserInfoFromContext(c *gin.Context) (*uuid.UUID, string) {
	claims, exists := middleware.GetUserClaims(c)
	if !exists || claims == nil {
		return nil, models.SIDOTBotActor
	}

	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, claims.Email
	}

	return &uid, claims.Email
}
