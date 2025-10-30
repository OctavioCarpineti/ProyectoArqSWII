package services

import "bookings-api/domain"

// BookingService define la lógica de negocio para bookings
type BookingService interface {
	// CreateBooking crea una nueva reserva con validaciones concurrentes
	CreateBooking(req domain.CreateBookingRequest, token string) (*domain.BookingResponse, error)

	// GetBookingByID obtiene una reserva por su ID
	GetBookingByID(id string) (*domain.BookingResponse, error)

	// GetBookingsByUser obtiene todas las reservas de un usuario
	GetBookingsByUser(userID uint) ([]domain.BookingResponse, error)

	// CancelBooking cancela una reserva
	CancelBooking(id string, userID uint, userRole string, token string) error
}
