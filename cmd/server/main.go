package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"craftsbite-backend/internal/config"
	"craftsbite-backend/internal/database"
	"craftsbite-backend/internal/handlers"
	"craftsbite-backend/internal/middleware"
	"craftsbite-backend/internal/repository"
	"craftsbite-backend/internal/routes"
	"craftsbite-backend/internal/services"
	"craftsbite-backend/internal/sse"
	"craftsbite-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logFormat := "console"
	if cfg.IsProduction() {
		logFormat = "json"
	}
	if err := logger.Init(cfg.Logging.Level, logFormat); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Print configuration (for verification during development)
	if cfg.IsDevelopment() {
		fmt.Println("=================================")
		fmt.Println("CraftsBite Backend Configuration")
		fmt.Println("=================================")
		fmt.Printf("Environment: %s\n", cfg.Server.Env)
		fmt.Printf("Server Address: %s\n", cfg.Server.GetServerAddress())
		fmt.Printf("Database: %s@%s:%s/%s\n",
			cfg.Database.User,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
		)
		fmt.Printf("JWT Expiration: %s\n", cfg.JWT.Expiration)
		fmt.Printf("CORS Allowed Origins: %v\n", cfg.CORS.AllowedOrigins)
		fmt.Printf("Log Level: %s\n", cfg.Logging.Level)
		fmt.Println("=================================")
	}

	// Initialize database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	mealRepo := repository.NewMealRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	bulkOptOutRepo := repository.NewBulkOptOutRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	workLocationRepo := repository.NewWorkLocationRepository(db)
	workLocationHistoryRepo := repository.NewWorkLocationHistoryRepository(db)
	wfhPeriodRepo := repository.NewWFHPeriodRepository(db)

	sseHub := sse.NewHub()

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)
	userService := services.NewUserService(userRepo, teamRepo)
	participationResolver := services.NewParticipationResolver(mealRepo, scheduleRepo, bulkOptOutRepo, userRepo, cfg)
	mealService := services.NewMealService(mealRepo, scheduleRepo, historyRepo, userRepo, teamRepo, workLocationRepo, participationResolver, cfg)
	scheduleService := services.NewScheduleService(scheduleRepo)
	headcountService := services.NewHeadcountService(userRepo, scheduleRepo, participationResolver, teamRepo, workLocationRepo, wfhPeriodRepo, cfg)
	workLocationService := services.NewWorkLocationService(workLocationRepo, userRepo, teamRepo, wfhPeriodRepo, workLocationHistoryRepo, cfg)
	wfhPeriodService := services.NewWFHPeriodService(wfhPeriodRepo)

	// Phase 4: Initialize advanced feature services
	preferenceService := services.NewPreferenceService(userRepo, historyRepo)
	bulkOptOutService := services.NewBulkOptOutService(db, bulkOptOutRepo, historyRepo, teamRepo)
	historyService := services.NewHistoryService(historyRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, userService)
	userHandler := handlers.NewUserHandler(userService)
	mealHandler := handlers.NewMealHandler(mealService, teamRepo, headcountService, sseHub)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService)
	headcountHandler := handlers.NewHeadcountHandler(headcountService, sseHub)
	preferenceHandler := handlers.NewPreferenceHandler(preferenceService)
	bulkOptOutHandler := handlers.NewBulkOptOutHandler(bulkOptOutService)
	historyHandler := handlers.NewHistoryHandler(historyService)
	workLocationHandler := handlers.NewWorkLocationHandler(workLocationService)
	wfhPeriodHandler := handlers.NewWFHPeriodHandler(wfhPeriodService)

	// Phase 4: Initialize cleanup job
	// cleanupJob := jobs.NewCleanupJob(historyRepo, cfg.Cleanup.RetentionMonths)
	// cleanupScheduler, err := cleanupJob.StartScheduler(cfg.Cleanup.CronSchedule)
	// if err != nil {
	// 	log.Printf("Warning: Failed to start cleanup scheduler: %v", err)
	// }
	// defer jobs.StopScheduler(cleanupScheduler)

	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router without default middleware
	router := gin.New()

	// Apply global middleware in order
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware(cfg.CORS.AllowedOrigins))

	routes.RegisterRoutes(router, &routes.Handlers{
        Auth:       authHandler,
        User:       userHandler,
        Meal:       mealHandler,
        Schedule:   scheduleHandler,
        Headcount:  headcountHandler,
        Preference: preferenceHandler,
        BulkOptOut: bulkOptOutHandler,
        History:    historyHandler,
		WorkLocation: workLocationHandler,
		WFHPeriod:    wfhPeriodHandler,
    }, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:    cfg.Server.GetServerAddress(),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info(fmt.Sprintf("Starting CraftsBite API server on %s", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
