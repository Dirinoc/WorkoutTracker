package main

import (
	"WorkoutTracker/internal/config"
	"WorkoutTracker/internal/storage/postgresql"
	"WorkoutTracker/internal/transport/http/handlers"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {

	// Загружаем конфиг
	cfg, err := config.MustLoad()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("Config initialised successfully")

	storage, err := postgresql.New(cfg)
	if err != nil {
		panic("failed to initialize storage: " + err.Error())
	}
	log.Println("Storage initialised successfully")

	_ = storage

	router := gin.Default()

	workoutService := storage

	// Group routes under /workouts
	workoutRoutes := router.Group("/workouts")
	{
		workoutRoutes.POST("/", handlers.SaveWorkout(workoutService))        // Save a workout
		workoutRoutes.GET("/:id", handlers.GetWorkout(workoutService))       // Get a workout by ID
		workoutRoutes.DELETE("/:id", handlers.DeleteWorkout(workoutService)) // Delete a workout by ID
	}

	router.Run(":8080")

}
