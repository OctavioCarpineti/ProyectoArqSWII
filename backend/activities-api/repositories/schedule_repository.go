package repositories

import "activities-api/domain"

// ScheduleRepository define la interface para el repositorio de horarios
type ScheduleRepository interface {
	// Create crea un nuevo horario
	Create(schedule *domain.Schedule) error

	// GetByID obtiene un horario por su ID
	GetByID(id string) (*domain.Schedule, error)

	// GetByActivityID obtiene todos los horarios de una actividad
	GetByActivityID(activityID string) ([]domain.Schedule, error)

	// GetAll obtiene todos los horarios
	GetAll() ([]domain.Schedule, error)

	// Update actualiza un horario existente
	Update(schedule *domain.Schedule) error

	// Delete elimina un horario
	Delete(id string) error

	// CheckConflict verifica si hay conflicto de horario/sala
	CheckConflict(dayOfWeek, startTime, endTime, location string, excludeID string) (bool, error)

	// IncrementBookings incrementa el contador de reservas
	IncrementBookings(id string) error

	// DecrementBookings decrementa el contador de reservas
	DecrementBookings(id string) error
}
