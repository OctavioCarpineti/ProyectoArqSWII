package controllers

import (
	"log"
	"net/http"
	"search-api/domain"
	"search-api/services"

	"github.com/gin-gonic/gin"
)

// SearchController maneja las peticiones HTTP de búsqueda
type SearchController struct {
	service services.SearchService
}

// NewSearchController crea una nueva instancia del controller
func NewSearchController(service services.SearchService) *SearchController {
	return &SearchController{
		service: service,
	}
}

// Search maneja GET /search
func (c *SearchController) Search(ctx *gin.Context) {
	var req domain.SearchRequest

	// Parsear query parameters
	if err := ctx.ShouldBindQuery(&req); err != nil {
		log.Printf("❌ Error al parsear query params: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	log.Printf("📥 Request de búsqueda recibido: %+v", req)

	// Ejecutar búsqueda
	response, err := c.service.Search(req)
	if err != nil {
		log.Printf("❌ Error en búsqueda: %v", err)

		// Mapear errores a códigos HTTP
		switch err {
		case domain.ErrSolrNotAvailable:
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "Search service temporarily unavailable"})
		case domain.ErrInvalidSearchQuery:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search query"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		}
		return
	}

	log.Printf("✅ Búsqueda exitosa: %d resultados encontrados", len(response.Results))
	ctx.JSON(http.StatusOK, response)
}

// GetScheduleByID maneja GET /search/:id
func (c *SearchController) GetScheduleByID(ctx *gin.Context) {
	scheduleID := ctx.Param("id")

	schedule, err := c.service.GetScheduleByID(scheduleID)
	if err != nil {
		if err == domain.ErrDocumentNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get schedule"})
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

// Health maneja GET /health
func (c *SearchController) Health(ctx *gin.Context) {
	health := c.service.Health()

	// Determinar status code basado en la salud
	statusCode := http.StatusOK
	if health["solr"] != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	ctx.JSON(statusCode, gin.H{
		"status":  "running",
		"service": "search-api",
		"health":  health,
	})
}
