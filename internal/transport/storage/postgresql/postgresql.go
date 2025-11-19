package postgresql

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Storage struct {
	db *sql.DB
}

// Creating a new instance of storage (a database on PostgreSQL)
func New() (*Storage, error) {

	const op = "storage.postgresql.New"

	// Загрузка переменных окружения из .env файла
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("%s: error loading .env file %w", op, err)
	}

	// Получение переменных окружения
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbSSL := os.Getenv("DB_SSLMODE")

	// Формируем connStr (строку подключения) из переменных окружения
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		dbUser, dbPassword, dbName, dbHost, dbPort, dbSSL)

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
