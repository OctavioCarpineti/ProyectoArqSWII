package repositories

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"search-api/domain"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/karlseguin/ccache/v2"
)

// CacheRepositoryImpl implementa caché multinivel con CCache (L1) y Memcached (L2)
type CacheRepositoryImpl struct {
	localCache     *ccache.Cache    // L1: Caché local en memoria
	memcached      *memcache.Client // L2: Caché distribuida
	localTTL       time.Duration    // TTL para L1 (5 minutos)
	distributedTTL int32            // TTL para L2 (10 minutos)
}

// NewCacheRepository crea una nueva instancia del repositorio de caché
func NewCacheRepository(memcachedHost string) CacheRepository {
	// Configurar CCache (L1 - local)
	localCache := ccache.New(ccache.Configure().MaxSize(1000))

	// Configurar Memcached (L2 - distribuida)
	mc := memcache.New(memcachedHost)
	mc.Timeout = 1 * time.Second
	mc.MaxIdleConns = 10

	return &CacheRepositoryImpl{
		localCache:     localCache,
		memcached:      mc,
		localTTL:       5 * time.Minute, // L1: 5 minutos
		distributedTTL: 600,             // L2: 10 minutos (en segundos)
	}
}

// Get obtiene datos del caché (estrategia multinivel)
func (r *CacheRepositoryImpl) Get(key string) (*domain.SearchResponse, bool) {
	cacheKey := r.generateKey(key)

	// NIVEL 1: Buscar en caché local (CCache)
	item := r.localCache.Get(cacheKey)
	if item != nil && !item.Expired() {
		if response, ok := item.Value().(*domain.SearchResponse); ok {
			log.Printf("🎯 Cache L1 HIT: %s", key)
			return response, true
		}
	}

	log.Printf("❌ Cache L1 MISS: %s", key)

	// NIVEL 2: Buscar en caché distribuida (Memcached)
	mcItem, err := r.memcached.Get(cacheKey)
	if err == nil && mcItem != nil {
		var response domain.SearchResponse
		if err := json.Unmarshal(mcItem.Value, &response); err == nil {
			log.Printf("🎯 Cache L2 HIT: %s (promoting to L1)", key)

			// Promover a L1
			r.localCache.Set(cacheKey, &response, r.localTTL)

			return &response, true
		}
	}

	log.Printf("❌ Cache L2 MISS: %s", key)
	return nil, false
}

// Set guarda datos en ambos niveles de caché
func (r *CacheRepositoryImpl) Set(key string, value *domain.SearchResponse, ttl int) error {
	cacheKey := r.generateKey(key)

	// Guardar en L1 (CCache)
	r.localCache.Set(cacheKey, value, r.localTTL)
	log.Printf("💾 Guardado en Cache L1: %s", key)

	// Guardar en L2 (Memcached)
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	item := &memcache.Item{
		Key:        cacheKey,
		Value:      data,
		Expiration: r.distributedTTL,
	}

	if err := r.memcached.Set(item); err != nil {
		log.Printf("⚠️  Warning: Failed to set L2 cache: %v", err)
		// No retornar error, L1 ya está guardado
	} else {
		log.Printf("💾 Guardado en Cache L2: %s", key)
	}

	return nil
}

// Delete elimina una entrada del caché en ambos niveles
func (r *CacheRepositoryImpl) Delete(key string) error {
	cacheKey := r.generateKey(key)

	// Eliminar de L1
	r.localCache.Delete(cacheKey)
	log.Printf("🗑️  Eliminado de Cache L1: %s", key)

	// Eliminar de L2
	if err := r.memcached.Delete(cacheKey); err != nil && err != memcache.ErrCacheMiss {
		log.Printf("⚠️  Warning: Failed to delete from L2 cache: %v", err)
	} else {
		log.Printf("🗑️  Eliminado de Cache L2: %s", key)
	}

	return nil
}

// InvalidatePattern invalida todas las keys que coincidan con un patrón
// Nota: Esta implementación es simplificada. En producción usarías Redis con SCAN
func (r *CacheRepositoryImpl) InvalidatePattern(pattern string) error {
	// Para CCache (L1), limpiamos todo el caché local
	log.Printf("🗑️  InvalidatePattern called with: %s (clearing L1 and L2 cache)", pattern)
	r.localCache.Clear()

	// Para Memcached (L2), hacemos flush completo ya que no soporta pattern invalidation
	// Esto asegura que los datos actualizados se obtengan de Solr
	if err := r.memcached.FlushAll(); err != nil {
		log.Printf("⚠️  Warning: Failed to flush Memcached: %v", err)
	} else {
		log.Printf("✅ Memcached (L2) flushed successfully")
	}

	return nil
}

// Clear limpia todo el caché
func (r *CacheRepositoryImpl) Clear() error {
	// Limpiar L1
	r.localCache.Clear()
	log.Println("🗑️  Cache L1 completamente limpiado")

	// Limpiar L2 (flush all)
	if err := r.memcached.FlushAll(); err != nil {
		log.Printf("⚠️  Warning: Failed to flush L2 cache: %v", err)
	} else {
		log.Println("🗑️  Cache L2 completamente limpiado")
	}

	return nil
}

// generateKey genera una key de caché única usando MD5
func (r *CacheRepositoryImpl) generateKey(key string) string {
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("search:%x", hash)
}

// GenerateCacheKey genera una key basada en los parámetros de búsqueda
func GenerateCacheKey(req domain.SearchRequest) string {
	return fmt.Sprintf("search:q=%s:cat=%s:day=%s:inst=%s:loc=%s:tf=%s:tt=%s:avail=%t:sort=%s:%s:page=%d:size=%d",
		req.Query,
		req.Category,
		req.DayOfWeek,
		req.Instructor,
		req.Location,
		req.TimeFrom,
		req.TimeTo,
		req.Available,
		req.GetSortField(),
		req.GetSortOrder(),
		req.GetPage(),
		req.GetSize(),
	)
}
