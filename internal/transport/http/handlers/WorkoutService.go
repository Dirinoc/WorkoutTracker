package handlers

import (
	"WorkoutTracker/internal/domain/models"
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

		var req Request

		// Bind JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			errs := make(map[string]string)
			for _, fe := range validateErr {
				errs[fe.Field()] = fe.Tag()
			}

			c.JSON(http.StatusBadRequest, gin.H{"validation_errors": errs})
			return
		}

		// ID is auto-generated in DB — ignore req.ID
		id, err := workout.SaveWorkout(0, req.UserID, req.Date, req.Exercises)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, Response{WorkoutID: int(id)})
	}
}

// Delete workout using date (ID for now will change for date later)
func DeleteWorkout(workout WorkoutService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Extract workout ID from URL
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
			return
		}

		// Delete workout via service
		err = workout.DeleteWorkout(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to delete workout",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Workout deleted successfully",
			"workout_id": id,
		})
	}
}

func GetWorkout(workout WorkoutService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Extract workout ID from URL: /workouts/:id
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
			return
		}

		// Fetch workout from service
		w, err := workout.GetWorkout(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
			return
		}

		// Send full workout struct to client
		c.JSON(http.StatusOK, w)
	}
}
