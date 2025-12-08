package usecases

import (
	"WorkoutTracker/internal/domain/models"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidRequest  = errors.New("invalid request")
	ErrInvalidDate     = errors.New("invalid date")
	ErrNoExercises     = errors.New("workout must contain at least one exercise")
	ErrNegativeValue   = errors.New("exercise values must be non-negative")
	ErrWorkoutNotFound = errors.New("workout not found")
)

// TODO: chatgpt также давал txID. зачем? Выяснить
// WorkoutStore описывает минимальный набор операций, которые нужны сервису.
// Это интерфейс для инверсии зависимости от конкретной реализации хранилища.
type WorkoutStore interface {
	SaveWorkout(ID, userID int, date time.Time, exercises []models.Exercise) (int64, error)
	GetWorkout(workoutID int) (models.Workout, error)
	DeleteWorkout(workoutID int) error
}

type WorkoutService struct {
	store WorkoutStore
}

// Пока такой конструктор, мб потом добавлю что-нибудь еще сюды
func NewWorkoutService(store WorkoutStore) *WorkoutService {
	return &WorkoutService{
		store: store,
	}
}

func (s *WorkoutService) SaveWorkout(ID, userID int, date time.Time, exercises []models.Exercise) (int64, error) {
	const op = "usecases.WorkoutService.SaveWorkout"

	// Валидация
	if len(exercises) == 0 {
		return 0, fmt.Errorf("%s: %w: exercises required", op, ErrInvalidRequest)
	}

	maxFuture := time.Now().Add(30 * 24 * time.Hour)
	if date.After(maxFuture) {
		return 0, fmt.Errorf("%s: %w: date too far in future", op, ErrInvalidRequest)
	}

	for i, ex := range exercises {
		if ex.ExcName == "" {
			return 0, fmt.Errorf("%s: %w: exercises[%d].exc_name required", op, ErrInvalidRequest, i)
		}
		if ex.Sets <= 0 {
			return 0, fmt.Errorf("%s: %w: exercises[%d].sets >= 1 required", op, ErrInvalidRequest, i)
		}
		if ex.Reps <= 0 {
			return 0, fmt.Errorf("%s: %w: exercises[%d].reps >= 1 required", op, ErrInvalidRequest, i)
		}
		if ex.Weight < 0 {
			return 0, fmt.Errorf("%s: %w: exercises[%d].weight >= 0 required", op, ErrInvalidRequest, i)
		}
	}

	// Передает валидированные данные на хранение
	workoutID, err := s.store.SaveWorkout(ID, userID, date, exercises)
	if err != nil {
		return 0, fmt.Errorf("%s: storage save: %w", op, err)
	}

	return workoutID, nil
}

// GetWorkout — получает тренировку по id (без проверки владельца)
func (s *WorkoutService) GetWorkout(workoutID int) (models.Workout, error) {
	const op = "service.WorkoutService.GetWorkout"

	w, err := s.store.GetWorkout(workoutID)
	if err != nil {
		return models.Workout{}, fmt.Errorf("%s: %w", op, err)
	}
	return w, nil
}

// DeleteWorkout — удаляет тренировку по id (без проверки владельца)
func (s *WorkoutService) DeleteWorkout(workoutID int) error {
	const op = "service.WorkoutService.DeleteWorkout"

	if err := s.store.DeleteWorkout(workoutID); err != nil {
		return fmt.Errorf("%s: delete workout: %w", op, err)
	}
	return nil
}
