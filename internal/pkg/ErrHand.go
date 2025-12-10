package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Структура ошибки приложения. В ней держим код ошибки HTTP, сообщение для пользователя без лишних подробностей, и суть ошибки, которую пользователь не видит (Code, Message, Error соответственно)
type AppError struct {
	// Код статуса http
	Code int `json:"-"`
	// То, что показываем пользователю (без лишних деталей)
	Message string `json:"message"`
	// Суть ошибки. Для дебаггинга.
	Err error `json:"-"`
}

// JSON который возвращаем пользователю. НЕ ЗАПОЛНЯТЬ Err
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code,omitempty"`
}

// Если чота не так то компилятор по башке даст
var _ error = (*AppError)(nil)

func (a *AppError) Error() string {
	if a == nil {
		return "<nil AppError>"
	}
	if a.Err != nil {
		return fmt.Sprintf("%s: %v", a.Message, a.Err)
	}
	return a.Message
}

// Берет структуру ошибки и выдает из нее только Err (т. е. поле, созданное для содержание информации об ошибке для дебаггинга)
func (a *AppError) Unwrap() error {
	if a == nil {
		return nil
	}
	return a.Err
}

// Создает и заполняет новую структуру ошибки
func New(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Конструкторы для http кодов (самые частые) TODO: Добавить forbidden когда сделаю grpc auth
func BadRequest(message string, err error) *AppError {
	return New(http.StatusBadRequest, message, err)
}

func Unauthorized(message string, err error) *AppError {
	return New(http.StatusUnauthorized, message, err)
}

func NotFound(message string, err error) *AppError {
	return New(http.StatusNotFound, message, err)
}

// Очень важно - без неё пользователю может возвращаться конфиденциальная информация о БД
func Internal(message string, err error) *AppError {
	return New(http.StatusInternalServerError, message, nil)
}

// Проверяем является ли ошибка AppError (или оберткой AppError). TODO: узнать зачем оно в целом нужно. Понимаю только что это частая практика, но не совсем подозреваю где можно применять
func IsAppError(err error) bool {
	var a *AppError
	return errors.As(err, &a)
}

// Конвертирует ошибку в сообщение для пользователя
// Использует AppError.Message по возможности, иначе выдает стандартное сообщение
func ToErrorResponse(err error) ErrorResponse {
	if err == nil {
		return ErrorResponse{Error: ""}
	}
	var appErr *AppError
	// Разворачивает ошибку err и ищет в ней AppError. Если она есть то записывает её в appErr и возвращает true.
	if errors.As(err, &appErr) {
		// Возвращает Message
		return ErrorResponse{Error: appErr.Message}
	}
	// При незнакомой ошибке возвращает стандартный текст. Позволяет избежать слива кода/конфиденциальной информации.
	return ErrorResponse{Error: http.StatusText(http.StatusInternalServerError)}
}

// StatusCode returns the status code for the provided error.
// If the error is an AppError it returns its Code; otherwise it returns 500.
func StatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Code == 0 {
			return http.StatusInternalServerError
		}
		return appErr.Code
	}
	// default for unknown errors
	return http.StatusInternalServerError
}

// WriteHTTPError writes the provided error as a JSON HTTP response using the proper status code.
// It sets "Content-Type: application/json" and writes a safe client-visible message.
// If marshaling fails it falls back to a minimal text response.
func WriteHTTPError(w http.ResponseWriter, err error) {
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Определяем, какой HTTP-код вернуть.
	code := StatusCode(err)

	// Создаём JSON-тело, которое отправим клиенту.
	resp := ToErrorResponse(err)

	// Устанавливаем заголовок, чтобы клиент понимал что это JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	// Отправляем json пользователю
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, resp.Error, code)
	}
}
