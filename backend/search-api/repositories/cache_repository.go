package repositories

import "search-api/domain"

// CacheRepository define las operaciones de caché multinivel
type CacheRepository interface {
	// Get obtiene datos del caché (primero L1, luego L2)
	Get(key string) (*domain.SearchResponse, bool)

	// Set guarda datos en ambos niveles de caché
	Set(key string, value *domain.SearchResponse, ttl int) error

	// Delete elimina una entrada del caché en ambos niveles
	Delete(key string) error

	// InvalidatePattern invalida todas las keys que coincidan con un patrón
	InvalidatePattern(pattern string) error

	// Clear limpia todo el caché
	Clear() error
}
