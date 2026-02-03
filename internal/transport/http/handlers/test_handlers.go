package handlers

import (
	"WorkoutTracker/internal/domain/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type mockWorkoutService struct {
	saveCalled bool
	getCalled  bool
	delCalled  bool

	returnWorkout models.Workout
	requestDate   time.Time
	getErr        error

	saveResult int64
	saveErr    error

	delReqID int
	delErr   error
}

func (m *mockWorkoutService) SaveWorkout(ID, UserID int, Date time.Time, Exercises []models.Exercise) (int64, error) {
	m.saveCalled = true
	return m.saveResult, m.saveErr
}

func (m *mockWorkoutService) DeleteWorkout(workoutID int) error {
	m.delCalled = true     // фиксируем факт вызова
	m.delReqID = workoutID // записываем айди из URL
	return nil
}

func (m *mockWorkoutService) GetWorkout(date time.Time) (models.Workout, error) {
	m.getCalled = true   // фиксируем факт вызова
	m.requestDate = date // записываем айди из URL
	return m.returnWorkout, m.getErr
}

func TestWorkoutService_SaveWorkout(t *testing.T) {

	gin.SetMode(gin.TestMode)

	router := gin.New()

	mockService := &mockWorkoutService{
		saveResult: 1,
		saveErr:    nil,
	}

	router.POST("/workouts", SaveWorkout(mockService))

	body := `{
    "user_id": 1,
    "date": "2024-01-01T00:00:00Z",
    "exercises": []
	}`

	// Аналогично Save запросу в постмане. 1-я строка - метод (Save), 2 строка - адрес, 3 строка парсит body в json. Только для запросов у которых есть тело.
	req := httptest.NewRequest(
		http.MethodPost,
		"/workouts",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !mockService.saveCalled {
		t.Fatalf("expected SaveWorkout to be called")
	}
}

func TestWorkoutService_GetWorkout(t *testing.T) {

	gin.SetMode(gin.TestMode)

	router := gin.New()

	expectedDate := time.Date(2026, time.February, 4, 12, 0, 0, 0, time.UTC)

	// mockService это НЕ РЕЗУЛЬТАТ ТЕСТА, это, по сути, его сценарий. Мы забиваем сюда данные для проведения теста. В данном случае - мы хотим получить тренировку с датой 4 февраля 2006 года. Уточнить у Дани
	mockService := &mockWorkoutService{
		returnWorkout: models.Workout{
			Date: time.Date(2026, time.February, 4, 12, 0, 0, 0, time.UTC),
		},
		getErr: nil,
	}

	router.GET("/workouts/:id", GetWorkout(mockService))

	req := httptest.NewRequest(
		http.MethodGet,
		"/workouts/10",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !mockService.getCalled {
		t.Fatalf("expected GetWorkout to be called")
	}

	if !mockService.requestDate.Equal(expectedDate) {
		t.Fatalf("expected date %v, got %v", expectedDate, mockService.requestDate)
	}
}

func TestWorkoutService_DeleteWorkout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	mockService := &mockWorkoutService{
		delErr: nil,
	}

	router.DELETE("/workouts/:id", DeleteWorkout(mockService))

	req := httptest.NewRequest(
		http.MethodDelete,
		"/workouts/10",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !mockService.delCalled {
		t.Fatalf("expected DeleteWorkout to be called")
	}

	if mockService.delReqID != 10 {
		t.Fatalf("expected id 10, got %d", mockService.delReqID)
	}
}
