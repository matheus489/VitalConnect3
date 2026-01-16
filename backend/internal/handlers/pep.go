package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// PEPEventInput represents the event received from PEP agents
type PEPEventInput struct {
	HospitalIDOrigem  string `json:"hospital_id_origem"`          // ID from source PEP system
	HospitalID        string `json:"hospital_id" binding:"required"` // VitalConnect hospital UUID
	NomePaciente      string `json:"nome_paciente" binding:"required"`
	DataObito         string `json:"data_obito" binding:"required"`
	CausaMortis       string `json:"causa_mortis" binding:"required"`
	DataNascimento    string `json:"data_nascimento,omitempty"`
	Idade             int    `json:"idade"`
	CNS               string `json:"cns,omitempty"`               // Cartão Nacional de Saúde
	CPFMasked         string `json:"cpf_masked,omitempty"`        // Already masked CPF
	Setor             string `json:"setor,omitempty"`
	Leito             string `json:"leito,omitempty"`
	Prontuario        string `json:"prontuario,omitempty"`
	TimestampDeteccao string `json:"timestamp_deteccao,omitempty"`
}

var (
	pepRedisClient *redis.Client
	pepAPIKeys     map[string]uuid.UUID // API Key -> Hospital UUID mapping
)

// SetPEPRedisClient sets the Redis client for PEP handlers
func SetPEPRedisClient(client *redis.Client) {
	pepRedisClient = client
}

// SetPEPAPIKeys sets the API keys for PEP authentication
// This should be loaded from hospital configurations
func SetPEPAPIKeys(keys map[string]uuid.UUID) {
	pepAPIKeys = keys
}

// ValidatePEPAPIKey validates the API key from the request
func ValidatePEPAPIKey(c *gin.Context) (*uuid.UUID, bool) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		return nil, false
	}

	// Check in-memory cache first
	if hospitalID, ok := pepAPIKeys[apiKey]; ok {
		return &hospitalID, true
	}

	// TODO: Check database for API key if not in cache
	// This allows dynamic API key management

	return nil, false
}

// ReceivePEPEvent handles incoming events from PEP agents
// POST /api/v1/pep/eventos
func ReceivePEPEvent(c *gin.Context) {
	// Validate API Key
	hospitalID, valid := ValidatePEPAPIKey(c)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid or missing API key",
			"code":  "INVALID_API_KEY",
		})
		return
	}

	var input PEPEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate hospital ID matches API key
	inputHospitalID, err := uuid.Parse(input.HospitalID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid hospital_id format",
		})
		return
	}

	if *hospitalID != inputHospitalID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "hospital_id does not match API key",
			"code":  "HOSPITAL_MISMATCH",
		})
		return
	}

	// Generate event ID
	eventID := uuid.New().String()

	// Set detection timestamp if not provided
	timestamp := input.TimestampDeteccao
	if timestamp == "" {
		timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	// Create event for Redis Stream (compatible with ObitoEvent)
	event := map[string]interface{}{
		"obito_id":                 eventID,
		"hospital_id":              input.HospitalID,
		"timestamp_deteccao":       timestamp,
		"nome_paciente":            input.NomePaciente,
		"data_obito":               input.DataObito,
		"causa_mortis":             input.CausaMortis,
		"setor":                    input.Setor,
		"leito":                    input.Leito,
		"idade":                    input.Idade,
		"identificacao_desconhecida": false,
		"source":                   "pep",           // Mark as coming from PEP agent
		"hospital_id_origem":       input.HospitalIDOrigem,
		"cns":                      input.CNS,
		"cpf_masked":               input.CPFMasked,
		"prontuario":               input.Prontuario,
	}

	// Add data_nascimento if provided
	if input.DataNascimento != "" {
		event["data_nascimento"] = input.DataNascimento
	}

	// Publish to Redis Stream
	if pepRedisClient != nil {
		ctx := context.Background()
		eventJSON, err := json.Marshal(event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to serialize event",
			})
			return
		}

		_, err = pepRedisClient.XAdd(ctx, &redis.XAddArgs{
			Stream: "obitos:detectados",
			Values: map[string]interface{}{
				"event": string(eventJSON),
			},
		}).Result()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to publish event to stream",
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Event received successfully",
		"event_id": eventID,
	})
}

// GetPEPStatus returns the status of PEP integration
// GET /api/v1/pep/status
func GetPEPStatus(c *gin.Context) {
	configured := pepRedisClient != nil && len(pepAPIKeys) > 0

	c.JSON(http.StatusOK, gin.H{
		"configured":     configured,
		"hospitals_count": len(pepAPIKeys),
		"message": func() string {
			if configured {
				return "PEP integration is enabled"
			}
			return "PEP integration requires API key configuration"
		}(),
	})
}
