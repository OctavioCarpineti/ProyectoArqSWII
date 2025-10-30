package services

import (
	"activities-api/clients"
	"activities-api/domain"
	"activities-api/messaging"
	"activities-api/repositories"
	"fmt"
	"regexp"
	"sync"
)

// ScheduleServiceImpl implementa ScheduleService
type ScheduleServiceImpl struct {
	scheduleRepo repositories.ScheduleRepository
	activityRepo repositories.ActivityRepository
	userClient   *clients.UserClient
	publisher    *messaging.RabbitMQPublisher
}

// NewScheduleService crea una nueva instancia del servicio
func NewScheduleService(
	scheduleRepo repositories.ScheduleRepository,
	activityRepo repositories.ActivityRepository,
	userClient *clients.UserClient,
	publisher *messaging.RabbitMQPublisher,
) ScheduleService {
	return &ScheduleServiceImpl{
		scheduleRepo: scheduleRepo,
		activityRepo: activityRepo,
		userClient:   userClient,
		publisher:    publisher,
	}
}

// CreateSchedule crea un nuevo horario con validaciones concurrentes
func (s *ScheduleServiceImpl) CreateSchedule(activityID string, req domain.CreateScheduleRequest) (*domain.ScheduleResponse, error) {
	// ============================================
	// CONCURRENCIA: Validaciones en paralelo
	// ============================================

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Goroutine 1: Validar que la actividad existe
	wg.Add(1)
	go func() {
		defer wg.Done()
		exists, err := s.activityRepo.Exists(activityID)
		if err != nil {
			errChan <- fmt.Errorf("error checking activity: %w", err)
			return
		}
		if !exists {
			errChan <- domain.ErrActivityNotFound
		}
	}()

	// Goroutine 2: Validar que no hay conflicto de horario/sala
	wg.Add(1)
	go func() {
		defer wg.Done()
		conflict, err := s.scheduleRepo.CheckConflict(
			req.DayOfWeek,
			req.StartTime,
			req.EndTime,
			req.Location,
			"", // No excluir ningún ID (es creación)
		)
		if err != nil {
			errChan <- fmt.Errorf("error checking conflict: %w", err)
			return
		}
		if conflict {
			errChan <- domain.ErrScheduleConflict
		}
	}()

	// Goroutine 3: Validar formato de tiempo
	wg.Add(1)
	go func() {
		defer wg.Done()
		timeRegex := regexp.MustCompile(`^([0-1][0-9]|2[0-3]):[0-5][0-9]$`)
		if !timeRegex.MatchString(req.StartTime) || !timeRegex.MatchString(req.EndTime) {
			errChan <- domain.ErrInvalidTimeFormat
		}
	}()

	// Esperar a que todas las goroutines terminen
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Verificar si hubo errores
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	// ============================================
	// Crear el schedule si todas las validaciones pasaron
	// ============================================

	schedule := &domain.Schedule{
		ActivityID:  activityID,
		Instructor:  req.Instructor,
		DayOfWeek:   req.DayOfWeek,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Location:    req.Location,
		MaxCapacity: req.MaxCapacity,
		Status:      "active",
	}

	err := s.scheduleRepo.Create(schedule)
	if err != nil {
		return nil, err
	}

	// Publicar evento a RabbitMQ
	event := domain.NewScheduleEvent("CREATE", schedule.ID.Hex(), activityID)
	err = s.publisher.PublishScheduleEvent(event)
	if err != nil {
		// Log error pero no fallar la operación
		fmt.Printf("Warning: failed to publish event: %v\n", err)
	}

	response := schedule.ToScheduleResponse()
	return &response, nil
}

/*
Punto clave: Concurrencia implementada ⭐
En el método CreateSchedule, implementamos 3 goroutines que se ejecutan en paralelo:

Goroutine 1: Valida que la actividad existe
Goroutine 2: Valida que no hay conflicto de horario/sala
Goroutine 3: Valida el formato de tiempo

Usamos:

✅ WaitGroup: Para esperar que terminen todas las goroutines
✅ Channel: Para comunicar errores entre goroutines
✅ Sync: Para sincronizar la ejecución
*/

// GetScheduleByID obtiene un horario por su ID
func (s *ScheduleServiceImpl) GetScheduleByID(id string) (*domain.ScheduleResponse, error) {
	schedule, err := s.scheduleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := schedule.ToScheduleResponse()
	return &response, nil
}

// GetSchedulesByActivityID obtiene todos los horarios de una actividad
func (s *ScheduleServiceImpl) GetSchedulesByActivityID(activityID string) ([]domain.ScheduleResponse, error) {
	schedules, err := s.scheduleRepo.GetByActivityID(activityID)
	if err != nil {
		return nil, err
	}

	responses := make([]domain.ScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		responses[i] = schedule.ToScheduleResponse()
	}

	return responses, nil
}

// GetAllSchedules obtiene todos los horarios
func (s *ScheduleServiceImpl) GetAllSchedules() ([]domain.ScheduleResponse, error) {
	schedules, err := s.scheduleRepo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := make([]domain.ScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		responses[i] = schedule.ToScheduleResponse()
	}

	return responses, nil
}

// UpdateSchedule actualiza un horario
func (s *ScheduleServiceImpl) UpdateSchedule(id string, req domain.UpdateScheduleRequest, userID uint, userRole string) (*domain.ScheduleResponse, error) {
	// Obtener el schedule actual
	schedule, err := s.scheduleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Obtener la actividad para verificar permisos
	activity, err := s.activityRepo.GetByID(schedule.ActivityID)
	if err != nil {
		return nil, err
	}

	// Verificar permisos
	if activity.OwnerID != userID && userRole != "admin" {
		return nil, domain.ErrUnauthorized
	}

	// Actualizar campos
	if req.Instructor != "" {
		schedule.Instructor = req.Instructor
	}
	if req.DayOfWeek != "" {
		schedule.DayOfWeek = req.DayOfWeek
	}
	if req.StartTime != "" {
		schedule.StartTime = req.StartTime
	}
	if req.EndTime != "" {
		schedule.EndTime = req.EndTime
	}
	if req.Location != "" {
		schedule.Location = req.Location
	}
	if req.MaxCapacity > 0 {
		schedule.MaxCapacity = req.MaxCapacity
	}
	if req.Status != "" {
		schedule.Status = req.Status
	}

	// Verificar conflictos si se cambió horario/ubicación
	if req.DayOfWeek != "" || req.StartTime != "" || req.EndTime != "" || req.Location != "" {
		conflict, err := s.scheduleRepo.CheckConflict(
			schedule.DayOfWeek,
			schedule.StartTime,
			schedule.EndTime,
			schedule.Location,
			id, // Excluir el schedule actual
		)
		if err != nil {
			return nil, err
		}
		if conflict {
			return nil, domain.ErrScheduleConflict
		}
	}

	err = s.scheduleRepo.Update(schedule)
	if err != nil {
		return nil, err
	}

	// Publicar evento UPDATE
	event := domain.NewScheduleEvent("UPDATE", id, schedule.ActivityID)
	s.publisher.PublishScheduleEvent(event)

	response := schedule.ToScheduleResponse()
	return &response, nil
}

// DeleteSchedule elimina un horario
func (s *ScheduleServiceImpl) DeleteSchedule(id string, userID uint, userRole string) error {
	// Obtener el schedule
	schedule, err := s.scheduleRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Obtener la actividad para verificar permisos
	activity, err := s.activityRepo.GetByID(schedule.ActivityID)
	if err != nil {
		return err
	}

	// Verificar permisos
	if activity.OwnerID != userID && userRole != "admin" {
		return domain.ErrUnauthorized
	}

	// Eliminar el schedule
	err = s.scheduleRepo.Delete(id)
	if err != nil {
		return err
	}

	// Publicar evento DELETE
	event := domain.NewScheduleEvent("DELETE", id, schedule.ActivityID)
	s.publisher.PublishScheduleEvent(event)

	return nil
}

// UpdateCurrentBookings actualiza el contador de reservas
func (s *ScheduleServiceImpl) UpdateCurrentBookings(scheduleID string, increment bool) error {
	if increment {
		return s.scheduleRepo.IncrementBookings(scheduleID)
	}
	return s.scheduleRepo.DecrementBookings(scheduleID)
}
