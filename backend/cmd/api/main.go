package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/vitalconnect/backend/config"
	"github.com/vitalconnect/backend/internal/handlers"
	"github.com/vitalconnect/backend/internal/middleware"
	"github.com/vitalconnect/backend/internal/models"
	"github.com/vitalconnect/backend/internal/repository"
	"github.com/vitalconnect/backend/internal/services/audit"
	"github.com/vitalconnect/backend/internal/services/auth"
	"github.com/vitalconnect/backend/internal/services/health"
	"github.com/vitalconnect/backend/internal/services/listener"
	"github.com/vitalconnect/backend/internal/services/notification"
	"github.com/vitalconnect/backend/internal/services/report"
	"github.com/vitalconnect/backend/internal/services/triagem"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Printf("Warning: Database ping failed: %v", err)
	}

	// Initialize Redis client
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Failed to parse Redis URL: %v, using defaults", err)
		redisOpts = &redis.Options{
			Addr: "localhost:6379",
			DB:   0,
		}
	}
	redisClient := redis.NewClient(redisOpts)

	// Test Redis connection
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Printf("Warning: Redis ping failed: %v", err)
	}
	defer redisClient.Close()

	// Initialize JWT service
	jwtService, err := auth.NewJWTService(
		cfg.JWTSecret,
		cfg.JWTRefreshSecret,
		cfg.JWTAccessDuration,
		cfg.JWTRefreshDuration,
	)
	if err != nil {
		log.Fatalf("Failed to initialize JWT service: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	hospitalRepo := repository.NewHospitalRepository(db)
	occurrenceRepo := repository.NewOccurrenceRepository(db)
	occurrenceHistoryRepo := repository.NewOccurrenceHistoryRepository(db)
	triagemRuleRepo := repository.NewTriagemRuleRepository(db, redisClient)
	indicatorsRepo := repository.NewIndicatorsRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	shiftRepo := repository.NewShiftRepository(db)
	pushSubRepo := repository.NewPushSubscriptionRepository(db)

	// Initialize auth service
	authService := auth.NewAuthService(jwtService, userRepo, redisClient)

	// Initialize auth handler and set global handler
	authHandler := handlers.NewAuthHandler(authService)
	handlers.SetGlobalAuthHandler(authHandler)

	// Set repositories for handlers
	handlers.SetHospitalRepository(hospitalRepo)
	handlers.SetUserRepository(userRepo)
	handlers.SetOccurrenceRepository(occurrenceRepo)
	handlers.SetOccurrenceHistoryRepository(occurrenceHistoryRepo)
	handlers.SetTriagemRuleRepository(triagemRuleRepo)
	handlers.SetMetricsOccurrenceRepository(occurrenceRepo)
	handlers.SetIndicatorsRepository(indicatorsRepo)
	handlers.SetAuditLogRepository(auditLogRepo)

	// Initialize audit service
	auditService := audit.NewAuditService(auditLogRepo)
	handlers.SetAuditService(auditService)

	// Initialize report service
	reportService := report.NewReportService(db)
	handlers.SetReportService(reportService)

	// Initialize shift handler
	shiftHandler := handlers.NewShiftHandler(shiftRepo, userRepo)

	// Initialize map handler for geographic dashboard
	mapHandler := handlers.NewMapHandler(hospitalRepo, occurrenceRepo, shiftRepo)

	// Initialize Push Notification Service
	pushConfig := &notification.PushConfig{
		ServerKey: cfg.FCMServerKey,
	}
	pushService := notification.NewPushService(pushConfig)
	handlers.SetPushService(pushService)
	handlers.SetPushSubscriptionRepository(pushSubRepo)

	if cfg.IsFCMConfigured() {
		log.Println("[PushService] FCM push notifications enabled")
	} else {
		log.Println("[PushService] FCM not configured - push notifications disabled (set FCM_SERVER_KEY)")
	}

	// Initialize PEP Integration
	handlers.SetPEPRedisClient(redisClient)
	// TODO: Load PEP API keys from database or configuration
	// For now, use empty map - can be configured via hospital settings
	handlers.SetPEPAPIKeys(make(map[string]uuid.UUID))
	log.Println("[PEP] PEP integration endpoint initialized")

	// Initialize SSE Hub for real-time notifications
	sseHub := notification.NewSSEHub(redisClient, db)
	handlers.SetGlobalSSEHub(sseHub)

	// Initialize Email Service
	emailConfig := &notification.EmailConfig{
		SMTPHost:     cfg.SMTPHost,
		SMTPPort:     cfg.SMTPPort,
		SMTPUser:     cfg.SMTPUser,
		SMTPPassword: cfg.SMTPPassword,
		SMTPFrom:     cfg.SMTPFrom,
	}
	emailService := notification.NewEmailService(emailConfig)

	// Initialize Email Queue Worker
	emailQueueWorker := notification.NewEmailQueueWorker(redisClient, emailService, db)

	// Initialize and start obito listener
	obitoListener := listener.NewObitoListener(db, redisClient, cfg.ListenerPollInterval)
	handlers.SetGlobalListener(obitoListener)

	// Initialize and start triagem motor
	triagemMotor := triagem.NewTriagemMotor(db, redisClient)
	handlers.SetGlobalTriagemMotor(triagemMotor)

	// Initialize Health Monitor Service
	healthMonitor := health.NewHealthMonitorService(db, redisClient, emailService, cfg.AdminAlertEmail)
	healthMonitor.SetSSEHub(sseHub)
	healthMonitor.SetListener(obitoListener)
	healthMonitor.SetTriagemMotor(triagemMotor)
	healthMonitor.SetCheckInterval(cfg.HealthCheckInterval)
	healthMonitor.SetCooldownPeriod(time.Duration(cfg.AlertCooldownMinutes) * time.Minute)
	handlers.SetGlobalHealthMonitor(healthMonitor)

	// Set callback for new occurrences to trigger SSE notifications
	triagemMotor.SetOnOccurrenceCreated(func(ctx context.Context, occurrence *models.Occurrence, hospitalNome string) {
		// Publish SSE event for dashboard notifications
		if err := sseHub.PublishNewOccurrence(ctx, occurrence, hospitalNome); err != nil {
			log.Printf("Warning: Failed to publish SSE event: %v", err)
		}

		// Queue email notifications for operators if email service is configured
		if emailService.IsConfigured() {
			// Get operators to notify (you could filter by hospital if needed)
			operators, err := userRepo.ListByRole(ctx, "operador")
			if err != nil {
				log.Printf("Warning: Failed to get operators for email notification: %v", err)
				return
			}

			// Get occurrence details for email
			var completeData models.OccurrenceCompleteData
			if err := json.Unmarshal(occurrence.DadosCompletos, &completeData); err != nil {
				log.Printf("Warning: Failed to parse occurrence data for email: %v", err)
				return
			}

			emailData := &notification.ObitoNotificationData{
				HospitalNome:  hospitalNome,
				Setor:         completeData.Setor,
				HoraObito:     occurrence.DataObito,
				TempoRestante: occurrence.FormatTimeRemaining(),
				OccurrenceID:  occurrence.ID.String(),
				Prioridade:    occurrence.ScorePriorizacao,
				DashboardURL:  "http://localhost:3000/dashboard", // Configure via env
			}

			for _, operator := range operators {
				userID := operator.ID
				if err := emailQueueWorker.EnqueueEmail(ctx, occurrence.ID, operator.Email, &userID, emailData); err != nil {
					log.Printf("Warning: Failed to queue email for %s: %v", operator.Email, err)
				}
			}
		}
	})

	// Create context for background services
	ctx, cancelBackground := context.WithCancel(context.Background())

	// Start background services
	if err := obitoListener.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start obito listener: %v", err)
	}

	if err := triagemMotor.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start triagem motor: %v", err)
	}

	if err := sseHub.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start SSE hub: %v", err)
	}

	if err := emailQueueWorker.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start email queue worker: %v", err)
	}

	// Start health monitor service
	if err := healthMonitor.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start health monitor: %v", err)
	}
	log.Println("[HealthMonitor] Health monitor service initialized")

	// Initialize router
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORS(cfg.CORSOrigins))
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.SetJWTService(jwtService))

	// Health check endpoint (basic)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		authRoutes := v1.Group("/auth")
		{
			// Apply rate limiting to login endpoint
			authRoutes.POST("/login", middleware.LoginRateLimit(redisClient, cfg.LoginRateLimit), handlers.Login)
			authRoutes.POST("/refresh", handlers.RefreshToken)
			authRoutes.POST("/logout", handlers.Logout)
			authRoutes.GET("/me", middleware.AuthRequired(), handlers.Me)
		}

		// SSE stream with query param authentication (for EventSource - no auth header support)
		v1.GET("/notifications/stream", handlers.NotificationStream)

		// Public health summary endpoint (for load balancers)
		v1.GET("/health/summary", handlers.HealthSummary)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// Hospitals
			hospitals := protected.Group("/hospitals")
			{
				hospitals.GET("", handlers.ListHospitals)
				hospitals.GET("/:id", handlers.GetHospital)
				hospitals.POST("", middleware.RequireRole("admin"), handlers.CreateHospital)
				hospitals.PATCH("/:id", middleware.RequireRole("admin"), handlers.UpdateHospital)
				hospitals.DELETE("/:id", middleware.RequireRole("admin"), handlers.DeleteHospital)
			}

			// Users
			users := protected.Group("/users")
			{
				users.GET("", middleware.RequireRole("admin"), handlers.ListUsers)
				users.GET("/:id", handlers.GetUser)
				users.POST("", middleware.RequireRole("admin"), handlers.CreateUser)
				users.PATCH("/:id", handlers.UpdateUser)
				users.DELETE("/:id", middleware.RequireRole("admin"), handlers.DeleteUser)
			}

			// Occurrences
			occurrences := protected.Group("/occurrences")
			{
				occurrences.GET("", handlers.ListOccurrences)
				occurrences.GET("/:id", handlers.GetOccurrence)
				occurrences.GET("/:id/history", handlers.GetOccurrenceHistory)
				occurrences.PATCH("/:id/status", handlers.UpdateOccurrenceStatus)
				occurrences.POST("/:id/outcome", handlers.RegisterOutcome)
			}

			// Triagem Rules
			rules := protected.Group("/triagem-rules")
			{
				rules.GET("", middleware.RequireRole("gestor", "admin"), handlers.ListTriagemRules)
				rules.POST("", middleware.RequireRole("gestor", "admin"), handlers.CreateTriagemRule)
				rules.PATCH("/:id", middleware.RequireRole("gestor", "admin"), handlers.UpdateTriagemRule)
				rules.DELETE("/:id", middleware.RequireRole("gestor", "admin"), handlers.DeleteTriagemRule)
			}

			// Metrics
			protected.GET("/metrics/dashboard", handlers.GetDashboardMetrics)
			protected.GET("/metrics/indicators", handlers.GetIndicators)

			// Health checks (protected - for detailed info)
			protected.GET("/health/listener", handlers.ListenerHealth)
			protected.GET("/health/sse", handlers.SSEHealth)

			// Shifts (plantoes)
			shifts := protected.Group("/shifts")
			{
				shifts.POST("", middleware.RequireRole("admin", "gestor"), shiftHandler.Create)
				shifts.GET("/:id", shiftHandler.GetByID)
				shifts.PUT("/:id", middleware.RequireRole("admin", "gestor"), shiftHandler.Update)
				shifts.DELETE("/:id", middleware.RequireRole("admin", "gestor"), shiftHandler.Delete)
				shifts.GET("/me", shiftHandler.GetMyShifts)
			}

			// Hospital-specific shift routes
			protected.GET("/hospitals/:id/shifts", shiftHandler.ListByHospital)
			protected.GET("/hospitals/:id/shifts/today", shiftHandler.GetTodayShifts)
			protected.GET("/hospitals/:id/shifts/coverage", shiftHandler.GetCoverageGaps)

			// Map routes (Dashboard Geografico)
			mapRoutes := protected.Group("/map")
			{
				mapRoutes.GET("/hospitals", mapHandler.GetMapHospitals)
			}

			// Audit Logs
			protected.GET("/audit-logs", middleware.RequireRole("admin", "gestor"), handlers.ListAuditLogs)
			protected.GET("/occurrences/:id/timeline", handlers.GetOccurrenceTimeline)

			// Reports
			reports := protected.Group("/reports")
			{
				reports.GET("/csv", middleware.RequireRole("admin", "gestor"), handlers.ExportCSV)
				reports.GET("/pdf", middleware.RequireRole("admin", "gestor"), handlers.ExportPDF)
			}

			// Push Notifications
			push := protected.Group("/push")
			{
				push.POST("/subscribe", handlers.SubscribePush)
				push.DELETE("/unsubscribe", handlers.UnsubscribePush)
				push.GET("/subscriptions", handlers.GetMySubscriptions)
				push.GET("/status", handlers.GetPushStatus)
			}
		}

		// PEP Integration (API Key authentication, not user auth)
		pep := v1.Group("/pep")
		{
			pep.POST("/eventos", handlers.ReceivePEPEvent)
			pep.GET("/status", handlers.GetPEPStatus)
		}
	}

	// Create HTTP server with longer timeouts for SSE
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // Disable for SSE (long-lived connections)
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("VitalConnect API server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Cancel background services context
	cancelBackground()

	// Stop background services
	obitoListener.Stop()
	triagemMotor.Stop()
	sseHub.Stop()
	emailQueueWorker.Stop()
	healthMonitor.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
