package domain

import "errors"

var (
	// Errores de actividad
	ErrActivityNotFound      = errors.New("activity not found")
	ErrActivityAlreadyExists = errors.New("activity already exists")
	ErrInvalidCategory       = errors.New("invalid category")

	// Errores de schedule
	ErrScheduleNotFound  = errors.New("schedule not found")
	ErrScheduleConflict  = errors.New("schedule conflict: time slot already occupied")
	ErrInvalidTimeFormat = errors.New("invalid time format (use HH:MM)")
	ErrInvalidDayOfWeek  = errors.New("invalid day of week")
	ErrInstructorBusy    = errors.New("instructor not available at this time")

	// Errores de validación
	ErrInvalidOwner     = errors.New("invalid owner: user not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInsufficientRole = errors.New("insufficient role permissions")

	// Errores de base de datos
	ErrDatabaseConnection = errors.New("database connection error")
	ErrDatabaseQuery      = errors.New("database query error")

	// Errores de mensajería
	ErrRabbitMQConnection = errors.New("rabbitmq connection error")
	ErrRabbitMQPublish    = errors.New("failed to publish message")
)
