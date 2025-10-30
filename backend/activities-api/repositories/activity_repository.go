package repositories

import "activities-api/domain"

// ActivityRepository define la interface para el repositorio de actividades
type ActivityRepository interface {
	// Create crea una nueva actividad
	Create(activity *domain.Activity) error

	// GetByID obtiene una actividad por su ID
	GetByID(id string) (*domain.Activity, error)

	// GetAll obtiene todas las actividades
	GetAll() ([]domain.Activity, error)

	// GetByOwnerID obtiene todas las actividades de un owner
	GetByOwnerID(ownerID uint) ([]domain.Activity, error)

	// Update actualiza una actividad existente
	Update(activity *domain.Activity) error

	// Delete elimina una actividad
	Delete(id string) error

	// Exists verifica si existe una actividad
	Exists(id string) (bool, error)
}
