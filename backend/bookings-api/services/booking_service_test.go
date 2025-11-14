package services

import (
	"bookings-api/clients"
	"bookings-api/domain"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mocks

type MockBookingRepository struct {
	mock.Mock
}

func (m *MockBookingRepository) Create(booking *domain.Booking) error {
	args := m.Called(booking)
	return args.Error(0)
}

func (m *MockBookingRepository) GetByID(id string) (*domain.Booking, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *MockBookingRepository) GetByUserID(userID uint) ([]domain.Booking, error) {
	args := m.Called(userID)
	return args.Get(0).([]domain.Booking), args.Error(1)
}

func (m *MockBookingRepository) Update(booking *domain.Booking) error {
	args := m.Called(booking)
	return args.Error(0)
}

func (m *MockBookingRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockBookingRepository) CheckDuplicateBooking(userID uint, scheduleID string) (bool, error) {
	args := m.Called(userID, scheduleID)
	return args.Bool(0), args.Error(1)
}

type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetUserByID(userID uint) (*clients.UserResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clients.UserResponse), args.Error(1)
}

type MockActivitiesClient struct {
	mock.Mock
}

func (m *MockActivitiesClient) GetScheduleByID(scheduleID string) (*clients.ScheduleResponse, error) {
	args := m.Called(scheduleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clients.ScheduleResponse), args.Error(1)
}

func (m *MockActivitiesClient) GetActivityByID(activityID string) (*clients.ActivityResponse, error) {
	args := m.Called(activityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*clients.ActivityResponse), args.Error(1)
}

func (m *MockActivitiesClient) IncrementBookings(scheduleID, token string) error {
	args := m.Called(scheduleID, token)
	return args.Error(0)
}

func (m *MockActivitiesClient) DecrementBookings(scheduleID, token string) error {
	args := m.Called(scheduleID, token)
	return args.Error(0)
}

// Tests para CreateBooking (con concurrencia)

func TestCreateBooking_Success_ConcurrentValidations(t *testing.T) {
	// Arrange
	mockBookingRepo := new(MockBookingRepository)
	mockUserClient := new(MockUserClient)
	mockActivitiesClient := new(MockActivitiesClient)

	service := NewBookingService(mockBookingRepo, mockUserClient, mockActivitiesClient)

	scheduleID := primitive.NewObjectID().Hex()
	activityID := primitive.NewObjectID().Hex()

	req := domain.CreateBookingRequest{
		UserID:     1,
		ScheduleID: scheduleID,
	}
	token := "test-token"

	// Goroutine 1: User exists
	userResp := &clients.UserResponse{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}
	mockUserClient.On("GetUserByID", uint(1)).Return(userResp, nil)

	// Goroutine 2: Schedule exists and has capacity
	scheduleResp := &clients.ScheduleResponse{
		ID:              scheduleID,
		ActivityID:      activityID,
		Instructor:      "Test Instructor",
		DayOfWeek:       "monday",
		StartTime:       "18:00",
		EndTime:         "19:00",
		Location:        "Sala 1",
		MaxCapacity:     20,
		CurrentBookings: 5,
		AvailableSpots:  15,
		Status:          "active",
	}
	mockActivitiesClient.On("GetScheduleByID", scheduleID).Return(scheduleResp, nil)

	// Get activity details
	activityResp := &clients.ActivityResponse{
		ID:       activityID,
		Name:     "Yoga Intermedio",
		Category: "Yoga",
		Price:    1500,
	}
	mockActivitiesClient.On("GetActivityByID", activityID).Return(activityResp, nil)

	// Goroutine 3: No duplicate booking
	mockBookingRepo.On("CheckDuplicateBooking", uint(1), scheduleID).Return(false, nil)

	// Create booking
	mockBookingRepo.On("Create", mock.AnythingOfType("*domain.Booking")).Return(nil)

	// Increment bookings in activities-api
	mockActivitiesClient.On("IncrementBookings", scheduleID, token).Return(nil)

	// Act
	response, err := service.CreateBooking(req, token)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, uint(1), response.UserID)
	assert.Equal(t, scheduleID, response.ScheduleID)
	assert.Equal(t, "Yoga Intermedio", response.ActivityName)
	assert.Equal(t, "Yoga", response.ActivityCategory)
	assert.Equal(t, "Test Instructor", response.Instructor)
	assert.Equal(t, float64(1500), response.Price)
	assert.Equal(t, "confirmed", response.Status)

	mockUserClient.AssertExpectations(t)
	mockActivitiesClient.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
}

func TestCreateBooking_UserNotFound(t *testing.T) {
	// Arrange
	mockBookingRepo := new(MockBookingRepository)
	mockUserClient := new(MockUserClient)
	mockActivitiesClient := new(MockActivitiesClient)

	service := NewBookingService(mockBookingRepo, mockUserClient, mockActivitiesClient)

	scheduleID := primitive.NewObjectID().Hex()
	req := domain.CreateBookingRequest{
		UserID:     999,
		ScheduleID: scheduleID,
	}
	token := "test-token"

	// Goroutine 1: User does NOT exist
	mockUserClient.On("GetUserByID", uint(999)).Return(nil, errors.New("user not found"))

	// Goroutine 2: Schedule exists
	scheduleResp := &clients.ScheduleResponse{
		ID:              scheduleID,
		MaxCapacity:     20,
		CurrentBookings: 5,
		Status:          "active",
	}
	mockActivitiesClient.On("GetScheduleByID", scheduleID).Return(scheduleResp, nil)

	// Goroutine 3: No duplicate
	mockBookingRepo.On("CheckDuplicateBooking", uint(999), scheduleID).Return(false, nil)

	// Act
	response, err := service.CreateBooking(req, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidUser, err)
	assert.Nil(t, response)

	mockUserClient.AssertExpectations(t)
	mockActivitiesClient.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
}

func TestCreateBooking_ScheduleNotFound(t *testing.T) {
	// Arrange
	mockBookingRepo := new(MockBookingRepository)
	mockUserClient := new(MockUserClient)
	mockActivitiesClient := new(MockActivitiesClient)

	service := NewBookingService(mockBookingRepo, mockUserClient, mockActivitiesClient)

	scheduleID := primitive.NewObjectID().Hex()
	req := domain.CreateBookingRequest{
		UserID:     1,
		ScheduleID: scheduleID,
	}
	token := "test-token"

	// Goroutine 1: User exists
	userResp := &clients.UserResponse{ID: 1, Username: "testuser"}
	mockUserClient.On("GetUserByID", uint(1)).Return(userResp, nil)

	// Goroutine 2: Schedule does NOT exist
	mockActivitiesClient.On("GetScheduleByID", scheduleID).Return(nil, errors.New("not found"))

	// Goroutine 3: No duplicate
	mockBookingRepo.On("CheckDuplicateBooking", uint(1), scheduleID).Return(false, nil)

	// Act
	response, err := service.CreateBooking(req, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidSchedule, err)
	assert.Nil(t, response)

	mockUserClient.AssertExpectations(t)
	mockActivitiesClient.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
}

func TestCreateBooking_ScheduleFull(t *testing.T) {
	// Arrange
	mockBookingRepo := new(MockBookingRepository)
	mockUserClient := new(MockUserClient)
	mockActivitiesClient := new(MockActivitiesClient)

	service := NewBookingService(mockBookingRepo, mockUserClient, mockActivitiesClient)

	scheduleID := primitive.NewObjectID().Hex()
	req := domain.CreateBookingRequest{
		UserID:     1,
		ScheduleID: scheduleID,
	}
	token := "test-token"

	// Goroutine 1: User exists
	userResp := &clients.UserResponse{ID: 1, Username: "testuser"}
	mockUserClient.On("GetUserByID", uint(1)).Return(userResp, nil)

	// Goroutine 2: Schedule is FULL
	scheduleResp := &clients.ScheduleResponse{
		ID:              scheduleID,
		MaxCapacity:     20,
		CurrentBookings: 20, // Full!
		Status:          "active",
	}
	mockActivitiesClient.On("GetScheduleByID", scheduleID).Return(scheduleResp, nil)

	// Goroutine 3: No duplicate
	mockBookingRepo.On("CheckDuplicateBooking", uint(1), scheduleID).Return(false, nil)

	// Act
	response, err := service.CreateBooking(req, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrScheduleFull, err)
	assert.Nil(t, response)

	mockUserClient.AssertExpectations(t)
	mockActivitiesClient.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
}

func TestCreateBooking_DuplicateBooking(t *testing.T) {
	// Arrange
	mockBookingRepo := new(MockBookingRepository)
	mockUserClient := new(MockUserClient)
	mockActivitiesClient := new(MockActivitiesClient)

	service := NewBookingService(mockBookingRepo, mockUserClient, mockActivitiesClient)

	scheduleID := primitive.NewObjectID().Hex()
	req := domain.CreateBookingRequest{
		UserID:     1,
		ScheduleID: scheduleID,
	}
	token := "test-token"

	// Goroutine 1: User exists
	userResp := &clients.UserResponse{ID: 1, Username: "testuser"}
	mockUserClient.On("GetUserByID", uint(1)).Return(userResp, nil)

	// Goroutine 2: Schedule exists
	scheduleResp := &clients.ScheduleResponse{
		ID:              scheduleID,
		MaxCapacity:     20,
		CurrentBookings: 5,
		Status:          "active",
	}
	mockActivitiesClient.On("GetScheduleByID", scheduleID).Return(scheduleResp, nil)

	// Goroutine 3: Duplicate booking found
	mockBookingRepo.On("CheckDuplicateBooking", uint(1), scheduleID).Return(true, nil)

	// Act
	response, err := service.CreateBooking(req, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrBookingAlreadyExists, err)
	assert.Nil(t, response)

	mockUserClient.AssertExpectations(t)
	mockActivitiesClient.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
}

// Tests para CancelBooking

func TestCancelBooking_Success(t *testing.T) {
	// Arrange
	mockBookingRepo := new(MockBookingRepository)
	mockUserClient := new(MockUserClient)
	mockActivitiesClient := new(MockActivitiesClient)

	service := NewBookingService(mockBookingRepo, mockUserClient, mockActivitiesClient)

	bookingID := primitive.NewObjectID().Hex()
	scheduleID := primitive.NewObjectID().Hex()
	token := "test-token"

	booking := &domain.Booking{
		ID:         bookingID,
		UserID:     1,
		ScheduleID: scheduleID,
		Status:     "confirmed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockBookingRepo.On("GetByID", bookingID).Return(booking, nil)
	mockBookingRepo.On("Update", mock.AnythingOfType("*domain.Booking")).Return(nil)
	mockActivitiesClient.On("DecrementBookings", scheduleID, token).Return(nil)

	// Act
	err := service.CancelBooking(bookingID, token)

	// Assert
	assert.NoError(t, err)
	mockBookingRepo.AssertExpectations(t)
	mockActivitiesClient.AssertExpectations(t)
}

func TestCancelBooking_NotFound(t *testing.T) {
	// Arrange
	mockBookingRepo := new(MockBookingRepository)
	mockUserClient := new(MockUserClient)
	mockActivitiesClient := new(MockActivitiesClient)

	service := NewBookingService(mockBookingRepo, mockUserClient, mockActivitiesClient)

	bookingID := primitive.NewObjectID().Hex()
	token := "test-token"

	mockBookingRepo.On("GetByID", bookingID).Return(nil, errors.New("not found"))

	// Act
	err := service.CancelBooking(bookingID, token)

	// Assert
	assert.Error(t, err)
	mockBookingRepo.AssertExpectations(t)
}

// Tests para GetBookingByID

func TestGetBookingByID_Success(t *testing.T) {
	// Arrange
	mockBookingRepo := new(MockBookingRepository)
	mockUserClient := new(MockUserClient)
	mockActivitiesClient := new(MockActivitiesClient)

	service := NewBookingService(mockBookingRepo, mockUserClient, mockActivitiesClient)

	bookingID := primitive.NewObjectID().Hex()
	booking := &domain.Booking{
		ID:             bookingID,
		UserID:         1,
		ScheduleID:     "schedule123",
		ActivityName:   "Yoga",
		ActivityCategory: "Fitness",
		Instructor:     "Test",
		DayOfWeek:      "monday",
		StartTime:      "18:00",
		EndTime:        "19:00",
		Location:       "Sala 1",
		Price:          1500,
		Status:         "confirmed",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mockBookingRepo.On("GetByID", bookingID).Return(booking, nil)

	// Act
	response, err := service.GetBookingByID(bookingID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, bookingID, response.ID)
	assert.Equal(t, uint(1), response.UserID)
	assert.Equal(t, "Yoga", response.ActivityName)

	mockBookingRepo.AssertExpectations(t)
}
