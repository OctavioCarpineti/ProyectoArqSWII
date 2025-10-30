package services

import (
	"bookings-api/clients"
	"bookings-api/domain"
	"bookings-api/repositories"
	"fmt"
	"log"
	"sync"
)

// BookingServiceImpl implementa BookingService
type BookingServiceImpl struct {
	bookingRepo      repositories.BookingRepository
	userClient       *clients.UserClient
	activitiesClient *clients.ActivitiesClient
}

// NewBookingService crea una nueva instancia del servicio
func NewBookingService(
	bookingRepo repositories.BookingRepository,
	userClient *clients.UserClient,
	activitiesClient *clients.ActivitiesClient,
) BookingService {
	return &BookingServiceImpl{
		bookingRepo:      bookingRepo,
		userClient:       userClient,
		activitiesClient: activitiesClient,
	}
}

// CreateBooking crea una nueva reserva con validaciones concurrentes
func (s *BookingServiceImpl) CreateBooking(req domain.CreateBookingRequest, token string) (*domain.BookingResponse, error) {
	log.Printf("📋 Iniciando creación de reserva para user_id=%d, schedule_id=%s", req.UserID, req.ScheduleID)

	// ============================================
	// CONCURRENCIA: Validaciones en paralelo
	// ============================================

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Variables para almacenar resultados de las validaciones
	var user *clients.UserResponse
	var schedule *clients.ScheduleResponse
	var activity *clients.ActivityResponse

	// Goroutine 1: Validar que el usuario existe
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("🔍 Goroutine 1: Validando usuario %d...", req.UserID)

		u, err := s.userClient.GetUserByID(req.UserID)
		if err != nil {
			log.Printf("❌ Goroutine 1: Usuario no encontrado: %v", err)
			errChan <- domain.ErrInvalidUser
			return
		}
		user = u
		log.Printf("✅ Goroutine 1: Usuario %d existe (%s)", user.ID, user.Username)
	}()

	// Goroutine 2: Validar que el schedule existe y tiene cupos disponibles
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("🔍 Goroutine 2: Validando schedule %s...", req.ScheduleID)

		sch, err := s.activitiesClient.GetScheduleByID(req.ScheduleID)
		if err != nil {
			log.Printf("❌ Goroutine 2: Schedule no encontrado: %v", err)
			errChan <- domain.ErrInvalidSchedule
			return
		}

		// Verificar que tenga cupos disponibles
		if sch.CurrentBookings >= sch.MaxCapacity {
			log.Printf("❌ Goroutine 2: Schedule lleno (%d/%d)", sch.CurrentBookings, sch.MaxCapacity)
			errChan <- domain.ErrScheduleFull
			return
		}

		// Verificar que el schedule esté activo
		if sch.Status != "active" {
			log.Printf("❌ Goroutine 2: Schedule no está activo (status=%s)", sch.Status)
			errChan <- fmt.Errorf("schedule is not active")
			return
		}

		schedule = sch
		log.Printf("✅ Goroutine 2: Schedule válido - %d/%d cupos ocupados", sch.CurrentBookings, sch.MaxCapacity)
	}()

	// Goroutine 3: Validar que el usuario no tenga una reserva duplicada
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("🔍 Goroutine 3: Verificando duplicados para user=%d, schedule=%s...", req.UserID, req.ScheduleID)

		exists, err := s.bookingRepo.CheckDuplicateBooking(req.UserID, req.ScheduleID)
		if err != nil {
			log.Printf("❌ Goroutine 3: Error al verificar duplicados: %v", err)
			errChan <- fmt.Errorf("error checking duplicate: %w", err)
			return
		}
		if exists {
			log.Printf("❌ Goroutine 3: Reserva duplicada encontrada")
			errChan <- domain.ErrBookingAlreadyExists
			return
		}
		log.Printf("✅ Goroutine 3: No hay reservas duplicadas")
	}()

	// Esperar a que todas las goroutines terminen
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Verificar si hubo errores
	for err := range errChan {
		if err != nil {
			log.Printf("❌ Error en validación concurrente: %v", err)
			return nil, err
		}
	}

	log.Printf("✅ Todas las validaciones concurrentes pasaron exitosamente")

	// ============================================
	// Obtener información de la actividad para desnormalizar
	// ============================================

	log.Printf("📥 Obteniendo información de actividad %s...", schedule.ActivityID)
	act, err := s.activitiesClient.GetActivityByID(schedule.ActivityID)
	if err != nil {
		log.Printf("❌ Error al obtener actividad: %v", err)
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}
	activity = act
	log.Printf("✅ Actividad obtenida: %s (%s)", activity.Name, activity.Category)

	// ============================================
	// Crear el booking con información desnormalizada
	// ============================================

	booking := &domain.Booking{
		UserID:           req.UserID,
		ScheduleID:       req.ScheduleID,
		ActivityName:     activity.Name,
		ActivityCategory: activity.Category,
		Instructor:       schedule.Instructor,
		DayOfWeek:        schedule.DayOfWeek,
		StartTime:        schedule.StartTime,
		EndTime:          schedule.EndTime,
		Location:         schedule.Location,
		Price:            activity.Price,
		Status:           "confirmed",
	}

	log.Printf("💾 Creando booking en la base de datos...")
	err = s.bookingRepo.Create(booking)
	if err != nil {
		log.Printf("❌ Error al crear booking: %v", err)
		return nil, err
	}

	log.Printf("✅ Booking creado con ID: %s", booking.ID.Hex())

	// ============================================
	// Incrementar current_bookings en el schedule
	// ============================================

	log.Printf("📈 Incrementando contador de reservas en schedule %s...", req.ScheduleID)
	err = s.activitiesClient.UpdateScheduleBookings(req.ScheduleID, true, token)
	if err != nil {
		log.Printf("⚠️  Warning: No se pudo actualizar contador de reservas: %v", err)
		// No fallar la operación, el booking ya fue creado
	} else {
		log.Printf("✅ Contador de reservas actualizado")
	}

	log.Printf("🎉 Reserva creada exitosamente: %s", booking.ID.Hex())
	response := booking.ToBookingResponse()
	return &response, nil
}

/*
Punto clave: Concurrencia implementada ⭐
En el método CreateBooking, implementamos 3 goroutines que se ejecutan en paralelo:

Goroutine 1: Valida que el usuario existe (llamada HTTP a users-api)
Goroutine 2: Valida que el schedule existe y tiene cupos disponibles (llamada HTTP a activities-api)
Goroutine 3: Valida que no hay reserva duplicada (query a MongoDB)

Usamos:
✅ WaitGroup: Para esperar que terminen todas las goroutines
✅ Channel: Para comunicar errores entre goroutines
✅ Sync: Para sincronizar la ejecución

Esto mejora el rendimiento al ejecutar las validaciones en paralelo en lugar de secuencialmente.
*/

// GetBookingByID obtiene una reserva por su ID
func (s *BookingServiceImpl) GetBookingByID(id string) (*domain.BookingResponse, error) {
	booking, err := s.bookingRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := booking.ToBookingResponse()
	return &response, nil
}

// GetBookingsByUser obtiene todas las reservas de un usuario
func (s *BookingServiceImpl) GetBookingsByUser(userID uint) ([]domain.BookingResponse, error) {
	bookings, err := s.bookingRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]domain.BookingResponse, len(bookings))
	for i, booking := range bookings {
		responses[i] = booking.ToBookingResponse()
	}

	return responses, nil
}

// CancelBooking cancela una reserva
func (s *BookingServiceImpl) CancelBooking(id string, userID uint, userRole string, token string) error {
	// Obtener la reserva
	booking, err := s.bookingRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Verificar permisos: solo el dueño o un admin pueden cancelar
	if booking.UserID != userID && userRole != "admin" {
		return domain.ErrUnauthorized
	}

	// Verificar que la reserva esté confirmada
	if booking.Status != "confirmed" {
		return fmt.Errorf("booking is already cancelled or completed")
	}

	// Verificar que la clase no haya pasado (opcional, pero buena práctica)
	// Por ahora solo cambiamos el status

	// Cancelar la reserva (soft delete)
	err = s.bookingRepo.Delete(id)
	if err != nil {
		return err
	}

	// Decrementar current_bookings en el schedule
	err = s.activitiesClient.UpdateScheduleBookings(booking.ScheduleID, false, token)
	if err != nil {
		log.Printf("Warning: failed to decrement schedule bookings: %v", err)
		// No fallar la operación
	}

	return nil
}
