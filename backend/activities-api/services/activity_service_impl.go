package services

import (
	"activities-api/clients"
	"activities-api/domain"
	"activities-api/messaging"
	"activities-api/repositories"
	"fmt"
)

// ActivityServiceImpl implementa ActivityService
type ActivityServiceImpl struct {
	activityRepo repositories.ActivityRepository
	scheduleRepo repositories.ScheduleRepository
	userClient   *clients.UserClient
	publisher    *messaging.RabbitMQPublisher
}

// NewActivityService crea una nueva instancia del servicio
func NewActivityService(
	activityRepo repositories.ActivityRepository,
	scheduleRepo repositories.ScheduleRepository,
	userClient *clients.UserClient,
	publisher *messaging.RabbitMQPublisher,
) ActivityService {
	return &ActivityServiceImpl{
		activityRepo: activityRepo,
		scheduleRepo: scheduleRepo,
		userClient:   userClient,
		publisher:    publisher,
	}
}

// CreateActivity crea una nueva actividad
func (s *ActivityServiceImpl) CreateActivity(req domain.CreateActivityRequest) (*domain.ActivityResponse, error) {
	// Validar que el owner existe
	err := s.userClient.ValidateUser(req.OwnerID)
	if err != nil {
		return nil, domain.ErrInvalidOwner
	}

	// Validar categoría
	validCategory := false
	for _, cat := range domain.ValidCategories {
		if cat == req.Category {
			validCategory = true
			break
		}
	}
	if !validCategory {
		return nil, domain.ErrInvalidCategory
	}

	// Crear actividad
	activity := &domain.Activity{
		OwnerID:     req.OwnerID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Duration:    req.Duration,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Status:      "active",
	}

	err = s.activityRepo.Create(activity)
	if err != nil {
		return nil, err
	}

	response := activity.ToActivityResponse()
	return &response, nil
}

// GetActivityByID obtiene una actividad por su ID
func (s *ActivityServiceImpl) GetActivityByID(id string) (*domain.ActivityResponse, error) {
	activity, err := s.activityRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := activity.ToActivityResponse()
	return &response, nil
}

// GetAllActivities obtiene todas las actividades
func (s *ActivityServiceImpl) GetAllActivities() ([]domain.ActivityResponse, error) {
	activities, err := s.activityRepo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := make([]domain.ActivityResponse, len(activities))
	for i, activity := range activities {
		responses[i] = activity.ToActivityResponse()
	}

	return responses, nil
}

// UpdateActivity actualiza una actividad
func (s *ActivityServiceImpl) UpdateActivity(id string, req domain.UpdateActivityRequest, userID uint, userRole string) (*domain.ActivityResponse, error) {
	// Obtener la actividad actual
	activity, err := s.activityRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Verificar permisos (solo owner o admin)
	if activity.OwnerID != userID && userRole != "admin" {
		return nil, domain.ErrUnauthorized
	}

	// Bandera para saber si necesitamos actualizar Solr
	// Solo si cambian campos que están denormalizados en ScheduleSearch
	needsSolrUpdate := false

	// Actualizar campos
	if req.Name != "" {
		activity.Name = req.Name
		needsSolrUpdate = true
	}
	if req.Description != "" {
		activity.Description = req.Description
	}
	if req.Category != "" {
		// Validar categoría
		validCategory := false
		for _, cat := range domain.ValidCategories {
			if cat == req.Category {
				validCategory = true
				break
			}
		}
		if !validCategory {
			return nil, domain.ErrInvalidCategory
		}
		activity.Category = req.Category
		needsSolrUpdate = true
	}
	if req.Duration > 0 {
		activity.Duration = req.Duration
	}
	if req.Price >= 0 {
		activity.Price = req.Price
		needsSolrUpdate = true
	}
	if req.ImageURL != "" {
		activity.ImageURL = req.ImageURL
	}
	if req.Status != "" {
		activity.Status = req.Status
	}

	err = s.activityRepo.Update(activity)
	if err != nil {
		return nil, err
	}

	// Si cambiaron campos denormalizados, actualizar todos los schedules en Solr
	if needsSolrUpdate && s.publisher != nil {
		// Obtener todos los schedules de esta actividad
		schedules, err := s.scheduleRepo.GetByActivityID(id)
		if err != nil {
			// Log pero no fallar la operación principal
			fmt.Printf("Warning: failed to get schedules for Solr update: %v\n", err)
		} else {
			// Publicar evento UPDATE para cada schedule
			for _, schedule := range schedules {
				event := domain.NewScheduleEvent("UPDATE", schedule.ID.Hex(), id)
				err = s.publisher.PublishScheduleEvent(event)
				if err != nil {
					fmt.Printf("Warning: failed to publish schedule update event: %v\n", err)
				}
			}
		}
	}

	response := activity.ToActivityResponse()
	return &response, nil
}

// DeleteActivity elimina una actividad
func (s *ActivityServiceImpl) DeleteActivity(id string, userID uint, userRole string) error {
	// Obtener la actividad
	activity, err := s.activityRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Verificar permisos
	if activity.OwnerID != userID && userRole != "admin" {
		return domain.ErrUnauthorized
	}

	// Eliminar todos los schedules asociados
	schedules, err := s.scheduleRepo.GetByActivityID(id)
	if err != nil {
		return err
	}

	for _, schedule := range schedules {
		scheduleID := schedule.ID.Hex()

		// Eliminar el schedule de MongoDB
		err = s.scheduleRepo.Delete(scheduleID)
		if err != nil {
			return fmt.Errorf("failed to delete schedule: %w", err)
		}

		// Publicar evento DELETE a RabbitMQ para sincronizar con Solr
		if s.publisher != nil {
			event := domain.NewScheduleEvent("DELETE", scheduleID, id)
			err = s.publisher.PublishScheduleEvent(event)
			if err != nil {
				// Log error pero no fallar la operación
				fmt.Printf("Warning: failed to publish schedule delete event: %v\n", err)
			}
		}
	}

	// Eliminar la actividad
	return s.activityRepo.Delete(id)
}
