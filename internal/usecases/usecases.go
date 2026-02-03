package usecases

import (
	"WorkoutTracker/internal/domain/models"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Уровни логирования и их назначение:

// DEBUG — детальная отладочная информация, логирование переменных, параметров функций, входные/выходные значения (использовать только в development)
// INFO — информация о нормальной работе приложения, важные бизнес-события (создание заказа, вход пользователя)
// WARN — неожиданные, но обрабатываемые ситуации, которые не являются ошибками (retry попытки, deprecated API calls)
// ERROR — ошибки требующие внимания оператора или разработчика, но не требующие остановки приложения
// FATAL — критические ошибки, после которых приложение не может продолжать работу и должно остановиться

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
	GetWorkout(Date time.Time) (models.Workout, error)
	DeleteWorkout(workoutID int) error
}

type WorkoutService struct {
	store WorkoutStore
	log   *slog.Logger
}

// Пока такой конструктор, мб потом добавлю что-нибудь еще сюды. TODO: спросить у Дани чо оно и каво
func NewWorkoutService(store WorkoutStore) *WorkoutService {
	return &WorkoutService{
		store: store,
		log:   slog.Default(),
	}
}

func (s *WorkoutService) SaveWorkout(ID, userID int, date time.Time, exercises []models.Exercise) (int64, error) {
	const op = "usecases.WorkoutService.SaveWorkout"

	log := s.log.With(
		slog.String("op", op),
		slog.Int("workout_id", ID),
		slog.Int("user_id", userID),
		slog.Time("date", date),
		slog.Int("exercises_count", len(exercises)),
	)

	log.Info("Starting workout saving process")

	// Валидация входных данных
	if len(exercises) == 0 {
		log.Warn("Validation failed: No exercises provided in the workout")
		return 0, fmt.Errorf("%s: %w: exercises required", op, ErrInvalidRequest)
	}

	maxFuture := time.Now().Add(30 * 24 * time.Hour)
	if date.After(maxFuture) {
		log.Warn("Validation failed: Date too far in the future", slog.Time("provided_date", date))
		return 0, fmt.Errorf("%s: %w: date too far in future", op, ErrInvalidRequest)
	}

	for i, ex := range exercises {
		if ex.ExcName == "" {
			log.Warn("Validation failed: unnamed excercise")
			return 0, fmt.Errorf("%s: %w: exercises[%d].exc_name required", op, ErrInvalidRequest, i)
		}
		if ex.Sets <= 0 {
			log.Warn("Validation failed: negative or zero sets")
			return 0, fmt.Errorf("%s: %w: exercises[%d].sets >= 1 required", op, ErrInvalidRequest, i)
		}
		if ex.Reps <= 0 {
			log.Warn("Validation failed: negative or zero reps")
			return 0, fmt.Errorf("%s: %w: exercises[%d].reps >= 1 required", op, ErrInvalidRequest, i)
		}
		if ex.Weight < 0 {
			log.Warn("Validation failed: negative weight")
			return 0, fmt.Errorf("%s: %w: exercises[%d].weight >= 0 required", op, ErrInvalidRequest, i)
		}
	}

	// Передает валидированные данные на хранение
	workoutID, err := s.store.SaveWorkout(ID, userID, date, exercises)
	if err != nil {
		log.Error("workout save failed")
		return 0, fmt.Errorf("%s: storage save: %w", op, err)
	}

	return workoutID, nil
}

// GetWorkout — получает тренировку по дате (без проверки владельца)
func (s *WorkoutService) GetWorkout(date time.Time) (models.Workout, error) {
	const op = "service.WorkoutService.GetWorkout"

	log := s.log.With(
		slog.String("op", op),
		slog.Time("date", date),
	)

	log.Info("workout fetching process begin")

	w, err := s.store.GetWorkout(date)
	if err != nil {
		log.Error("failed to fetch workout")
		return models.Workout{}, fmt.Errorf("%s: %w", op, err)
	}
	return w, nil
}

// DeleteWorkout — удаляет тренировку по id (без проверки владельца)
func (s *WorkoutService) DeleteWorkout(workoutID int) error {
	const op = "service.WorkoutService.DeleteWorkout"

	log := s.log.With(
		slog.String("op", op),
		slog.Int("workout_id", workoutID),
	)

	log.Info("workout deletion process begin")

	if err := s.store.DeleteWorkout(workoutID); err != nil {
		log.Error("failed to delete workout")
		return fmt.Errorf("%s: delete workout: %w", op, err)
	}
	return nil
}
