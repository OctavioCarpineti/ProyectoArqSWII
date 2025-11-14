package services

import (
	"errors"
	"search-api/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks

type MockSolrRepository struct {
	mock.Mock
}

func (m *MockSolrRepository) IndexSchedule(schedule *domain.ScheduleSearch) error {
	args := m.Called(schedule)
	return args.Error(0)
}

func (m *MockSolrRepository) Search(params domain.SearchParams) (*domain.SearchResponse, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SearchResponse), args.Error(1)
}

func (m *MockSolrRepository) GetByID(id string) (*domain.ScheduleSearch, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ScheduleSearch), args.Error(1)
}

func (m *MockSolrRepository) DeleteByID(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Get(key string) (*domain.SearchResponse, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SearchResponse), args.Error(1)
}

func (m *MockCacheRepository) Set(key string, value *domain.SearchResponse, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

// Tests para Search

func TestSearch_CacheHit(t *testing.T) {
	// Arrange
	mockSolrRepo := new(MockSolrRepository)
	mockCacheRepo := new(MockCacheRepository)
	service := NewSearchService(mockSolrRepo, mockCacheRepo)

	params := domain.SearchParams{
		Query:    "yoga",
		Page:     1,
		Size:     10,
		SortBy:   "start_time",
		SortOrder: "asc",
	}

	cachedResponse := &domain.SearchResponse{
		Results: []domain.ScheduleSearch{
			{
				ID:           "123",
				ActivityName: "Yoga Intermedio",
				Instructor:   "María",
			},
		},
		TotalResults: 1,
		Page:         1,
		Size:         10,
		TotalPages:   1,
	}

	// Cache HIT - no debe llamar a Solr
	mockCacheRepo.On("Get", mock.Anything).Return(cachedResponse, nil)

	// Act
	response, err := service.Search(params)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, response.TotalResults)
	assert.Equal(t, "Yoga Intermedio", response.Results[0].ActivityName)

	mockCacheRepo.AssertExpectations(t)
	mockSolrRepo.AssertNotCalled(t, "Search")
}

func TestSearch_CacheMiss_ThenSet(t *testing.T) {
	// Arrange
	mockSolrRepo := new(MockSolrRepository)
	mockCacheRepo := new(MockCacheRepository)
	service := NewSearchService(mockSolrRepo, mockCacheRepo)

	params := domain.SearchParams{
		Query:    "yoga",
		Page:     1,
		Size:     10,
		SortBy:   "start_time",
		SortOrder: "asc",
	}

	solrResponse := &domain.SearchResponse{
		Results: []domain.ScheduleSearch{
			{
				ID:           "123",
				ActivityName: "Yoga Intermedio",
				Instructor:   "María",
			},
		},
		TotalResults: 1,
		Page:         1,
		Size:         10,
		TotalPages:   1,
	}

	// Cache MISS
	mockCacheRepo.On("Get", mock.Anything).Return(nil, errors.New("cache miss"))

	// Debe consultar Solr
	mockSolrRepo.On("Search", params).Return(solrResponse, nil)

	// Debe guardar en cache
	mockCacheRepo.On("Set", mock.Anything, solrResponse, mock.AnythingOfType("time.Duration")).Return(nil)

	// Act
	response, err := service.Search(params)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, response.TotalResults)
	assert.Equal(t, "Yoga Intermedio", response.Results[0].ActivityName)

	mockCacheRepo.AssertExpectations(t)
	mockSolrRepo.AssertExpectations(t)
}

func TestSearch_SolrError(t *testing.T) {
	// Arrange
	mockSolrRepo := new(MockSolrRepository)
	mockCacheRepo := new(MockCacheRepository)
	service := NewSearchService(mockSolrRepo, mockCacheRepo)

	params := domain.SearchParams{
		Query: "yoga",
		Page:  1,
		Size:  10,
	}

	// Cache MISS
	mockCacheRepo.On("Get", mock.Anything).Return(nil, errors.New("cache miss"))

	// Solr error
	mockSolrRepo.On("Search", params).Return(nil, errors.New("solr connection error"))

	// Act
	response, err := service.Search(params)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)

	mockCacheRepo.AssertExpectations(t)
	mockSolrRepo.AssertExpectations(t)
}

// Tests para GetByID

func TestGetByID_Success(t *testing.T) {
	// Arrange
	mockSolrRepo := new(MockSolrRepository)
	mockCacheRepo := new(MockCacheRepository)
	service := NewSearchService(mockSolrRepo, mockCacheRepo)

	schedule := &domain.ScheduleSearch{
		ID:           "123",
		ScheduleID:   "123",
		ActivityID:   "456",
		ActivityName: "Yoga Intermedio",
		Instructor:   "María González",
		DayOfWeek:    "monday",
		StartTime:    "18:00",
		EndTime:      "19:00",
		Location:     "Sala 1",
		Status:       "active",
	}

	mockSolrRepo.On("GetByID", "123").Return(schedule, nil)

	// Act
	response, err := service.GetByID("123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "123", response.ID)
	assert.Equal(t, "Yoga Intermedio", response.ActivityName)
	assert.Equal(t, "María González", response.Instructor)

	mockSolrRepo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	// Arrange
	mockSolrRepo := new(MockSolrRepository)
	mockCacheRepo := new(MockCacheRepository)
	service := NewSearchService(mockSolrRepo, mockCacheRepo)

	mockSolrRepo.On("GetByID", "999").Return(nil, errors.New("not found"))

	// Act
	response, err := service.GetByID("999")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)

	mockSolrRepo.AssertExpectations(t)
}

// Tests para IndexSchedule

func TestIndexSchedule_Success(t *testing.T) {
	// Arrange
	mockSolrRepo := new(MockSolrRepository)
	mockCacheRepo := new(MockCacheRepository)
	service := NewSearchService(mockSolrRepo, mockCacheRepo)

	schedule := &domain.ScheduleSearch{
		ID:           "123",
		ScheduleID:   "123",
		ActivityID:   "456",
		ActivityName: "Yoga Intermedio",
		Instructor:   "María González",
		DayOfWeek:    "monday",
		StartTime:    "18:00",
		EndTime:      "19:00",
		Location:     "Sala 1",
		MaxCapacity:  20,
		Status:       "active",
	}

	mockSolrRepo.On("IndexSchedule", schedule).Return(nil)

	// Act
	err := service.IndexSchedule(schedule)

	// Assert
	assert.NoError(t, err)
	mockSolrRepo.AssertExpectations(t)
}

func TestIndexSchedule_Error(t *testing.T) {
	// Arrange
	mockSolrRepo := new(MockSolrRepository)
	mockCacheRepo := new(MockCacheRepository)
	service := NewSearchService(mockSolrRepo, mockCacheRepo)

	schedule := &domain.ScheduleSearch{
		ID:           "123",
		ScheduleID:   "123",
		ActivityID:   "456",
		ActivityName: "Yoga Intermedio",
	}

	mockSolrRepo.On("IndexSchedule", schedule).Return(errors.New("solr indexing error"))

	// Act
	err := service.IndexSchedule(schedule)

	// Assert
	assert.Error(t, err)
	mockSolrRepo.AssertExpectations(t)
}

// Tests para DeleteByID

func TestDeleteByID_Success(t *testing.T) {
	// Arrange
	mockSolrRepo := new(MockSolrRepository)
	mockCacheRepo := new(MockCacheRepository)
	service := NewSearchService(mockSolrRepo, mockCacheRepo)

	mockSolrRepo.On("DeleteByID", "123").Return(nil)

	// Act
	err := service.DeleteByID("123")

	// Assert
	assert.NoError(t, err)
	mockSolrRepo.AssertExpectations(t)
}

// Tests de integración de búsquedas con filtros

func TestSearch_WithMultipleFilters(t *testing.T) {
	// Arrange
	mockSolrRepo := new(MockSolrRepository)
	mockCacheRepo := new(MockCacheRepository)
	service := NewSearchService(mockSolrRepo, mockCacheRepo)

	params := domain.SearchParams{
		Query:      "yoga",
		Category:   "Yoga",
		DayOfWeek:  "monday",
		Instructor: "María",
		Available:  true,
		Page:       1,
		Size:       10,
		SortBy:     "start_time",
		SortOrder:  "asc",
	}

	solrResponse := &domain.SearchResponse{
		Results: []domain.ScheduleSearch{
			{
				ID:             "123",
				ActivityName:   "Yoga Intermedio",
				ActivityCategory: "Yoga",
				Instructor:     "María González",
				DayOfWeek:      "monday",
				AvailableSpots: 15,
			},
		},
		TotalResults: 1,
		Page:         1,
		Size:         10,
		TotalPages:   1,
	}

	mockCacheRepo.On("Get", mock.Anything).Return(nil, errors.New("cache miss"))
	mockSolrRepo.On("Search", params).Return(solrResponse, nil)
	mockCacheRepo.On("Set", mock.Anything, solrResponse, mock.AnythingOfType("time.Duration")).Return(nil)

	// Act
	response, err := service.Search(params)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, response.TotalResults)
	assert.Equal(t, "Yoga", response.Results[0].ActivityCategory)
	assert.Equal(t, "monday", response.Results[0].DayOfWeek)
	assert.Greater(t, response.Results[0].AvailableSpots, 0)

	mockSolrRepo.AssertExpectations(t)
	mockCacheRepo.AssertExpectations(t)
}

func TestSearch_Pagination(t *testing.T) {
	// Arrange
	mockSolrRepo := new(MockSolrRepository)
	mockCacheRepo := new(MockCacheRepository)
	service := NewSearchService(mockSolrRepo, mockCacheRepo)

	params := domain.SearchParams{
		Page:  2,
		Size:  5,
	}

	solrResponse := &domain.SearchResponse{
		Results:      []domain.ScheduleSearch{},
		TotalResults: 12,
		Page:         2,
		Size:         5,
		TotalPages:   3,
	}

	mockCacheRepo.On("Get", mock.Anything).Return(nil, errors.New("cache miss"))
	mockSolrRepo.On("Search", params).Return(solrResponse, nil)
	mockCacheRepo.On("Set", mock.Anything, solrResponse, mock.AnythingOfType("time.Duration")).Return(nil)

	// Act
	response, err := service.Search(params)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 2, response.Page)
	assert.Equal(t, 5, response.Size)
	assert.Equal(t, 3, response.TotalPages)
	assert.Equal(t, 12, response.TotalResults)

	mockSolrRepo.AssertExpectations(t)
}
