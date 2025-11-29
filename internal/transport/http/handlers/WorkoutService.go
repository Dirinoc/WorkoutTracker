package handlers

import (
	"WorkoutTracker/internal/domain/models"
	"fmt"
	"log/slog"
	"net/http"
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
// TODO: Сделать нормальные логи, разобраться как писать ошибки нормально, пока так (Line 36, 45)
func New(workout WorkoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.workout.save.New"

		var req Request

		if err := c.ShouldBindJSON(&req); err != nil {
			slog.Error("failed to decode request body", slog.String("op", op))

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "failed to decode request",
				"details": err.Error(),
			})
		}

		fmt.Printf("request body decoded")

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			slog.Error("invalid request", slog.String("op", op))

			// Общая ошибка, написана пока так
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})

			// Детализированные ошибки валидации (создаем массив заполняемый ошибками)
			errs := make(map[string]string)
			for _, fe := range validateErr {
				errs[fe.Field()] = fe.Tag() // например: "Email": "required"
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"validation_errors": errs,
			})

			return
		}

		// TODO: сделать автоматическое заполнение айди тренировки

		// Сохраняем тренировку через интерфейс и передаем пользователю айди созданной тренировки
		id, err := workout.SaveWorkout(req.ID, req.UserID, req.Date, req.Exercises)
		if err != nil {
			slog.Error("failed to save workout", slog.String("op", op))

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to save workout",
			})

			return
		}

		slog.Info("workout saved", slog.String("op", op), slog.Int64("id", id))

		c.JSON(http.StatusOK, Response{
			WorkoutID: int(id),
		})
	}
}

// Delete workout using date
func DeleteWorkout(workout WorkoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.workout.delete.DeleteWorkout"

		var req WorkoutIDRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			slog.Error("failed to decode request body", slog.String("op", op))

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "failed to decode request",
				"details": err.Error(),
			})
		}

		fmt.Printf("request body decoded")

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			slog.Error("invalid request", slog.String("op", op))

			// Общая ошибка, написана пока так
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})

			// Детализированные ошибки валидации (создаем массив заполняемый ошибками)
			errs := make(map[string]string)
			for _, fe := range validateErr {
				errs[fe.Field()] = fe.Tag() // например: "Email": "required"
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"validation_errors": errs,
			})
		}

		// Удаляем тренировку через интерфейс
		err := workout.DeleteWorkout(req.WorkoutID)
		if err != nil {
			slog.Error("failed to save workout", slog.String("op", op))

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to save workout",
			})

			return
		}

		slog.Info("workout saved", slog.String("op", op), slog.Int64("id", int64(req.WorkoutID)))

		c.JSON(http.StatusOK, Response{
			WorkoutID: int(req.WorkoutID),
		})

	}

}

func GetWorkout(workout WorkoutService) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.workout.get.GetWorkout"

		var req WorkoutIDRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			slog.Error("failed to decode request body", slog.String("op", op))

			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "failed to decode request",
				"details": err.Error(),
			})
		}

		fmt.Printf("request body decoded")

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			slog.Error("invalid request", slog.String("op", op))

			// Общая ошибка, написана пока так
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})

			// Детализированные ошибки валидации (создаем массив заполняемый ошибками)
			errs := make(map[string]string)
			for _, fe := range validateErr {
				errs[fe.Field()] = fe.Tag() // например: "Email": "required"
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"validation_errors": errs,
			})
		}

		_, err := workout.GetWorkout(req.WorkoutID)
		if err != nil {
			slog.Error("failed to get workout", slog.String("op", op))

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get workout",
			})

			return
		}

		slog.Info("workout found", slog.String("op", op), slog.Int64("id", int64(req.WorkoutID)))

		c.JSON(http.StatusOK, Response{
			WorkoutID: int(req.WorkoutID),
		})

	}
}
