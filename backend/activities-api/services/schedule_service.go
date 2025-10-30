package services

import (
	"activities-api/domain"
	"activities-api/repositories"
)

// ScheduleService define la interface para el servicio de horarios
type ScheduleService interface {
	// CreateSchedule crea un nuevo horario (CON CONCURRENCIA)
	CreateSchedule(activityID string, req domain.CreateScheduleRequest) (*domain.ScheduleResponse, error)

	// GetScheduleByID obtiene un horario por su ID
	GetScheduleByID(id string) (*domain.ScheduleResponse, error)

	// GetSchedulesByActivityID obtiene todos los horarios de una actividad
	GetSchedulesByActivityID(activityID string) ([]domain.ScheduleResponse, error)

	// GetAllSchedules obtiene todos los horarios
	GetAllSchedules() ([]domain.ScheduleResponse, error)

	// UpdateSchedule actualiza un horario
	UpdateSchedule(id string, req domain.UpdateScheduleRequest, userID uint, userRole string) (*domain.ScheduleResponse, error)

	// DeleteSchedule elimina un horario
	DeleteSchedule(id string, userID uint, userRole string) error

	// UpdateCurrentBookings actualiza el contador de reservas
	UpdateCurrentBookings(scheduleID string, increment bool) error

	// Nuevo método para inyectar el publisher
	SetPublisher(publisher *repositories.RabbitMQPublisher)
}
