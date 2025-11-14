package services

import (
	"activities-api/domain"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mocks

type MockScheduleRepository struct {
	mock.Mock
}

func (m *MockScheduleRepository) Create(schedule *domain.Schedule) error {
	args := m.Called(schedule)
	return args.Error(0)
}

func (m *MockScheduleRepository) GetByID(id string) (*domain.Schedule, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Schedule), args.Error(1)
}

func (m *MockScheduleRepository) GetByActivityID(activityID string) ([]domain.Schedule, error) {
	args := m.Called(activityID)
	return args.Get(0).([]domain.Schedule), args.Error(1)
}

func (m *MockScheduleRepository) GetAll() ([]domain.Schedule, error) {
	args := m.Called()
	return args.Get(0).([]domain.Schedule), args.Error(1)
}

func (m *MockScheduleRepository) Update(schedule *domain.Schedule) error {
	args := m.Called(schedule)
	return args.Error(0)
}

func (m *MockScheduleRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockScheduleRepository) CheckConflict(dayOfWeek, startTime, endTime, location, excludeID string) (bool, error) {
	args := m.Called(dayOfWeek, startTime, endTime, location, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockScheduleRepository) IncrementBookings(scheduleID string) error {
	args := m.Called(scheduleID)
	return args.Error(0)
}

func (m *MockScheduleRepository) DecrementBookings(scheduleID string) error {
	args := m.Called(scheduleID)
	return args.Error(0)
}

type MockActivityRepository struct {
	mock.Mock
}

func (m *MockActivityRepository) Create(activity *domain.Activity) error {
	args := m.Called(activity)
	return args.Error(0)
}

func (m *MockActivityRepository) GetByID(id string) (*domain.Activity, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Activity), args.Error(1)
}

func (m *MockActivityRepository) GetAll() ([]domain.Activity, error) {
	args := m.Called()
	return args.Get(0).([]domain.Activity), args.Error(1)
}

func (m *MockActivityRepository) Update(activity *domain.Activity) error {
	args := m.Called(activity)
	return args.Error(0)
}

func (m *MockActivityRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockActivityRepository) Exists(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) PublishScheduleEvent(operation, scheduleID, activityID string) error {
	args := m.Called(operation, scheduleID, activityID)
	return args.Error(0)
}

type MockUserClient struct {
	mock.Mock
}

// Tests para CreateSchedule (con concurrencia)

func TestCreateSchedule_Success_ConcurrentValidations(t *testing.T) {
	// Arrange
	mockScheduleRepo := new(MockScheduleRepository)
	mockActivityRepo := new(MockActivityRepository)
	mockPublisher := new(MockPublisher)
	mockUserClient := new(MockUserClient)

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, mockUserClient, mockPublisher)

	activityID := primitive.NewObjectID().Hex()
	req := domain.CreateScheduleRequest{
		Instructor:  "Test Instructor",
		DayOfWeek:   "monday",
		StartTime:   "18:00",
		EndTime:     "19:00",
		Location:    "Sala 1",
		MaxCapacity: 20,
	}

	// Goroutine 1: Activity exists
	mockActivityRepo.On("Exists", activityID).Return(true, nil)

	// Goroutine 2: No conflict
	mockScheduleRepo.On("CheckConflict", "monday", "18:00", "19:00", "Sala 1", "").Return(false, nil)

	// Goroutine 3: Time format validation (no mock needed, it's internal)

	// Create schedule
	mockScheduleRepo.On("Create", mock.AnythingOfType("*domain.Schedule")).Return(nil)

	// Publish event
	mockPublisher.On("PublishScheduleEvent", "CREATE", mock.Anything, activityID).Return(nil)

	// Act
	response, err := service.CreateSchedule(activityID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, activityID, response.ActivityID)
	assert.Equal(t, "Test Instructor", response.Instructor)
	assert.Equal(t, "monday", response.DayOfWeek)
	assert.Equal(t, "18:00", response.StartTime)
	assert.Equal(t, "19:00", response.EndTime)
	assert.Equal(t, "Sala 1", response.Location)
	assert.Equal(t, 20, response.MaxCapacity)
	assert.Equal(t, 0, response.CurrentBookings)
	assert.Equal(t, 20, response.AvailableSpots)
	assert.Equal(t, "active", response.Status)

	mockActivityRepo.AssertExpectations(t)
	mockScheduleRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestCreateSchedule_ActivityNotFound(t *testing.T) {
	// Arrange
	mockScheduleRepo := new(MockScheduleRepository)
	mockActivityRepo := new(MockActivityRepository)
	mockPublisher := new(MockPublisher)
	mockUserClient := new(MockUserClient)

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, mockUserClient, mockPublisher)

	activityID := primitive.NewObjectID().Hex()
	req := domain.CreateScheduleRequest{
		Instructor:  "Test Instructor",
		DayOfWeek:   "monday",
		StartTime:   "18:00",
		EndTime:     "19:00",
		Location:    "Sala 1",
		MaxCapacity: 20,
	}

	// Goroutine 1: Activity does NOT exist
	mockActivityRepo.On("Exists", activityID).Return(false, nil)

	// Goroutine 2: No conflict
	mockScheduleRepo.On("CheckConflict", "monday", "18:00", "19:00", "Sala 1", "").Return(false, nil)

	// Act
	response, err := service.CreateSchedule(activityID, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrActivityNotFound, err)
	assert.Nil(t, response)

	mockActivityRepo.AssertExpectations(t)
	mockScheduleRepo.AssertExpectations(t)
}

func TestCreateSchedule_ScheduleConflict(t *testing.T) {
	// Arrange
	mockScheduleRepo := new(MockScheduleRepository)
	mockActivityRepo := new(MockActivityRepository)
	mockPublisher := new(MockPublisher)
	mockUserClient := new(MockUserClient)

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, mockUserClient, mockPublisher)

	activityID := primitive.NewObjectID().Hex()
	req := domain.CreateScheduleRequest{
		Instructor:  "Test Instructor",
		DayOfWeek:   "monday",
		StartTime:   "18:00",
		EndTime:     "19:00",
		Location:    "Sala 1",
		MaxCapacity: 20,
	}

	// Goroutine 1: Activity exists
	mockActivityRepo.On("Exists", activityID).Return(true, nil)

	// Goroutine 2: Conflict detected
	mockScheduleRepo.On("CheckConflict", "monday", "18:00", "19:00", "Sala 1", "").Return(true, nil)

	// Act
	response, err := service.CreateSchedule(activityID, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrScheduleConflict, err)
	assert.Nil(t, response)

	mockActivityRepo.AssertExpectations(t)
	mockScheduleRepo.AssertExpectations(t)
}

func TestCreateSchedule_InvalidTimeFormat(t *testing.T) {
	// Arrange
	mockScheduleRepo := new(MockScheduleRepository)
	mockActivityRepo := new(MockActivityRepository)
	mockPublisher := new(MockPublisher)
	mockUserClient := new(MockUserClient)

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, mockUserClient, mockPublisher)

	activityID := primitive.NewObjectID().Hex()
	req := domain.CreateScheduleRequest{
		Instructor:  "Test Instructor",
		DayOfWeek:   "monday",
		StartTime:   "25:00", // Invalid time format
		EndTime:     "19:00",
		Location:    "Sala 1",
		MaxCapacity: 20,
	}

	// Goroutine 1: Activity exists
	mockActivityRepo.On("Exists", activityID).Return(true, nil)

	// Goroutine 2: No conflict
	mockScheduleRepo.On("CheckConflict", "monday", "25:00", "19:00", "Sala 1", "").Return(false, nil)

	// Goroutine 3: Time format validation will fail

	// Act
	response, err := service.CreateSchedule(activityID, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidTimeFormat, err)
	assert.Nil(t, response)

	mockActivityRepo.AssertExpectations(t)
	mockScheduleRepo.AssertExpectations(t)
}

// Tests para IncrementBookings y DecrementBookings

func TestIncrementBookings_Success(t *testing.T) {
	// Arrange
	mockScheduleRepo := new(MockScheduleRepository)
	mockActivityRepo := new(MockActivityRepository)
	mockPublisher := new(MockPublisher)
	mockUserClient := new(MockUserClient)

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, mockUserClient, mockPublisher)

	scheduleID := primitive.NewObjectID().Hex()
	mockScheduleRepo.On("IncrementBookings", scheduleID).Return(nil)

	// Act
	err := service.IncrementBookings(scheduleID)

	// Assert
	assert.NoError(t, err)
	mockScheduleRepo.AssertExpectations(t)
}

func TestDecrementBookings_Success(t *testing.T) {
	// Arrange
	mockScheduleRepo := new(MockScheduleRepository)
	mockActivityRepo := new(MockActivityRepository)
	mockPublisher := new(MockPublisher)
	mockUserClient := new(MockUserClient)

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, mockUserClient, mockPublisher)

	scheduleID := primitive.NewObjectID().Hex()
	mockScheduleRepo.On("DecrementBookings", scheduleID).Return(nil)

	// Act
	err := service.DecrementBookings(scheduleID)

	// Assert
	assert.NoError(t, err)
	mockScheduleRepo.AssertExpectations(t)
}

// Tests para GetScheduleByID

func TestGetScheduleByID_Success(t *testing.T) {
	// Arrange
	mockScheduleRepo := new(MockScheduleRepository)
	mockActivityRepo := new(MockActivityRepository)
	mockPublisher := new(MockPublisher)
	mockUserClient := new(MockUserClient)

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, mockUserClient, mockPublisher)

	scheduleID := primitive.NewObjectID().Hex()
	activityID := primitive.NewObjectID().Hex()

	schedule := &domain.Schedule{
		ID:              scheduleID,
		ActivityID:      activityID,
		Instructor:      "Test Instructor",
		DayOfWeek:       "monday",
		StartTime:       "18:00",
		EndTime:         "19:00",
		Location:        "Sala 1",
		MaxCapacity:     20,
		CurrentBookings: 5,
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mockScheduleRepo.On("GetByID", scheduleID).Return(schedule, nil)

	// Act
	response, err := service.GetScheduleByID(scheduleID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, scheduleID, response.ID)
	assert.Equal(t, activityID, response.ActivityID)
	assert.Equal(t, "Test Instructor", response.Instructor)
	assert.Equal(t, 5, response.CurrentBookings)
	assert.Equal(t, 15, response.AvailableSpots)

	mockScheduleRepo.AssertExpectations(t)
}

func TestGetScheduleByID_NotFound(t *testing.T) {
	// Arrange
	mockScheduleRepo := new(MockScheduleRepository)
	mockActivityRepo := new(MockActivityRepository)
	mockPublisher := new(MockPublisher)
	mockUserClient := new(MockUserClient)

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, mockUserClient, mockPublisher)

	scheduleID := primitive.NewObjectID().Hex()
	mockScheduleRepo.On("GetByID", scheduleID).Return(nil, errors.New("not found"))

	// Act
	response, err := service.GetScheduleByID(scheduleID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)

	mockScheduleRepo.AssertExpectations(t)
}
