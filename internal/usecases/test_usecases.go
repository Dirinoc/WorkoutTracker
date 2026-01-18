package usecases

import (
	"WorkoutTracker/internal/domain/models"
	"testing"
	"time"
)

// Мок структура которая реализует интерфейс WorkoutStore. В ней хранятся данные о работе функций, подвергаемых тестированию.
type mockWorkoutStore struct {
	saveCalled   bool
	getCalled    bool
	deleteCalled bool

	saveWorkout int64
	saveErr     error

	getWorkoutID models.Workout
	getErr       error

	delErr error
}

// Далее, 3 функции которые реализует интерфейс ServiceStore через mockWorkoutStore
func (m *mockWorkoutStore) SaveWorkout(ID, userID int, date time.Time, exercises []models.Exercise) (int64, error) {
	m.saveCalled = true
	return m.saveWorkout, m.saveErr
}

func (m *mockWorkoutStore) GetWorkout(workoutID int) (models.Workout, error) {
	m.getCalled = true
	return m.getWorkoutID, m.getErr
}

func (m *mockWorkoutStore) DeleteWorkout(workoutID int) error {
	m.deleteCalled = true
	return m.delErr
}

// Тесты для WorkoutService. Берем адрес &mockWorkoutStore c айди 52, затем вызываем NewWorkoutService по адресу (&) 52, и проводим дальнейшие операции.
func TestWorkoutService_SaveWorkout(t *testing.T) {
	store := &mockWorkoutStore{
		saveWorkout: 52,
	}

	service := NewWorkoutService(store)

	exercises := []models.Exercise{
		{
			ExcName: "Bench Press",
			Sets:    3,
			Reps:    10,
			Weight:  80,
		},
	}

	id, err := service.SaveWorkout(
		0,
		1,
		time.Now(),
		exercises,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 52 {
		t.Fatalf("expected workoutID=52, got %d", id)
	}

	if !store.saveCalled {
		t.Fatal("expected SaveWorkout to be called")
	}

}

// Тест вызова GetWorkout, задаем ожидаемое значение (expected ID 10), создаем mockWorkoutStore с этим значением, вызываем сервис и проверяем результат.
func TestWorkoutService_GetWorkout(t *testing.T) {
	expected := models.Workout{ID: 10}

	store := &mockWorkoutStore{
		getWorkoutID: expected,
	}

	service := NewWorkoutService(store)

	w, err := service.GetWorkout(10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.ID != expected.ID {
		t.Fatalf("expected workout ID %d, got %d", expected.ID, w.ID)
	}

	if !store.getCalled {
		t.Fatal("expected GetWorkout to be called")
	}
}

// Тест вызова DeleteWorkout, создаем mockWorkoutStore, вызываем сервис и проверяем результат.
func TestWorkoutService_DeleteWorkout(t *testing.T) {
	store := &mockWorkoutStore{}

	service := NewWorkoutService(store)

	err := service.DeleteWorkout(10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !store.deleteCalled {
		t.Fatal("expected DeleteWorkout to be called")
	}
}
