package domain

// SearchRequest representa los parámetros de búsqueda
type SearchRequest struct {
	// Búsqueda de texto libre
	Query string `form:"q"` // Busca en: activity_name, instructor, category

	// Filtros específicos
	ActivityID string `form:"activity_id"` // Filtrar por actividad específica
	Category   string `form:"category"`    // Yoga, Spinning, etc.
	DayOfWeek  string `form:"day_of_week"` // monday, tuesday, etc.
	Instructor string `form:"instructor"`  // Nombre del instructor
	Location   string `form:"location"`    // Ubicación de la clase

	// Rango de horarios
	TimeFrom string `form:"time_from"` // HH:MM - Horarios desde
	TimeTo   string `form:"time_to"`   // HH:MM - Horarios hasta

	// Filtros de disponibilidad
	Available bool `form:"available"` // Solo clases con cupos disponibles

	// Ordenamiento
	SortBy string `form:"sort"`  // start_time, price, available_spots, activity_name
	Order  string `form:"order"` // asc, desc (default: asc)

	// Paginación
	Page int `form:"page"` // Número de página (default: 1)
	Size int `form:"size"` // Tamaño de página (default: 10)
}

// GetPage retorna el número de página, asegurando que sea >= 1
func (r *SearchRequest) GetPage() int {
	if r.Page < 1 {
		return 1
	}
	return r.Page
}

// GetSize retorna el tamaño de página, limitado entre 1 y 100
func (r *SearchRequest) GetSize() int {
	if r.Size < 1 {
		return 10
	}
	if r.Size > 100 {
		return 100
	}
	return r.Size
}

// GetOffset calcula el offset para SolR
func (r *SearchRequest) GetOffset() int {
	return (r.GetPage() - 1) * r.GetSize()
}

// GetSortField retorna el campo de ordenamiento validado
func (r *SearchRequest) GetSortField() string {
	validFields := map[string]bool{
		"start_time":      true,
		"price":           true,
		"available_spots": true,
		"activity_name":   true,
		"created_at":      true,
	}

	if validFields[r.SortBy] {
		return r.SortBy
	}
	return "start_time" // Default
}

// GetSortOrder retorna el orden validado
func (r *SearchRequest) GetSortOrder() string {
	if r.Order == "desc" {
		return "desc"
	}
	return "asc" // Default
}

// SearchResponse representa la respuesta de búsqueda con paginación
type SearchResponse struct {
	Results      []ScheduleSearch `json:"results"`
	TotalResults int              `json:"total_results"`
	Page         int              `json:"page"`
	Size         int              `json:"size"`
	TotalPages   int              `json:"total_pages"`
}

// NewSearchResponse crea una respuesta de búsqueda con paginación
func NewSearchResponse(results []ScheduleSearch, total int, page int, size int) *SearchResponse {
	totalPages := total / size
	if total%size > 0 {
		totalPages++
	}

	return &SearchResponse{
		Results:      results,
		TotalResults: total,
		Page:         page,
		Size:         size,
		TotalPages:   totalPages,
	}
}
