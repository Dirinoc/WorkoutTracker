package postgresql

import (
	"WorkoutTracker/internal/domain/models"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"WorkoutTracker/internal/config"
)

type Storage struct {
	db *sql.DB
}

// Creating a new instance of storage (a database on PostgreSQL)
func New(cfg *config.Config) (*Storage, error) {
	const op = "storage.postgresql.New"

	// Получаем connStr из структуры
	connStr := cfg.ConnStr()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Создать базу данных и таблицы, если они не существуют

	CreateTable := `
	CREATE TABLE IF NOT EXISTS user_info (
		id SERIAL PRIMARY KEY,
		passhash TEXT NOT NULL,
		is_admin BOOLEAN NOT NULL DEFAULT FALSE,
		registered_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS workouts (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		date TIMESTAMP NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS exercises (
		id SERIAL PRIMARY KEY,
		workoutsid INT NOT NULL,
		exc_name TEXT NOT NULL,
		weight FLOAT NOT NULL,
		sets INT NOT NULL,
		reps INT NOT NULL
	);
`

	if _, err := db.Exec(CreateTable); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// Сохранение тренировки в базу данных
func (s *Storage) SaveWorkout(ID int, user_id int, date time.Time, exercise []models.Exercise) (int64, error) {

	const op = "storage.postgresql.SaveWorkout"

	//Начало работы с дб (как я понял) с возможностью отката (rollback) в случае ошибки
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback()

	// Вставка тренировки в таблицу workouts
	var workoutID int64
	err = tx.QueryRow(
		"INSERT INTO workouts(user_id, date) VALUES($1, $2) RETURNING id",
		user_id,
		date,
	).Scan(&workoutID) // Получает айдишку только что вставленной тренировки чтобы использовать её позже для записи упражнения
	if err != nil {
		return 0, fmt.Errorf("%s: insert workout: %w", op, err)
	}

	// Вставка упражнения в таблицу exercises
	// TODO: Batch insert
	for _, ex := range exercise {
		_, err = tx.Exec(
			"INSERT INTO exercises(workoutsid, exc_name, weight, sets, reps) VALUES($1, $2, $3, $4, $5)",
			workoutID,
			ex.ExcName,
			ex.Weight,
			ex.Sets,
			ex.Reps,
		)
		if err != nil {
			return 0, fmt.Errorf("%s: insert exercise: %w", op, err)
		}
	}

	// Смотрим и подтверждаем указанные выше действия (так называемые транзакции)
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("%s: commit tx: %w", op, err)
	}

	return workoutID, nil
}

// Получить тренировку из базы данных по ID
// TODO: fix returns (bandaid fix for now returning models.Workout{}) - Change to request workout by Date
// TODO: What if there are n workouts in one day?
func (s *Storage) GetWorkout(workoutid int) (models.Workout, error) {
	const op = "storage.postgresql.GetWorkout"

	var w models.Workout

	// Fetch workout info
	err := s.db.QueryRow(
		"SELECT id, user_id, date FROM workouts WHERE id = $1",
		workoutid,
	).Scan(&w.ID, &w.UserID, &w.Date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Workout{}, fmt.Errorf("%s: workout not found: %w", op, err)
		}
		return models.Workout{}, fmt.Errorf("%s: select workout: %w", op, err)
	}

	// Fetch exercises
	rows, err := s.db.Query(
		"SELECT exc_name, weight, sets, reps FROM exercises WHERE workoutsid = $1",
		w.ID,
	)
	if err != nil {
		return models.Workout{}, fmt.Errorf("%s: select exercises: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var ex models.Exercise
		if err := rows.Scan(&ex.ExcName, &ex.Weight, &ex.Sets, &ex.Reps); err != nil {
			return models.Workout{}, fmt.Errorf("%s: scan exercise: %w", op, err)
		}
		w.Exercises = append(w.Exercises, ex)
	}

	// No transaction needed → no rollback/commit
	return w, nil
}

func (s *Storage) DeleteWorkout(workoutid int) error {
	const op = "storage.postgresql.DeleteWorkout"

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"DELETE FROM exercises WHERE workoutsid = $1",
		workoutid,
	)

	_, err = tx.Exec(
		"DELETE FROM workouts WHERE id = $1",
		workoutid,
	)

	if err != nil {
		return fmt.Errorf("%s: delete exercises: %w", op, err)
	}

	return nil
}
