package domain

import "errors"

var (
	// Errores de usuario
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Errores de validación
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidPassword = errors.New("password must be at least 6 characters")
	ErrInvalidUsername = errors.New("username must be between 3 and 50 characters")

	// Errores de token
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
	ErrUnauthorized = errors.New("unauthorized")

	// Errores de base de datos
	ErrDatabaseConnection = errors.New("database connection error")
	ErrDatabaseQuery      = errors.New("database query error")
)
