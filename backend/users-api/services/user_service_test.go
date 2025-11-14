package services

import (
	"errors"
	"testing"
	"users-api/domain"
	"users-api/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository es un mock del UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uint) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*domain.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Exists(username, email string) (bool, error) {
	args := m.Called(username, email)
	return args.Bool(0), args.Error(1)
}

// Tests para CreateUser

func TestCreateUser_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	req := domain.CreateUserRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "Password123!",
		FirstName: "Test",
		LastName:  "User",
	}

	mockRepo.On("Exists", "testuser", "test@example.com").Return(false, nil)
	mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

	// Act
	response, err := service.CreateUser(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "testuser", response.Username)
	assert.Equal(t, "test@example.com", response.Email)
	assert.Equal(t, "Test", response.FirstName)
	assert.Equal(t, "User", response.LastName)
	assert.Equal(t, "normal", response.Role)
	mockRepo.AssertExpectations(t)
}

func TestCreateUser_UserAlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	req := domain.CreateUserRequest{
		Username:  "existinguser",
		Email:     "existing@example.com",
		Password:  "Password123!",
		FirstName: "Existing",
		LastName:  "User",
	}

	mockRepo.On("Exists", "existinguser", "existing@example.com").Return(true, nil)

	// Act
	response, err := service.CreateUser(req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserAlreadyExists, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}

func TestCreateUser_DatabaseError(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	req := domain.CreateUserRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "Password123!",
		FirstName: "Test",
		LastName:  "User",
	}

	mockRepo.On("Exists", "testuser", "test@example.com").Return(false, errors.New("database error"))

	// Act
	response, err := service.CreateUser(req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrDatabaseQuery, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}

// Tests para Login

func TestLogin_Success_WithUsername(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	hashedPassword, _ := utils.HashPassword("Password123!")
	user := &domain.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  hashedPassword,
		FirstName: "Test",
		LastName:  "User",
		Role:      "normal",
	}

	req := domain.LoginRequest{
		UsernameOrEmail: "testuser",
		Password:        "Password123!",
	}

	mockRepo.On("GetByUsername", "testuser").Return(user, nil)

	// Act
	response, err := service.Login(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, uint(1), response.User.ID)
	assert.Equal(t, "testuser", response.User.Username)
	mockRepo.AssertExpectations(t)
}

func TestLogin_Success_WithEmail(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	hashedPassword, _ := utils.HashPassword("Password123!")
	user := &domain.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  hashedPassword,
		FirstName: "Test",
		LastName:  "User",
		Role:      "normal",
	}

	req := domain.LoginRequest{
		UsernameOrEmail: "test@example.com",
		Password:        "Password123!",
	}

	mockRepo.On("GetByUsername", "test@example.com").Return(nil, domain.ErrUserNotFound)
	mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)

	// Act
	response, err := service.Login(req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, uint(1), response.User.ID)
	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidCredentials_UserNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	req := domain.LoginRequest{
		UsernameOrEmail: "nonexistent",
		Password:        "Password123!",
	}

	mockRepo.On("GetByUsername", "nonexistent").Return(nil, domain.ErrUserNotFound)
	mockRepo.On("GetByEmail", "nonexistent").Return(nil, domain.ErrUserNotFound)

	// Act
	response, err := service.Login(req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidCredentials, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidCredentials_WrongPassword(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	hashedPassword, _ := utils.HashPassword("CorrectPassword123!")
	user := &domain.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  hashedPassword,
		FirstName: "Test",
		LastName:  "User",
		Role:      "normal",
	}

	req := domain.LoginRequest{
		UsernameOrEmail: "testuser",
		Password:        "WrongPassword123!",
	}

	mockRepo.On("GetByUsername", "testuser").Return(user, nil)

	// Act
	response, err := service.Login(req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidCredentials, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}

// Tests para GetUserByID

func TestGetUserByID_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	user := &domain.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "normal",
	}

	mockRepo.On("GetByID", uint(1)).Return(user, nil)

	// Act
	response, err := service.GetUserByID(1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, uint(1), response.ID)
	assert.Equal(t, "testuser", response.Username)
	assert.Equal(t, "test@example.com", response.Email)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByID_UserNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	mockRepo.On("GetByID", uint(999)).Return(nil, domain.ErrUserNotFound)

	// Act
	response, err := service.GetUserByID(999)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserNotFound, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}
