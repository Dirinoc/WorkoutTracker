package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso/internal/config"
	"sso/internal/domain/models"
	"sso/internal/storage"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(cfg *config.Config) (*Storage, error) {
	const op = "storage.postgresql.New"

	ConnStr := cfg.ConnStr()

	// Подключение к бд на постгрес
	db, err := sql.Open("postgres", ConnStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.postgresql.SaveUser"

	// Подготавливаем запрос в бдшку
	stmt, err := s.db.Prepare("INSERT INTO users(email, pass_hash) VALUES($1, $2)") // Помню что используем $1, $2 для параметров в PostgreSQL
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Исполняем подготовленные выше запрос
	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // тут чут чут навайбкодил признаюсь
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// Цапаем ответ на команду
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgresql.User"

	// Подготавливаем запрос в бдшку
	stmt, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email = $1")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	// Цапаем по запросу выше строку из бд которую позже сканим на предмет нужных значений
	row := stmt.QueryRowContext(ctx, email)

	// Соотв. сканим на предмет нужных значений
	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.postgresql.IsAdmin"

	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = $1")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool
	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

// Просто шоб остановить ес чо
func (s *Storage) Stop() error {
	return s.db.Close()
}
