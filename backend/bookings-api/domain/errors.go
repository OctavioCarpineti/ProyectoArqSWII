package domain

import "errors"

var (
	// Errores de booking
	ErrBookingNotFound      = errors.New("booking not found")
	ErrBookingAlreadyExists = errors.New("booking already exists for this user and schedule")
	ErrScheduleFull         = errors.New("schedule is full, no available spots")

	// Errores de validación
	ErrInvalidUser     = errors.New("invalid user: user not found")
	ErrInvalidSchedule = errors.New("invalid schedule: schedule not found")
	ErrUnauthorized    = errors.New("unauthorized")

	// Errores de base de datos
	ErrDatabaseConnection = errors.New("database connection error")
	ErrDatabaseQuery      = errors.New("database query error")
)
