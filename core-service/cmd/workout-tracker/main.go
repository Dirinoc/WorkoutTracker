package main

import (
	"WorkoutTracker/internal/config"
	"WorkoutTracker/internal/storage/postgresql"
	"WorkoutTracker/internal/transport/http/handlers"
	"WorkoutTracker/internal/usecases"
	"context"
	"log/slog"
	"os"

	ssogrpc "WorkoutTracker/internal/clients/sso/grpc"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {

	// Инициализация логгера
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Загружаем конфиг
	cfg, err := config.MustLoad()

	if err != nil {
		log.Error("failed to load config", slog.Any("err", err))
		os.Exit(1)
	}

	log.Info("config initialised successfully")

	ssoClient, err := ssogrpc.New(
		context.Background(),
		log,
		cfg.Clients.SSO.Address,
		cfg.Clients.SSO.Timeout,
		cfg.Clients.SSO.RetriesCount,
	)
	if err != nil {
		log.Error("failed to init sso client", slog.Any("err", err))
		os.Exit(1)
	}

	ssoClient.IsAdmin(context.Background(), 1)

	storage, err := postgresql.New(cfg)
	if err != nil {
		panic("failed to initialize storage: " + err.Error())
	}

	log.Info("storage initialised successfully")

	workoutService := usecases.NewWorkoutService(storage)

	router := gin.Default()

	// Group routes under /workouts
	workoutRoutes := router.Group("/workouts")
	{
		workoutRoutes.POST("/", handlers.SaveWorkout(workoutService))        // Save a workout
		workoutRoutes.GET("/:id", handlers.GetWorkout(workoutService))       // Get a workout by ID
		workoutRoutes.DELETE("/:id", handlers.DeleteWorkout(workoutService)) // Delete a workout by ID
	}

	router.Run(":8080")
}
