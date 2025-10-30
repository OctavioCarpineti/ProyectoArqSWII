package services

import (
	"log"
	"search-api/domain"
	"search-api/repositories"
)

// SearchServiceImpl implementa SearchService
type SearchServiceImpl struct {
	solrRepo  repositories.SolrRepository
	cacheRepo repositories.CacheRepository
}

// NewSearchService crea una nueva instancia del servicio
func NewSearchService(
	solrRepo repositories.SolrRepository,
	cacheRepo repositories.CacheRepository,
) SearchService {
	return &SearchServiceImpl{
		solrRepo:  solrRepo,
		cacheRepo: cacheRepo,
	}
}

// Search realiza una búsqueda con estrategia de caché multinivel
func (s *SearchServiceImpl) Search(req domain.SearchRequest) (*domain.SearchResponse, error) {
	log.Printf("🔍 Búsqueda solicitada: q=%s, category=%s, day=%s, page=%d",
		req.Query, req.Category, req.DayOfWeek, req.GetPage())

	// ============================================
	// ESTRATEGIA DE CACHÉ MULTINIVEL
	// ============================================

	// 1. Generar cache key basado en los parámetros de búsqueda
	cacheKey := repositories.GenerateCacheKey(req)
	log.Printf("🔑 Cache key: %s", cacheKey)

	// 2. NIVEL 1: Buscar en CCache (caché local)
	// 3. NIVEL 2: Buscar en Memcached (caché distribuida)
	if cachedResult, found := s.cacheRepo.Get(cacheKey); found {
		log.Printf("✅ Resultado obtenido de caché")
		return cachedResult, nil
	}

	// 4. NIVEL 3: Buscar en SolR (fuente de verdad)
	log.Printf("🔍 Buscando en SolR...")
	results, total, err := s.solrRepo.Search(req)
	if err != nil {
		log.Printf("❌ Error al buscar en SolR: %v", err)
		return nil, err
	}

	log.Printf("✅ SolR devolvió %d resultados (total: %d)", len(results), total)

	// 5. Crear respuesta con paginación
	response := domain.NewSearchResponse(results, total, req.GetPage(), req.GetSize())

	// 6. Guardar en caché (L1 y L2)
	if err := s.cacheRepo.Set(cacheKey, response, 600); err != nil {
		log.Printf("⚠️  Warning: No se pudo guardar en caché: %v", err)
		// No fallar la operación por error de caché
	}

	return response, nil
}

// GetScheduleByID obtiene un schedule específico por ID
func (s *SearchServiceImpl) GetScheduleByID(scheduleID string) (*domain.ScheduleSearch, error) {
	log.Printf("🔍 Obteniendo schedule: %s", scheduleID)

	schedule, err := s.solrRepo.GetByID(scheduleID)
	if err != nil {
		log.Printf("❌ Error al obtener schedule: %v", err)
		return nil, err
	}

	log.Printf("✅ Schedule obtenido: %s - %s", schedule.ActivityName, schedule.DayOfWeek)
	return schedule, nil
}

// Health verifica el estado de SolR y caché
func (s *SearchServiceImpl) Health() map[string]string {
	health := make(map[string]string)

	// Verificar SolR
	if err := s.solrRepo.Ping(); err != nil {
		health["solr"] = "unhealthy"
		log.Printf("❌ SolR health check failed: %v", err)
	} else {
		health["solr"] = "healthy"
		log.Printf("✅ SolR health check passed")
	}

	// Caché siempre está "healthy" porque es opcional
	health["cache"] = "healthy"

	return health
}
