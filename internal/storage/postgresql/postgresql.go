package postgresql

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Storage struct {
	db *sql.DB
}

type Excercise struct {
	ExcName string
	Weight  float64
	Sets    int
	Reps    int
}

type Workout struct {
	ID         int
	UserID     int
	Date       time.Time
	Excercises []Excercise
}

// Creating a new instance of storage (a database on PostgreSQL)
func New() (*Storage, error) {

	const op = "storage.postgresql.New"

	// Загрузка переменных окружения из .env файла
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("%s: error loading .env file %w", op, err)
	}

	// Запись переменных окружения для формирования connStr (строки подключения)
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbSSL := os.Getenv("DB_SSLMODE")

	// Формирование connStr (строки подключения) из переменных окружения
	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		dbUser, dbPassword, dbName, dbHost, dbPort, dbSSL,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println("Connected to PostgreSQL database successfully")

	// Создать базу данных и таблицы, если они не существуют

	CreateTable := `

	CREATE TABLE IF NOT EXISTS user_info (
		id SERIAL PRIMARY KEY,
		passhash TEXT NOT NULL,
		is_admin BOOLEAN NOT NULL DEFAULT FALSE,
		registered_at TIMESTAMP NOT NULL DEFAULT NOW(),
	);

	CREATE TABLE IF NOT EXISTS workouts (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		date TIMESTAMP NOT NULL,
	);
	
	CREATE TABLE IF NOT EXISTS excercises (
		id SERIAL PRIMARY KEY,
		workoutsid INT NOT NULL,
		exc_name TEXT NOT NULL,
		weight FLOAT NOT NULL,
		sets INT NOT NULL,
		reps INT NOT NULL,
	);`

	if _, err := db.Exec(CreateTable); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println("Ensured workouts table exists")

	return &Storage{db: db}, nil
}

// Сохранение тренировки в базу данных
func (s *Storage) SaveWorkout(user_id int, date time.Time, excercise string, weight int, sets int, reps int) (string, error) {

	const op = "storage.postgresql.SaveWorkout"

	//Начало работы с дб (как я понял) с возможностью отката (rollback) в случае ошибки
	tx, err := s.db.Begin()
	if err != nil {
		return "", fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback()

	// Вставка тренировки в таблицу workouts
	var workoutID int
	err = tx.QueryRow(
		"INSERT INTO workouts(user_id, date) VALUES($1, $2) RETURNING id",
		user_id,
		date,
	).Scan(&workoutID) // Получает айдишку только что вставленной тренировки чтобы использовать её позже для записи упражнения
	if err != nil {
		return "", fmt.Errorf("%s: insert workout: %w", op, err)
	}

	// Вставка упражнения в таблицу excercises
	_, err = tx.Exec(
		"INSERT INTO excercises(workoutsid, exc_name, weight, sets, reps) VALUES($1, $2, $3, $4, $5)",
		workoutID,
		excercise,
		weight,
		sets,
		reps,
	)
	if err != nil {
		return "", fmt.Errorf("%s: insert excercise: %w", op, err)
	}

	// Смотрим и подтверждаем указанные выше действия (так называемые транзакции)
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("%s: commit tx: %w", op, err)
	}

	return fmt.Sprintf("Workout saved with id: %d", workoutID), nil
}

// Получить тренировку из базы данных по ID
func (s *Storage) GetWorkout(workoutid int) (*Workout, error) {

	const op = "storage.postgresql.GetWorkout"

	var w Workout
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer tx.Rollback()

	var workoutID int
	err = tx.QueryRow(
		"SELECT id, user_id, date FROM workouts WHERE id = $1",
		workoutid,
	).Scan(&workoutID)
	if err != nil {
		return nil, fmt.Errorf("%s: select workout: %w", op, err)
	}

	rows, err := tx.Query(
		"SELECT exc_name, weight, sets, reps FROM excercises WHERE workoutsid = $1",
		workoutID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: select excercises: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var ex Excercise
		if err := rows.Scan(&ex.ExcName, &ex.Weight, &ex.Sets, &ex.Reps); err != nil {
			return nil, fmt.Errorf("%s: scan excercise: %w", op, err)
		}
		w.Excercises = append(w.Excercises, ex)
	}

	return &w, nil
}
