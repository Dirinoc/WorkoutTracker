package postgresql

import (
	"database/sql"
	"fmt"
)

type Storage struct {
	db *sql.DB
}

// Creating a new instance of storage (a database on PostgreSQL)
func New() (*Storage, error) {

	const op = "storage.postgresql.New"

	connStr := "user=postgres password=32853835 dbname=workoutracker host=localhost port=5431 sslmode=disable"

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
	CREATE TABLE IF NOT EXISTS workouts (
		id SERIAL PRIMARY KEY,
		exercise_name TEXT NOT NULL,
		repetitions INT NOT NULL,
		weight FLOAT NOT NULL,
		workout_date TIMESTAMP NOT NULL
	);`

	if _, err := db.Exec(CreateTable); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println("Ensured workouts table exists")

	return &Storage{db: db}, nil
}
