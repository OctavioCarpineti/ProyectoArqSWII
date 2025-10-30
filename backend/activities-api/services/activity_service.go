package services

import "activities-api/domain"

// ActivityService define la interface para el servicio de actividades
type ActivityService interface {
	// CreateActivity crea una nueva actividad
	CreateActivity(req domain.CreateActivityRequest) (*domain.ActivityResponse, error)

	// GetActivityByID obtiene una actividad por su ID
	GetActivityByID(id string) (*domain.ActivityResponse, error)

	// GetAllActivities obtiene todas las actividades
	GetAllActivities() ([]domain.ActivityResponse, error)

	// UpdateActivity actualiza una actividad
	UpdateActivity(id string, req domain.UpdateActivityRequest, userID uint, userRole string) (*domain.ActivityResponse, error)

	// DeleteActivity elimina una actividad
	DeleteActivity(id string, userID uint, userRole string) error
}
