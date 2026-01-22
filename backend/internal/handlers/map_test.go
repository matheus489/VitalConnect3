package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sidot/backend/internal/models"
)

// MockMapRepository simula o repositorio do mapa para testes
type MockMapRepository struct {
	hospitals   []models.Hospital
	occurrences []models.Occurrence
	shifts      []models.TodayShift
}

func NewMockMapRepository() *MockMapRepository {
	return &MockMapRepository{
		hospitals:   []models.Hospital{},
		occurrences: []models.Occurrence{},
		shifts:      []models.TodayShift{},
	}
}

func (r *MockMapRepository) AddHospital(h models.Hospital) {
	r.hospitals = append(r.hospitals, h)
}

func (r *MockMapRepository) AddOccurrence(o models.Occurrence) {
	r.occurrences = append(r.occurrences, o)
}

func (r *MockMapRepository) AddShift(s models.TodayShift) {
	r.shifts = append(r.shifts, s)
}

func (r *MockMapRepository) GetActiveHospitalsWithCoordinates(ctx context.Context) ([]models.Hospital, error) {
	var result []models.Hospital
	for _, h := range r.hospitals {
		if h.Ativo && h.Latitude != nil && h.Longitude != nil && h.DeletedAt == nil {
			result = append(result, h)
		}
	}
	return result, nil
}

func (r *MockMapRepository) GetActiveOccurrencesByHospitalID(ctx context.Context, hospitalID uuid.UUID) ([]models.Occurrence, error) {
	var result []models.Occurrence
	for _, o := range r.occurrences {
		if o.HospitalID == hospitalID && (o.Status == models.StatusPendente || o.Status == models.StatusEmAndamento) {
			result = append(result, o)
		}
	}
	return result, nil
}

func (r *MockMapRepository) GetCurrentOperatorByHospitalID(ctx context.Context, hospitalID uuid.UUID) (*models.TodayShift, error) {
	for _, s := range r.shifts {
		if s.HospitalID == hospitalID && s.IsActive {
			return &s, nil
		}
	}
	return nil, nil
}

// createTestHospitalWithCoordinates cria um hospital de teste com coordenadas
func createTestHospitalWithCoordinates(nome, codigo string, lat, lng float64, ativo bool) models.Hospital {
	return models.Hospital{
		ID:        uuid.New(),
		Nome:      nome,
		Codigo:    codigo,
		Latitude:  &lat,
		Longitude: &lng,
		Ativo:     ativo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// createTestOccurrenceForMap cria uma ocorrencia de teste para o mapa
func createTestOccurrenceForMap(hospitalID uuid.UUID, status models.OccurrenceStatus, hoursRemaining int) models.Occurrence {
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
		JanelaExpiraEm:        time.Now().Add(time.Duration(hoursRemaining) * time.Hour),
	}
}

// createTestTodayShift cria um plantao ativo de teste
func createTestTodayShift(hospitalID, userID uuid.UUID, userName string, isActive bool) models.TodayShift {
	return models.TodayShift{
		Shift: models.Shift{
			ID:         uuid.New(),
			HospitalID: hospitalID,
			UserID:     userID,
			DayOfWeek:  models.DayOfWeek(time.Now().Weekday()),
			StartTime:  "07:00",
			EndTime:    "19:00",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			User: &models.User{
				ID:    userID,
				Nome:  userName,
				Email: "operador@test.com",
				Role:  models.RoleOperador,
				Ativo: true,
			},
		},
		IsActive: isActive,
	}
}

// Test 1: Testar endpoint GET /api/v1/map/hospitals retorna hospitais ativos
func TestMapEndpointReturnsActiveHospitals(t *testing.T) {
	mockRepo := NewMockMapRepository()

	// Adicionar hospitais de teste
	lat1, lng1 := -16.6868, -49.2648
	lat2, lng2 := -16.7200, -49.3000
	hospital1 := createTestHospitalWithCoordinates("Hospital Geral de Goiania", "HGG", lat1, lng1, true)
	hospital2 := createTestHospitalWithCoordinates("Hospital de Urgencias", "HUGO", lat2, lng2, true)
	mockRepo.AddHospital(hospital1)
	mockRepo.AddHospital(hospital2)

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(uuid.New().String(), "operador"))

	router.GET("/api/v1/map/hospitals", func(c *gin.Context) {
		ctx := c.Request.Context()

		hospitals, err := mockRepo.GetActiveHospitalsWithCoordinates(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar hospitais"})
			return
		}

		var mapHospitals []models.MapHospitalResponse
		for _, h := range hospitals {
			mapHospitals = append(mapHospitals, models.MapHospitalResponse{
				ID:        h.ID,
				Nome:      h.Nome,
				Codigo:    h.Codigo,
				Latitude:  *h.Latitude,
				Longitude: *h.Longitude,
				Ativo:     h.Ativo,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  mapHospitals,
			"total": len(mapHospitals),
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/map/hospitals", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado status 200, recebido %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	total := int(response["total"].(float64))
	if total != 2 {
		t.Errorf("Esperado 2 hospitais, recebido %d", total)
	}
}

// Test 2: Testar que hospitais inativos nao aparecem no resultado
func TestMapEndpointExcludesInactiveHospitals(t *testing.T) {
	mockRepo := NewMockMapRepository()

	// Adicionar hospitais - um ativo e um inativo
	lat1, lng1 := -16.6868, -49.2648
	lat2, lng2 := -16.7200, -49.3000
	hospitalAtivo := createTestHospitalWithCoordinates("Hospital Ativo", "HA", lat1, lng1, true)
	hospitalInativo := createTestHospitalWithCoordinates("Hospital Inativo", "HI", lat2, lng2, false)
	mockRepo.AddHospital(hospitalAtivo)
	mockRepo.AddHospital(hospitalInativo)

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(uuid.New().String(), "operador"))

	router.GET("/api/v1/map/hospitals", func(c *gin.Context) {
		ctx := c.Request.Context()

		hospitals, _ := mockRepo.GetActiveHospitalsWithCoordinates(ctx)

		var mapHospitals []models.MapHospitalResponse
		for _, h := range hospitals {
			mapHospitals = append(mapHospitals, models.MapHospitalResponse{
				ID:        h.ID,
				Nome:      h.Nome,
				Codigo:    h.Codigo,
				Latitude:  *h.Latitude,
				Longitude: *h.Longitude,
				Ativo:     h.Ativo,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  mapHospitals,
			"total": len(mapHospitals),
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/map/hospitals", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	total := int(response["total"].(float64))
	if total != 1 {
		t.Errorf("Esperado apenas 1 hospital ativo, recebido %d", total)
	}
}

// Test 3: Testar agregacao de ocorrencias ativas por hospital
func TestMapEndpointAggregatesActiveOccurrences(t *testing.T) {
	mockRepo := NewMockMapRepository()

	lat, lng := -16.6868, -49.2648
	hospital := createTestHospitalWithCoordinates("Hospital Teste", "HT", lat, lng, true)
	mockRepo.AddHospital(hospital)

	// Adicionar ocorrencias - 2 ativas e 1 concluida
	o1 := createTestOccurrenceForMap(hospital.ID, models.StatusPendente, 5)
	o2 := createTestOccurrenceForMap(hospital.ID, models.StatusEmAndamento, 3)
	o3 := createTestOccurrenceForMap(hospital.ID, models.StatusConcluida, 1)
	mockRepo.AddOccurrence(o1)
	mockRepo.AddOccurrence(o2)
	mockRepo.AddOccurrence(o3)

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(uuid.New().String(), "operador"))

	router.GET("/api/v1/map/hospitals", func(c *gin.Context) {
		ctx := c.Request.Context()

		hospitals, _ := mockRepo.GetActiveHospitalsWithCoordinates(ctx)

		var mapHospitals []models.MapHospitalResponse
		for _, h := range hospitals {
			occurrences, _ := mockRepo.GetActiveOccurrencesByHospitalID(ctx, h.ID)

			mapHospitals = append(mapHospitals, models.MapHospitalResponse{
				ID:               h.ID,
				Nome:             h.Nome,
				Codigo:           h.Codigo,
				Latitude:         *h.Latitude,
				Longitude:        *h.Longitude,
				Ativo:            h.Ativo,
				OcorrenciasCount: len(occurrences),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  mapHospitals,
			"total": len(mapHospitals),
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/map/hospitals", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response struct {
		Data  []models.MapHospitalResponse `json:"data"`
		Total int                          `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)

	if len(response.Data) != 1 {
		t.Fatalf("Esperado 1 hospital, recebido %d", len(response.Data))
	}

	// Apenas 2 ocorrencias ativas (PENDENTE e EM_ANDAMENTO)
	if response.Data[0].OcorrenciasCount != 2 {
		t.Errorf("Esperado 2 ocorrencias ativas, recebido %d", response.Data[0].OcorrenciasCount)
	}
}

// Test 4: Testar calculo de urgencia maxima (verde/amarelo/vermelho)
func TestMapEndpointCalculatesMaxUrgency(t *testing.T) {
	testCases := []struct {
		name             string
		hoursRemaining   int
		expectedUrgency  models.UrgencyLevel
	}{
		{"Verde (>4h)", 5, models.UrgencyGreen},
		{"Amarelo (2-4h)", 3, models.UrgencyYellow},
		{"Vermelho (<2h)", 1, models.UrgencyRed},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := NewMockMapRepository()

			lat, lng := -16.6868, -49.2648
			hospital := createTestHospitalWithCoordinates("Hospital Teste", "HT", lat, lng, true)
			mockRepo.AddHospital(hospital)

			occurrence := createTestOccurrenceForMap(hospital.ID, models.StatusPendente, tc.hoursRemaining)
			mockRepo.AddOccurrence(occurrence)

			ctx := context.Background()
			hospitals, _ := mockRepo.GetActiveHospitalsWithCoordinates(ctx)

			for _, h := range hospitals {
				occurrences, _ := mockRepo.GetActiveOccurrencesByHospitalID(ctx, h.ID)

				urgencia := models.CalculateMaxUrgency(occurrences)

				if urgencia != tc.expectedUrgency {
					t.Errorf("Esperado urgencia %s, recebido %s", tc.expectedUrgency, urgencia)
				}
			}
		})
	}
}

// Test 5: Testar inclusao de dados do operador de plantao
func TestMapEndpointIncludesOnDutyOperator(t *testing.T) {
	mockRepo := NewMockMapRepository()

	lat, lng := -16.6868, -49.2648
	hospital := createTestHospitalWithCoordinates("Hospital Teste", "HT", lat, lng, true)
	mockRepo.AddHospital(hospital)

	userID := uuid.New()
	shift := createTestTodayShift(hospital.ID, userID, "Maria Operadora", true)
	mockRepo.AddShift(shift)

	router := setupTestRouter()
	router.Use(mockAuthMiddleware(uuid.New().String(), "operador"))

	router.GET("/api/v1/map/hospitals", func(c *gin.Context) {
		ctx := c.Request.Context()

		hospitals, _ := mockRepo.GetActiveHospitalsWithCoordinates(ctx)

		var mapHospitals []models.MapHospitalResponse
		for _, h := range hospitals {
			resp := models.MapHospitalResponse{
				ID:        h.ID,
				Nome:      h.Nome,
				Codigo:    h.Codigo,
				Latitude:  *h.Latitude,
				Longitude: *h.Longitude,
				Ativo:     h.Ativo,
			}

			// Buscar operador de plantao
			shift, _ := mockRepo.GetCurrentOperatorByHospitalID(ctx, h.ID)
			if shift != nil && shift.User != nil {
				resp.OperadorPlantao = &models.MapOperatorResponse{
					ID:     shift.User.ID,
					Nome:   shift.User.Nome,
					UserID: shift.UserID,
				}
			}

			mapHospitals = append(mapHospitals, resp)
		}

		c.JSON(http.StatusOK, gin.H{
			"data":  mapHospitals,
			"total": len(mapHospitals),
		})
	})

	req, _ := http.NewRequest("GET", "/api/v1/map/hospitals", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response struct {
		Data  []models.MapHospitalResponse `json:"data"`
		Total int                          `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)

	if len(response.Data) != 1 {
		t.Fatalf("Esperado 1 hospital, recebido %d", len(response.Data))
	}

	if response.Data[0].OperadorPlantao == nil {
		t.Error("Esperado operador de plantao presente")
	} else if response.Data[0].OperadorPlantao.Nome != "Maria Operadora" {
		t.Errorf("Esperado operador 'Maria Operadora', recebido '%s'", response.Data[0].OperadorPlantao.Nome)
	}
}

// Test 6: Testar filtragem por status de ocorrencia (apenas PENDENTE e EM_ANDAMENTO)
func TestMapEndpointFiltersOccurrencesByActiveStatus(t *testing.T) {
	mockRepo := NewMockMapRepository()

	lat, lng := -16.6868, -49.2648
	hospital := createTestHospitalWithCoordinates("Hospital Teste", "HT", lat, lng, true)
	mockRepo.AddHospital(hospital)

	// Adicionar ocorrencias com todos os status
	statuses := []models.OccurrenceStatus{
		models.StatusPendente,
		models.StatusEmAndamento,
		models.StatusAceita,
		models.StatusRecusada,
		models.StatusCancelada,
		models.StatusConcluida,
	}

	for _, status := range statuses {
		o := createTestOccurrenceForMap(hospital.ID, status, 3)
		mockRepo.AddOccurrence(o)
	}

	ctx := context.Background()
	activeOccurrences, _ := mockRepo.GetActiveOccurrencesByHospitalID(ctx, hospital.ID)

	// Apenas PENDENTE e EM_ANDAMENTO devem ser retornados
	if len(activeOccurrences) != 2 {
		t.Errorf("Esperado 2 ocorrencias ativas, recebido %d", len(activeOccurrences))
	}

	for _, o := range activeOccurrences {
		if o.Status != models.StatusPendente && o.Status != models.StatusEmAndamento {
			t.Errorf("Status inesperado: %s", o.Status)
		}
	}
}
