package services

import (
	"activities-api/domain"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ============================================
// MOCKS
// ============================================

// MockScheduleRepository implementa ScheduleRepository para tests
type MockScheduleRepository struct {
	CreateFunc            func(*domain.Schedule) error
	GetByIDFunc           func(string) (*domain.Schedule, error)
	GetByActivityIDFunc   func(string) ([]domain.Schedule, error)
	GetAllFunc            func() ([]domain.Schedule, error)
	UpdateFunc            func(*domain.Schedule) error
	DeleteFunc            func(string) error
	CheckConflictFunc     func(string, string, string, string, string) (bool, error)
	IncrementBookingsFunc func(string) error
	DecrementBookingsFunc func(string) error
}

func (m *MockScheduleRepository) Create(schedule *domain.Schedule) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(schedule)
	}
	schedule.ID = primitive.NewObjectID()
	return nil
}

func (m *MockScheduleRepository) GetByID(id string) (*domain.Schedule, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *MockScheduleRepository) GetByActivityID(activityID string) ([]domain.Schedule, error) {
	if m.GetByActivityIDFunc != nil {
		return m.GetByActivityIDFunc(activityID)
	}
	return nil, nil
}

func (m *MockScheduleRepository) GetAll() ([]domain.Schedule, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return nil, nil
}

func (m *MockScheduleRepository) Update(schedule *domain.Schedule) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(schedule)
	}
	return nil
}

func (m *MockScheduleRepository) Delete(id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m *MockScheduleRepository) CheckConflict(dayOfWeek, startTime, endTime, location, excludeID string) (bool, error) {
	if m.CheckConflictFunc != nil {
		return m.CheckConflictFunc(dayOfWeek, startTime, endTime, location, excludeID)
	}
	return false, nil
}

func (m *MockScheduleRepository) IncrementBookings(id string) error {
	if m.IncrementBookingsFunc != nil {
		return m.IncrementBookingsFunc(id)
	}
	return nil
}

func (m *MockScheduleRepository) DecrementBookings(id string) error {
	if m.DecrementBookingsFunc != nil {
		return m.DecrementBookingsFunc(id)
	}
	return nil
}

// MockActivityRepository implementa ActivityRepository para tests
type MockActivityRepository struct {
	CreateFunc       func(*domain.Activity) error
	GetByIDFunc      func(string) (*domain.Activity, error)
	GetByOwnerIDFunc func(uint) ([]domain.Activity, error)
	GetAllFunc       func() ([]domain.Activity, error)
	UpdateFunc       func(*domain.Activity) error
	DeleteFunc       func(string) error
	ExistsFunc       func(string) (bool, error)
}

func (m *MockActivityRepository) Create(activity *domain.Activity) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(activity)
	}
	return nil
}

func (m *MockActivityRepository) GetByID(id string) (*domain.Activity, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	objID, _ := primitive.ObjectIDFromHex(id)
	return &domain.Activity{ID: objID, Name: "Test Activity"}, nil
}

func (m *MockActivityRepository) GetByOwnerID(ownerID uint) ([]domain.Activity, error) {
	if m.GetByOwnerIDFunc != nil {
		return m.GetByOwnerIDFunc(ownerID)
	}
	return nil, nil
}

func (m *MockActivityRepository) GetAll() ([]domain.Activity, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return nil, nil
}

func (m *MockActivityRepository) Update(activity *domain.Activity) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(activity)
	}
	return nil
}

func (m *MockActivityRepository) Delete(id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m *MockActivityRepository) Exists(id string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(id)
	}
	return true, nil
}

// ============================================
// TESTS
// ============================================

func TestCreateSchedule_Success(t *testing.T) {
	// Arrange
	mockScheduleRepo := &MockScheduleRepository{
		CheckConflictFunc: func(dayOfWeek, startTime, endTime, location, excludeID string) (bool, error) {
			return false, nil // No conflict
		},
		CreateFunc: func(schedule *domain.Schedule) error {
			schedule.ID = primitive.NewObjectID()
			return nil
		},
	}

	mockActivityRepo := &MockActivityRepository{
		ExistsFunc: func(id string) (bool, error) {
			return true, nil // Activity exists
		},
		GetByIDFunc: func(id string) (*domain.Activity, error) {
			objID, _ := primitive.ObjectIDFromHex(id)
			return &domain.Activity{
				ID:       objID,
				Name:     "Yoga",
				Category: "flexibility",
			}, nil
		},
	}

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, nil, nil)

	req := domain.CreateScheduleRequest{
		DayOfWeek:   "monday",
		StartTime:   "09:00",
		EndTime:     "10:00",
		MaxCapacity: 20,
		Instructor:  "John Doe",
		Location:    "Room A",
	}

	// Act
	result, err := service.CreateSchedule("activity123", req)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	if result.DayOfWeek != "monday" {
		t.Errorf("Expected DayOfWeek to be 'monday', got %s", result.DayOfWeek)
	}

	if result.MaxCapacity != 20 {
		t.Errorf("Expected MaxCapacity to be 20, got %d", result.MaxCapacity)
	}
}

func TestCreateSchedule_ActivityNotFound(t *testing.T) {
	// Arrange
	mockScheduleRepo := &MockScheduleRepository{
		CheckConflictFunc: func(dayOfWeek, startTime, endTime, location, excludeID string) (bool, error) {
			return false, nil
		},
	}

	mockActivityRepo := &MockActivityRepository{
		ExistsFunc: func(id string) (bool, error) {
			return false, nil // Activity does NOT exist
		},
	}

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, nil, nil)

	req := domain.CreateScheduleRequest{
		DayOfWeek:   "monday",
		StartTime:   "09:00",
		EndTime:     "10:00",
		MaxCapacity: 20,
		Instructor:  "John Doe",
		Location:    "Room A",
	}

	// Act
	result, err := service.CreateSchedule("nonexistent", req)

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil")
	}

	if !errors.Is(err, domain.ErrActivityNotFound) {
		t.Errorf("Expected ErrActivityNotFound, got %v", err)
	}
}

func TestCreateSchedule_ConflictDetected(t *testing.T) {
	// Arrange
	mockScheduleRepo := &MockScheduleRepository{
		CheckConflictFunc: func(dayOfWeek, startTime, endTime, location, excludeID string) (bool, error) {
			return true, nil // CONFLICT detected
		},
	}

	mockActivityRepo := &MockActivityRepository{
		ExistsFunc: func(id string) (bool, error) {
			return true, nil
		},
	}

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, nil, nil)

	req := domain.CreateScheduleRequest{
		DayOfWeek:   "monday",
		StartTime:   "09:00",
		EndTime:     "10:00",
		MaxCapacity: 20,
		Instructor:  "John Doe",
		Location:    "Room A",
	}

	// Act
	result, err := service.CreateSchedule("activity123", req)

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil")
	}

	if !errors.Is(err, domain.ErrScheduleConflict) {
		t.Errorf("Expected ErrScheduleConflict, got %v", err)
	}
}

func TestCreateSchedule_InvalidTimeFormat(t *testing.T) {
	// Arrange
	mockScheduleRepo := &MockScheduleRepository{
		CheckConflictFunc: func(dayOfWeek, startTime, endTime, location, excludeID string) (bool, error) {
			return false, nil
		},
	}

	mockActivityRepo := &MockActivityRepository{
		ExistsFunc: func(id string) (bool, error) {
			return true, nil
		},
	}

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, nil, nil)

	req := domain.CreateScheduleRequest{
		DayOfWeek:   "monday",
		StartTime:   "25:00", // INVALID time format
		EndTime:     "10:00",
		MaxCapacity: 20,
		Instructor:  "John Doe",
		Location:    "Room A",
	}

	// Act
	result, err := service.CreateSchedule("activity123", req)

	// Assert
	if err == nil {
		t.Error("Expected error for invalid time format, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil")
	}

	if !errors.Is(err, domain.ErrInvalidTimeFormat) {
		t.Errorf("Expected ErrInvalidTimeFormat, got %v", err)
	}
}

func TestGetScheduleByID_Success(t *testing.T) {
	// Arrange
	testID := primitive.NewObjectID()
	expectedSchedule := &domain.Schedule{
		ID:              testID,
		ActivityID:      "activity123",
		DayOfWeek:       "monday",
		StartTime:       "09:00",
		EndTime:         "10:00",
		MaxCapacity:     20,
		CurrentBookings: 5,
		Instructor:      "John Doe",
		Location:        "Room A",
		Status:          "active",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mockScheduleRepo := &MockScheduleRepository{
		GetByIDFunc: func(id string) (*domain.Schedule, error) {
			return expectedSchedule, nil
		},
	}

	mockActivityRepo := &MockActivityRepository{
		GetByIDFunc: func(id string) (*domain.Activity, error) {
			objID, _ := primitive.ObjectIDFromHex(id)
			return &domain.Activity{
				ID:       objID,
				Name:     "Yoga",
				Category: "flexibility",
			}, nil
		},
	}

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, nil, nil)

	// Act
	result, err := service.GetScheduleByID(testID.Hex())

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	if result.CurrentBookings != 5 {
		t.Errorf("Expected CurrentBookings to be 5, got %d", result.CurrentBookings)
	}

	if result.AvailableSpots != 15 { // 20 - 5
		t.Errorf("Expected AvailableSpots to be 15, got %d", result.AvailableSpots)
	}
}

func TestGetScheduleByID_NotFound(t *testing.T) {
	// Arrange
	mockScheduleRepo := &MockScheduleRepository{
		GetByIDFunc: func(id string) (*domain.Schedule, error) {
			return nil, domain.ErrScheduleNotFound
		},
	}

	service := NewScheduleService(mockScheduleRepo, nil, nil, nil)

	// Act
	result, err := service.GetScheduleByID("nonexistent")

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Error("Expected result to be nil")
	}

	if !errors.Is(err, domain.ErrScheduleNotFound) {
		t.Errorf("Expected ErrScheduleNotFound, got %v", err)
	}
}

func TestUpdateCurrentBookings_Increment(t *testing.T) {
	// Arrange
	mockScheduleRepo := &MockScheduleRepository{
		IncrementBookingsFunc: func(scheduleID string) error {
			return nil
		},
	}

	service := NewScheduleService(mockScheduleRepo, nil, nil, nil)

	// Act
	err := service.UpdateCurrentBookings("schedule123", true)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestUpdateCurrentBookings_Decrement(t *testing.T) {
	// Arrange
	mockScheduleRepo := &MockScheduleRepository{
		DecrementBookingsFunc: func(scheduleID string) error {
			return nil
		},
	}

	service := NewScheduleService(mockScheduleRepo, nil, nil, nil)

	// Act
	err := service.UpdateCurrentBookings("schedule123", false)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestGetSchedulesByActivityID_Success(t *testing.T) {
	// Arrange
	testID1 := primitive.NewObjectID()
	testID2 := primitive.NewObjectID()

	expectedSchedules := []domain.Schedule{
		{
			ID:              testID1,
			ActivityID:      "activity123",
			DayOfWeek:       "monday",
			StartTime:       "09:00",
			EndTime:         "10:00",
			MaxCapacity:     20,
			CurrentBookings: 5,
			Status:          "active",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              testID2,
			ActivityID:      "activity123",
			DayOfWeek:       "wednesday",
			StartTime:       "14:00",
			EndTime:         "15:00",
			MaxCapacity:     15,
			CurrentBookings: 3,
			Status:          "active",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	mockScheduleRepo := &MockScheduleRepository{
		GetByActivityIDFunc: func(activityID string) ([]domain.Schedule, error) {
			return expectedSchedules, nil
		},
	}

	mockActivityRepo := &MockActivityRepository{
		GetByIDFunc: func(id string) (*domain.Activity, error) {
			objID, _ := primitive.ObjectIDFromHex(id)
			return &domain.Activity{
				ID:       objID,
				Name:     "Yoga",
				Category: "flexibility",
			}, nil
		},
	}

	service := NewScheduleService(mockScheduleRepo, mockActivityRepo, nil, nil)

	// Act
	result, err := service.GetSchedulesByActivityID("activity123")

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 schedules, got %d", len(result))
	}

	if result[0].AvailableSpots != 15 { // 20 - 5
		t.Errorf("Expected first schedule to have 15 available spots, got %d", result[0].AvailableSpots)
	}

	if result[1].AvailableSpots != 12 { // 15 - 3
		t.Errorf("Expected second schedule to have 12 available spots, got %d", result[1].AvailableSpots)
	}
}
