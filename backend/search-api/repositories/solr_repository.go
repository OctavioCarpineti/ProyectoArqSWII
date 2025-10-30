package repositories

import "search-api/domain"

// SolrRepository define las operaciones con SolR
type SolrRepository interface {
	// Index indexa un documento en SolR
	Index(schedule *domain.ScheduleSearch) error

	// Update actualiza un documento en SolR
	Update(schedule *domain.ScheduleSearch) error

	// Delete elimina un documento de SolR por ID
	Delete(scheduleID string) error

	// Search realiza una búsqueda en SolR
	Search(req domain.SearchRequest) ([]domain.ScheduleSearch, int, error)

	// GetByID obtiene un documento por su ID
	GetByID(scheduleID string) (*domain.ScheduleSearch, error)

	// Ping verifica la conexión con SolR
	Ping() error
}
