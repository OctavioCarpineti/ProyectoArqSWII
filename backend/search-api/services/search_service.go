package services

import "search-api/domain"

// SearchService define la lógica de negocio para búsqueda
type SearchService interface {
	// Search realiza una búsqueda con caché multinivel
	Search(req domain.SearchRequest) (*domain.SearchResponse, error)

	// GetScheduleByID obtiene un schedule específico por ID
	GetScheduleByID(scheduleID string) (*domain.ScheduleSearch, error)

	// Health verifica el estado de SolR y caché
	Health() map[string]string
}
