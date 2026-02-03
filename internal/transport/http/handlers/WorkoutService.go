package handlers

import (
	"WorkoutTracker/internal/domain/models"
	"fmt"
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
	GetWorkout(Date time.Time) (models.Workout, error)
}

// Respond with error in JSON format (внутренний пакет)
func RespondErrorJSON(c *gin.Context, code int, msg string) error {
	c.JSON(code, gin.H{"error": msg})
	return nil
}

// Создать новый хендлер (gin.HandlerFunc allows use of go funcs as http handlers)
// TODO: Condense funcs into one (preferably understand it fully)
func SaveWorkout(workout WorkoutService) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req Request

		// Bind JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			RespondErrorJSON(c, http.StatusBadRequest, "invalid request payload")
			return
		}

		// Validate
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			errs := make(map[string]string)
			for _, fe := range validateErr {
				errs[fe.Field()] = fe.Tag()
			}

			RespondErrorJSON(c, http.StatusBadRequest, "Validation failed: "+fmt.Sprint(errs))
			return
		}

		// ID is auto-generated in DB — ignore req.ID
		id, err := workout.SaveWorkout(0, req.UserID, req.Date, req.Exercises)
		if err != nil {
			RespondErrorJSON(c, http.StatusInternalServerError, "Saving failed")
			return
		}

		RespondErrorJSON(c, http.StatusOK, "Workout saved successfully"+fmt.Sprintf(", ID: %d", id))
	}
}

// Delete workout using date (ID for now will change for date later)
func DeleteWorkout(workout WorkoutService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Extract workout ID from URL
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			RespondErrorJSON(c, http.StatusBadRequest, "invalid workout id")
			return
		}

		// Delete workout via service
		err = workout.DeleteWorkout(id)
		if err != nil {
			RespondErrorJSON(c, http.StatusInternalServerError, "failed to delete workout")
			return
		}

		RespondErrorJSON(c, http.StatusOK, "Workout deleted successfully")
	}
}

func GetWorkout(workout WorkoutService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Extract workout ID from URL: /workouts/:id
		dateStr := c.Param("date")

		// Parsing the date
		date, err := time.Parse("2006-01-02", dateStr)

		if err != nil {
			RespondErrorJSON(c, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
			return
		}

		// Fetch workout from service
		w, err := workout.GetWorkout(date)
		if err != nil {
			RespondErrorJSON(c, http.StatusNotFound, "workout not found")
			return
		}

		// Send full workout struct to client
		RespondErrorJSON(c, http.StatusOK, fmt.Sprintf("%+v", w))
	}
}
