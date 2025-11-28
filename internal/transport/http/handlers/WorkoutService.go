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

type Request struct {
	ID         int
	UserID     int
	Date       time.Time
	Excercises []models.Excercise
}

// TODO: Учтонить чо тут писать именно, пока что затычка такая, я понимаю что это ответ сервера клиенту но хз
type Response struct {
	WorkoutID int `json:"WorkoutID,omitempty"`
}

type WorkoutSaver interface {
	SaveWorkout(ID, UserID int, Date time.Time, Excercises []models.Excercise) (int64, error)
}

// Создать новый хендлер (gin.HandlerFunc allows use of go funcs as http handlers)
// TODO: Сделать нормальные логи, разобраться как писать ошибки нормально, пока так (Line 36, 45)
func New(workout WorkoutSaver) gin.HandlerFunc {
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
		id, err := workout.SaveWorkout(req.ID, req.UserID, req.Date, req.Excercises)
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

func DeleteWorkout(req Request) {
	const op = "handlers.workout.delete.DeleteWorkout"

}
