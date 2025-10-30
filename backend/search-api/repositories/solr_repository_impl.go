package repositories

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"search-api/domain"
	"strings"
	"time"
)

// SolrRepositoryImpl implementa SolrRepository usando HTTP Client
type SolrRepositoryImpl struct {
	baseURL    string
	httpClient *http.Client
}

// NewSolrRepository crea una nueva instancia del repositorio
func NewSolrRepository(solrURL string) SolrRepository {
	return &SolrRepositoryImpl{
		baseURL: solrURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Index indexa un documento en SolR
func (r *SolrRepositoryImpl) Index(schedule *domain.ScheduleSearch) error {
	// Calcular available_spots antes de indexar
	schedule.CalculateAvailableSpots()

	url := fmt.Sprintf("%s/update?commit=true", r.baseURL)

	// Convertir a JSON
	data, err := json.Marshal([]interface{}{schedule})
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	// Enviar a SolR
	req, err := http.NewRequest("POST", url, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("solr returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Update actualiza un documento en SolR (usa el mismo método que Index)
func (r *SolrRepositoryImpl) Update(schedule *domain.ScheduleSearch) error {
	return r.Index(schedule) // SolR hace upsert automáticamente
}

// Delete elimina un documento de SolR
func (r *SolrRepositoryImpl) Delete(scheduleID string) error {
	url := fmt.Sprintf("%s/update?commit=true", r.baseURL)

	deleteQuery := map[string]interface{}{
		"delete": map[string]string{
			"id": scheduleID,
		},
	}

	data, err := json.Marshal(deleteQuery)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("solr returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Search realiza una búsqueda en SolR
func (r *SolrRepositoryImpl) Search(req domain.SearchRequest) ([]domain.ScheduleSearch, int, error) {
	// Construir query de SolR
	params := r.buildSearchQuery(req)

	searchURL := fmt.Sprintf("%s/select?%s", r.baseURL, params.Encode())

	resp, err := r.httpClient.Get(searchURL)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("solr returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parsear respuesta de SolR
	var solrResponse struct {
		Response struct {
			NumFound int                     `json:"numFound"`
			Docs     []domain.ScheduleSearch `json:"docs"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&solrResponse); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return solrResponse.Response.Docs, solrResponse.Response.NumFound, nil
}

// buildSearchQuery construye los parámetros de búsqueda para SolR
func (r *SolrRepositoryImpl) buildSearchQuery(req domain.SearchRequest) url.Values {
	params := url.Values{}

	// Query principal
	if req.Query != "" {
		// Buscar en múltiples campos
		params.Set("q", fmt.Sprintf("activity_name:%s OR instructor:%s OR activity_category:%s",
			req.Query, req.Query, req.Query))
	} else {
		params.Set("q", "*:*") // Buscar todo
	}

	// Filtros (fq - filter query)
	var filters []string

	if req.ActivityID != "" {
		filters = append(filters, fmt.Sprintf("activity_id:%s", req.ActivityID))
	}

	if req.Category != "" {
		filters = append(filters, fmt.Sprintf("activity_category:%s", req.Category))
	}

	if req.DayOfWeek != "" {
		filters = append(filters, fmt.Sprintf("day_of_week:%s", req.DayOfWeek))
	}

	if req.Instructor != "" {
		filters = append(filters, fmt.Sprintf("instructor:*%s*", req.Instructor))
	}

	if req.Location != "" {
		filters = append(filters, fmt.Sprintf("location:*%s*", req.Location))
	}

	if req.TimeFrom != "" {
		filters = append(filters, fmt.Sprintf("start_time:[%s TO *]", req.TimeFrom))
	}

	if req.TimeTo != "" {
		filters = append(filters, fmt.Sprintf("start_time:[* TO %s]", req.TimeTo))
	}

	if req.Available {
		filters = append(filters, "available_spots:[1 TO *]")
		filters = append(filters, "status:active")
	}

	// Agregar todos los filtros
	for _, filter := range filters {
		params.Add("fq", filter)
	}

	// Ordenamiento
	sortField := req.GetSortField()
	sortOrder := req.GetSortOrder()
	params.Set("sort", fmt.Sprintf("%s %s", sortField, sortOrder))

	// Paginación
	params.Set("start", fmt.Sprintf("%d", req.GetOffset()))
	params.Set("rows", fmt.Sprintf("%d", req.GetSize()))

	// Formato de respuesta
	params.Set("wt", "json")

	return params
}

// GetByID obtiene un documento por su ID
func (r *SolrRepositoryImpl) GetByID(scheduleID string) (*domain.ScheduleSearch, error) {
	url := fmt.Sprintf("%s/select?q=id:%s&wt=json", r.baseURL, scheduleID)

	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	defer resp.Body.Close()

	var solrResponse struct {
		Response struct {
			NumFound int                     `json:"numFound"`
			Docs     []domain.ScheduleSearch `json:"docs"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&solrResponse); err != nil {
		return nil, err
	}

	if solrResponse.Response.NumFound == 0 {
		return nil, domain.ErrDocumentNotFound
	}

	return &solrResponse.Response.Docs[0], nil
}

// Ping verifica la conexión con SolR
func (r *SolrRepositoryImpl) Ping() error {
	url := fmt.Sprintf("%s/admin/ping?wt=json", r.baseURL)

	resp, err := r.httpClient.Get(url)
	if err != nil {
		return domain.ErrSolrNotAvailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.ErrSolrNotAvailable
	}

	return nil
}
