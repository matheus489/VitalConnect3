package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vitalconnect/backend/internal/middleware"
	"github.com/vitalconnect/backend/internal/models"
)

// MockOccurrenceRepository for testing
type MockOccurrenceRepository struct {
	occurrences []models.Occurrence
	totalCount  int
}

func NewMockOccurrenceRepository() *MockOccurrenceRepository {
	return &MockOccurrenceRepository{
		occurrences: []models.Occurrence{},
		totalCount:  0,
	}
}

func (r *MockOccurrenceRepository) List(ctx context.Context, filters models.OccurrenceListFilters) ([]models.Occurrence, int, error) {
	result := []models.Occurrence{}
	for _, o := range r.occurrences {
		// Apply status filter
		if filters.Status != nil && o.Status != *filters.Status {
			continue
		}
		// Apply hospital filter
		if filters.HospitalID != nil && o.HospitalID.String() != *filters.HospitalID {
			continue
		}
		result = append(result, o)
	}
	return result, len(result), nil
}

func (r *MockOccurrenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Occurrence, error) {
	for _, o := range r.occurrences {
		if o.ID == id {
			return &o, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *MockOccurrenceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.OccurrenceStatus) error {
	for i, o := range r.occurrences {
		if o.ID == id {
			r.occurrences[i].Status = status
			return nil
		}
	}
	return sql.ErrNoRows
}

func (r *MockOccurrenceRepository) GetTodayEligibleCount(ctx context.Context) (int, error) {
	return len(r.occurrences), nil
}

func (r *MockOccurrenceRepository) GetAverageNotificationTime(ctx context.Context) (float64, error) {
	return 45.5, nil
}

func (r *MockOccurrenceRepository) GetPendingCount(ctx context.Context) (int, error) {
	count := 0
	for _, o := range r.occurrences {
		if o.Status == models.StatusPendente {
			count++
		}
	}
	return count, nil
}

func (r *MockOccurrenceRepository) GetEmAndamentoCount(ctx context.Context) (int, error) {
	count := 0
	for _, o := range r.occurrences {
		if o.Status == models.StatusEmAndamento {
			count++
		}
	}
	return count, nil
}

func (r *MockOccurrenceRepository) AddOccurrence(o models.Occurrence) {
	r.occurrences = append(r.occurrences, o)
	r.totalCount++
}

// MockOccurrenceHistoryRepository for testing
type MockOccurrenceHistoryRepository struct {
	histories []models.OccurrenceHistory
}

func NewMockOccurrenceHistoryRepository() *MockOccurrenceHistoryRepository {
	return &MockOccurrenceHistoryRepository{
		histories: []models.OccurrenceHistory{},
	}
}

func (r *MockOccurrenceHistoryRepository) Create(ctx context.Context, input *models.CreateHistoryInput) (*models.OccurrenceHistory, error) {
	h := models.OccurrenceHistory{
		ID:             uuid.New(),
		OccurrenceID:   input.OccurrenceID,
		UserID:         input.UserID,
		Acao:           input.Acao,
		StatusAnterior: input.StatusAnterior,
		StatusNovo:     input.StatusNovo,
		Observacoes:    input.Observacoes,
		Desfecho:       input.Desfecho,
		CreatedAt:      time.Now(),
	}
	r.histories = append(r.histories, h)
	return &h, nil
}

func (r *MockOccurrenceHistoryRepository) GetByOccurrenceID(ctx context.Context, occurrenceID uuid.UUID) ([]models.OccurrenceHistory, error) {
	var result []models.OccurrenceHistory
	for _, h := range r.histories {
		if h.OccurrenceID == occurrenceID {
			result = append(result, h)
		}
	}
	return result, nil
}

func (r *MockOccurrenceHistoryRepository) GetOutcomeByOccurrenceID(ctx context.Context, occurrenceID uuid.UUID) (*models.OccurrenceHistory, error) {
	for _, h := range r.histories {
		if h.OccurrenceID == occurrenceID && h.Desfecho != nil {
			return &h, nil
		}
	}
	return nil, nil
}

// setupTestRouter creates a test router with mocked middleware
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// mockAuthMiddleware creates a middleware that sets user claims
func mockAuthMiddleware(userID, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := &middleware.UserClaims{
			UserID: userID,
			Email:  "test@vitalconnect.gov.br",
			Role:   role,
		}
		c.Set("user_claims", claims)
		c.Next()
	}
}

// Helper function to create test occurrences
func createTestOccurrence(status models.OccurrenceStatus, hospitalID uuid.UUID) models.Occurrence {
	dataObito := time.Now().Add(-2 * time.Hour)
	completeData := models.OccurrenceCompleteData{
		NomePaciente:   "Joao Silva",
		DataNascimento: time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC),
		DataObito:      dataObito,
		CausaMortis:    "Causa natural",
		Idade:          75,
		Setor:          "UTI",
	}
	dataJSON, _ := json.Marshal(completeData)

	return models.Occurrence{
		ID:                    uuid.New(),
		ObitoID:               uuid.New(),
		HospitalID:            hospitalID,
		Status:                status,
		ScorePriorizacao:      100,
		NomePacienteMascarado: "Jo** Si***",
		DadosCompletos:        dataJSON,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		DataObito:             dataObito,
		JanelaExpiraEm:        dataObito.Add(6 * time.Hour),
		Hospital: &models.Hospital{
			ID:     hospitalID,
			Nome:   "Hospital Teste",
			Codigo: "HT",
			Ativo:  true,
		},
	}
}

// Test 1: Testar listagem de ocorrencias com paginacao
func TestListOccurrencesWithPagination(t *testing.T) {
	mockRepo := NewMockOccurrenceRepository()
	hospitalID := uuid.New()

	// Add test occurrences
	for i := 0; i < 25; i++ {
		o := createTestOccurrence(models.StatusPendente, hospitalID)
		mockRepo.AddOccurrence(o)
	}

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(uuid.New().String(), "operador"))

	// Create a handler that uses the mock
	router.GET("/occurrences", func(c *gin.Context) {
		filters := models.DefaultFilters()

		// Parse pagination
		if page := c.Query("page"); page != "" {
			filters.Page = 2
		}
		filters.PageSize = 10

		occurrences, total, err := mockRepo.List(c.Request.Context(), filters)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := make([]models.OccurrenceListResponse, 0, len(occurrences))
		for _, o := range occurrences {
			response = append(response, o.ToListResponse())
		}

		c.JSON(http.StatusOK, models.NewPaginatedResponse(response, filters.Page, filters.PageSize, total))
	})

	// Test request
	req, _ := http.NewRequest("GET", "/occurrences?page=2&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.TotalItems != 25 {
		t.Errorf("Expected total_items 25, got %d", response.TotalItems)
	}
}

// Test 2: Testar filtro por status e hospital
func TestListOccurrencesWithFilters(t *testing.T) {
	mockRepo := NewMockOccurrenceRepository()
	hospitalID1 := uuid.New()
	hospitalID2 := uuid.New()

	// Add occurrences with different statuses and hospitals
	o1 := createTestOccurrence(models.StatusPendente, hospitalID1)
	o2 := createTestOccurrence(models.StatusEmAndamento, hospitalID1)
	o3 := createTestOccurrence(models.StatusPendente, hospitalID2)
	mockRepo.AddOccurrence(o1)
	mockRepo.AddOccurrence(o2)
	mockRepo.AddOccurrence(o3)

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(uuid.New().String(), "operador"))

	router.GET("/occurrences", func(c *gin.Context) {
		filters := models.DefaultFilters()

		if status := c.Query("status"); status != "" {
			s := models.OccurrenceStatus(status)
			filters.Status = &s
		}

		if hospitalID := c.Query("hospital_id"); hospitalID != "" {
			filters.HospitalID = &hospitalID
		}

		occurrences, total, _ := mockRepo.List(c.Request.Context(), filters)

		response := make([]models.OccurrenceListResponse, 0, len(occurrences))
		for _, o := range occurrences {
			response = append(response, o.ToListResponse())
		}

		c.JSON(http.StatusOK, models.NewPaginatedResponse(response, filters.Page, filters.PageSize, total))
	})

	// Test filter by status
	req, _ := http.NewRequest("GET", "/occurrences?status=PENDENTE", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should return only PENDENTE occurrences
	if response.TotalItems != 2 {
		t.Errorf("Expected 2 PENDENTE occurrences, got %d", response.TotalItems)
	}

	// Test filter by hospital
	req2, _ := http.NewRequest("GET", fmt.Sprintf("/occurrences?hospital_id=%s", hospitalID1.String()), nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var response2 models.PaginatedResponse
	json.Unmarshal(w2.Body.Bytes(), &response2)

	// Should return only hospital1 occurrences
	if response2.TotalItems != 2 {
		t.Errorf("Expected 2 occurrences for hospital1, got %d", response2.TotalItems)
	}
}

// Test 3: Testar transicao de status valida
func TestValidStatusTransition(t *testing.T) {
	mockOccRepo := NewMockOccurrenceRepository()
	mockHistRepo := NewMockOccurrenceHistoryRepository()
	hospitalID := uuid.New()

	// Create occurrence in PENDENTE status
	occurrence := createTestOccurrence(models.StatusPendente, hospitalID)
	mockOccRepo.AddOccurrence(occurrence)

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(uuid.New().String(), "operador"))

	router.PATCH("/occurrences/:id/status", func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
			return
		}

		var input models.UpdateStatusInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		occ, err := mockOccRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Validate transition
		if !occ.Status.CanTransitionTo(input.Status) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "invalid status transition",
				"current_status": occ.Status,
				"target_status":  input.Status,
			})
			return
		}

		// Update status
		mockOccRepo.UpdateStatus(c.Request.Context(), id, input.Status)

		// Create history entry
		historyInput := &models.CreateHistoryInput{
			OccurrenceID:   id,
			Acao:           models.ActionStatusChanged,
			StatusAnterior: &occ.Status,
			StatusNovo:     &input.Status,
		}
		mockHistRepo.Create(c.Request.Context(), historyInput)

		c.JSON(http.StatusOK, gin.H{
			"message":    "status updated",
			"new_status": input.Status,
		})
	})

	// Test valid transition: PENDENTE -> EM_ANDAMENTO
	body := models.UpdateStatusInput{Status: models.StatusEmAndamento}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/occurrences/%s/status", occurrence.ID.String()), bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify history was created
	histories, _ := mockHistRepo.GetByOccurrenceID(context.Background(), occurrence.ID)
	if len(histories) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(histories))
	}
}

// Test 4: Testar rejeicao de transicao invalida
func TestInvalidStatusTransition(t *testing.T) {
	mockOccRepo := NewMockOccurrenceRepository()
	hospitalID := uuid.New()

	// Create occurrence in PENDENTE status
	occurrence := createTestOccurrence(models.StatusPendente, hospitalID)
	mockOccRepo.AddOccurrence(occurrence)

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(uuid.New().String(), "operador"))

	router.PATCH("/occurrences/:id/status", func(c *gin.Context) {
		idParam := c.Param("id")
		id, _ := uuid.Parse(idParam)

		var input models.UpdateStatusInput
		c.ShouldBindJSON(&input)

		occ, err := mockOccRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Validate transition
		if !occ.Status.CanTransitionTo(input.Status) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "invalid status transition",
				"current_status": occ.Status,
				"target_status":  input.Status,
				"allowed":        models.StatusTransitions[occ.Status],
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Test invalid transition: PENDENTE -> ACEITA (should be rejected)
	body := models.UpdateStatusInput{Status: models.StatusAceita}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/occurrences/%s/status", occurrence.ID.String()), bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] != "invalid status transition" {
		t.Errorf("Expected 'invalid status transition' error, got %v", response["error"])
	}
}

// Test 5: Testar registro de desfecho
func TestRegisterOutcome(t *testing.T) {
	mockOccRepo := NewMockOccurrenceRepository()
	mockHistRepo := NewMockOccurrenceHistoryRepository()
	hospitalID := uuid.New()

	// Create occurrence in ACEITA status (allows outcome registration)
	occurrence := createTestOccurrence(models.StatusAceita, hospitalID)
	mockOccRepo.AddOccurrence(occurrence)

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(uuid.New().String(), "operador"))

	router.POST("/occurrences/:id/outcome", func(c *gin.Context) {
		idParam := c.Param("id")
		id, _ := uuid.Parse(idParam)

		var input models.RegisterOutcomeInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		occ, err := mockOccRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Validate status allows outcome
		if occ.Status != models.StatusAceita && occ.Status != models.StatusRecusada {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "outcome can only be registered for ACEITA or RECUSADA",
				"current_status": occ.Status,
			})
			return
		}

		// Create history with outcome
		historyInput := &models.CreateHistoryInput{
			OccurrenceID: id,
			Acao:         models.ActionOutcomeRegistered,
			Desfecho:     &input.Desfecho,
			Observacoes:  input.Observacoes,
		}

		history, _ := mockHistRepo.Create(c.Request.Context(), historyInput)

		c.JSON(http.StatusCreated, gin.H{
			"message":  "outcome registered",
			"desfecho": input.Desfecho,
			"history":  history.ToResponse(),
		})
	})

	// Test registering outcome
	body := models.RegisterOutcomeInput{Desfecho: models.OutcomeSucessoCaptacao}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/occurrences/%s/outcome", occurrence.ID.String()), bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify outcome was registered in history
	outcome, _ := mockHistRepo.GetOutcomeByOccurrenceID(context.Background(), occurrence.ID)
	if outcome == nil {
		t.Error("Expected outcome to be registered in history")
	}
	if outcome != nil && *outcome.Desfecho != models.OutcomeSucessoCaptacao {
		t.Errorf("Expected desfecho %s, got %s", models.OutcomeSucessoCaptacao, *outcome.Desfecho)
	}
}

// Test 6: Testar endpoint de metricas
func TestDashboardMetrics(t *testing.T) {
	mockRepo := NewMockOccurrenceRepository()
	hospitalID := uuid.New()

	// Add test occurrences with different statuses
	o1 := createTestOccurrence(models.StatusPendente, hospitalID)
	o2 := createTestOccurrence(models.StatusPendente, hospitalID)
	o3 := createTestOccurrence(models.StatusEmAndamento, hospitalID)
	mockRepo.AddOccurrence(o1)
	mockRepo.AddOccurrence(o2)
	mockRepo.AddOccurrence(o3)

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(uuid.New().String(), "gestor"))

	router.GET("/metrics/dashboard", func(c *gin.Context) {
		ctx := c.Request.Context()

		obitosPotenciais, _ := mockRepo.GetTodayEligibleCount(ctx)
		tempoMedio, _ := mockRepo.GetAverageNotificationTime(ctx)
		pendentes, _ := mockRepo.GetPendingCount(ctx)
		emAndamento, _ := mockRepo.GetEmAndamentoCount(ctx)

		metrics := &models.DashboardMetrics{
			ObitosElegiveisHoje:    obitosPotenciais,
			TempoMedioNotificacao:  tempoMedio,
			CorneasPotenciais:      obitosPotenciais * 2,
			OccurrencesPendentes:   pendentes,
			OccurrencesEmAndamento: emAndamento,
			UltimaAtualizacao:      time.Now(),
		}

		c.JSON(http.StatusOK, metrics.ToResponse())
	})

	req, _ := http.NewRequest("GET", "/metrics/dashboard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.MetricsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify metrics
	if response.ObitosElegiveisHoje != 3 {
		t.Errorf("Expected obitos_elegiveis_hoje 3, got %d", response.ObitosElegiveisHoje)
	}

	if response.CorneasPotenciais != 6 {
		t.Errorf("Expected corneas_potenciais 6, got %d", response.CorneasPotenciais)
	}

	if response.OccurrencesPendentes != 2 {
		t.Errorf("Expected occurrences_pendentes 2, got %d", response.OccurrencesPendentes)
	}

	if response.OccurrencesEmAndamento != 1 {
		t.Errorf("Expected occurrences_em_andamento 1, got %d", response.OccurrencesEmAndamento)
	}

	if response.TempoMedioNotificacao != 45.5 {
		t.Errorf("Expected tempo_medio_notificacao 45.5, got %f", response.TempoMedioNotificacao)
	}
}

// Test 7: Testar que CONCLUIDA requer desfecho
func TestConclusionRequiresOutcome(t *testing.T) {
	mockOccRepo := NewMockOccurrenceRepository()
	mockHistRepo := NewMockOccurrenceHistoryRepository()
	hospitalID := uuid.New()

	// Create occurrence in ACEITA status
	occurrence := createTestOccurrence(models.StatusAceita, hospitalID)
	mockOccRepo.AddOccurrence(occurrence)

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(uuid.New().String(), "operador"))

	router.PATCH("/occurrences/:id/status", func(c *gin.Context) {
		idParam := c.Param("id")
		id, _ := uuid.Parse(idParam)

		var input models.UpdateStatusInput
		c.ShouldBindJSON(&input)

		occ, _ := mockOccRepo.GetByID(c.Request.Context(), id)

		// Check transition validity
		if !occ.Status.CanTransitionTo(input.Status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transition"})
			return
		}

		// Special check for CONCLUIDA
		if input.Status == models.StatusConcluida {
			outcome, _ := mockHistRepo.GetOutcomeByOccurrenceID(c.Request.Context(), id)
			if outcome == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "outcome must be registered before completing",
					"hint":  "use POST /api/v1/occurrences/:id/outcome first",
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Try to transition to CONCLUIDA without outcome
	body := models.UpdateStatusInput{Status: models.StatusConcluida}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/occurrences/%s/status", occurrence.ID.String()), bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 (outcome required), got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] != "outcome must be registered before completing" {
		t.Errorf("Expected error about outcome requirement, got %v", response["error"])
	}
}

// Test 8: Testar historico de ocorrencia
func TestOccurrenceHistory(t *testing.T) {
	mockHistRepo := NewMockOccurrenceHistoryRepository()
	occurrenceID := uuid.New()
	userID := uuid.New()

	// Add history entries
	status1 := models.StatusPendente
	status2 := models.StatusEmAndamento
	status3 := models.StatusAceita

	mockHistRepo.Create(context.Background(), &models.CreateHistoryInput{
		OccurrenceID:   occurrenceID,
		UserID:         nil,
		Acao:           models.ActionOccurrenceCreated,
		StatusNovo:     &status1,
	})

	mockHistRepo.Create(context.Background(), &models.CreateHistoryInput{
		OccurrenceID:   occurrenceID,
		UserID:         &userID,
		Acao:           models.ActionOccurrenceAssigned,
		StatusAnterior: &status1,
		StatusNovo:     &status2,
	})

	mockHistRepo.Create(context.Background(), &models.CreateHistoryInput{
		OccurrenceID:   occurrenceID,
		UserID:         &userID,
		Acao:           models.ActionOccurrenceAccepted,
		StatusAnterior: &status2,
		StatusNovo:     &status3,
	})

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(userID.String(), "operador"))

	router.GET("/occurrences/:id/history", func(c *gin.Context) {
		idParam := c.Param("id")
		id, _ := uuid.Parse(idParam)

		histories, err := mockHistRepo.GetByOccurrenceID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := make([]models.OccurrenceHistoryResponse, 0, len(histories))
		for _, h := range histories {
			response = append(response, h.ToResponse())
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  response,
			"total": len(response),
		})
	})

	req, _ := http.NewRequest("GET", fmt.Sprintf("/occurrences/%s/history", occurrenceID.String()), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response struct {
		Data  []models.OccurrenceHistoryResponse `json:"data"`
		Total int                                 `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Total != 3 {
		t.Errorf("Expected 3 history entries, got %d", response.Total)
	}

	// Verify actions are in order
	expectedActions := []string{
		models.ActionOccurrenceCreated,
		models.ActionOccurrenceAssigned,
		models.ActionOccurrenceAccepted,
	}

	for i, expected := range expectedActions {
		if i < len(response.Data) && response.Data[i].Acao != expected {
			t.Errorf("Expected action %s at index %d, got %s", expected, i, response.Data[i].Acao)
		}
	}
}
