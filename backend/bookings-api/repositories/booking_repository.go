package repositories

import "bookings-api/domain"

// BookingRepository define las operaciones de persistencia para bookings
type BookingRepository interface {
	// Create crea una nueva reserva
	Create(booking *domain.Booking) error

	// GetByID obtiene una reserva por su ID
	GetByID(id string) (*domain.Booking, error)

	// GetByUserID obtiene todas las reservas de un usuario
	GetByUserID(userID uint) ([]domain.Booking, error)

	// GetByScheduleID obtiene todas las reservas de un horario
	GetByScheduleID(scheduleID string) ([]domain.Booking, error)

	// CheckDuplicateBooking verifica si ya existe una reserva del usuario para ese schedule
	CheckDuplicateBooking(userID uint, scheduleID string) (bool, error)

	// Update actualiza una reserva
	Update(booking *domain.Booking) error

	// Delete elimina una reserva (soft delete cambiando status)
	Delete(id string) error
}
