package handlers

import (
	"WorkoutTracker/internal/domain/models"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type (
	Request struct {
		ID        int               `json:"id"`
		UserID    int               `json:"user_id" validate:"required"`
		Date      time.Time         `json:"date" validate:"required"`
		Exercises []models.Exercise `json:"exercises" validate:"required,dive"`
	}

	WorkoutIDRequest struct {
		WorkoutID int `json:"workout_id" validate:"required"`
	}

	Response struct {
		WorkoutID int `json:"workout_id,omitempty"`
	}
)

type WorkoutService interface {
	SaveWorkout(ID, UserID int, Date time.Time, Exercises []models.Exercise) (int64, error)
	DeleteWorkout(WorkoutID int) error
	GetWorkout(WorkoutID int) (models.Workout, error)
}

// Создать новый хендлер (gin.HandlerFunc allows use of go funcs as http handlers)
// TODO: remove logs (unnecessary) and const ops
// TODO: Condense funcs into one (preferably understand it fully)
func SaveWorkout(workout WorkoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.workout.save.New"

		var req Request

		// Bind JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			slog.Error("failed to decode request body", slog.String("op", op), slog.Any("err", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		slog.Info("request body decoded", slog.String("op", op))

		// Validate
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			errs := make(map[string]string)
			for _, fe := range validateErr {
				errs[fe.Field()] = fe.Tag()
			}

			slog.Error("invalid request", slog.String("op", op), slog.Any("err", errs))
			c.JSON(http.StatusBadRequest, gin.H{"validation_errors": errs})
			return
		}

		// ID is auto-generated in DB — ignore req.ID
		id, err := workout.SaveWorkout(0, req.UserID, req.Date, req.Exercises)
		if err != nil {
			slog.Error("failed to save workout", slog.String("op", op), slog.Any("err", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		slog.Info("workout saved", slog.String("op", op), slog.Int64("id", id))

		c.JSON(http.StatusOK, Response{WorkoutID: int(id)})
	}
}

// Delete workout using date (ID for now will change for date later)
func DeleteWorkout(workout WorkoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.workout.delete.DeleteWorkout"

		// Extract workout ID from URL
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			slog.Error("invalid workout id", slog.String("op", op))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
			return
		}

		// Delete workout via service
		err = workout.DeleteWorkout(id)
		if err != nil {
			slog.Error("failed to delete workout", slog.String("op", op), slog.Int("id", id))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to delete workout",
			})
			return
		}

		slog.Info("workout deleted", slog.String("op", op), slog.Int("id", id))

		c.JSON(http.StatusOK, gin.H{
			"message":    "Workout deleted successfully",
			"workout_id": id,
		})
	}
}

func GetWorkout(workout WorkoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.workout.get.GetWorkout"

		// Extract workout ID from URL: /workouts/:id
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			slog.Error("invalid workout id", slog.String("op", op))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
			return
		}

		// Fetch workout from service
		w, err := workout.GetWorkout(id)
		if err != nil {
			slog.Error("failed to get workout", slog.String("op", op), slog.Int("id", id))
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
			return
		}

		slog.Info("workout found", slog.String("op", op), slog.Int("id", id))

		// Send full workout struct to client
		c.JSON(http.StatusOK, w)
	}
}
