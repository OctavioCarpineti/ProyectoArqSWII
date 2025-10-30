package services

import (
	"users-api/domain"
	"users-api/repositories"
	"users-api/utils"
)

// UserServiceImpl implementa UserService
type UserServiceImpl struct {
	userRepo repositories.UserRepository
}

// NewUserService crea una nueva instancia del servicio
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &UserServiceImpl{
		userRepo: userRepo,
	}
}

// CreateUser crea un nuevo usuario
func (s *UserServiceImpl) CreateUser(req domain.CreateUserRequest) (*domain.UserResponse, error) {
	// Validar que el usuario no exista
	exists, err := s.userRepo.Exists(req.Username, req.Email)
	if err != nil {
		return nil, domain.ErrDatabaseQuery
	}
	if exists {
		return nil, domain.ErrUserAlreadyExists
	}

	// Hashear la contraseña
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Crear el usuario
	user := &domain.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "normal", // Por defecto los usuarios son normales
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	// Convertir a UserResponse (sin password)
	response := user.ToUserResponse()
	return &response, nil
}

// GetUserByID obtiene un usuario por su ID
func (s *UserServiceImpl) GetUserByID(id uint) (*domain.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := user.ToUserResponse()
	return &response, nil
}

// Login autentica un usuario y devuelve un token JWT
func (s *UserServiceImpl) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
	// Intentar buscar por username
	user, err := s.userRepo.GetByUsername(req.UsernameOrEmail)

	// Si no se encuentra por username, intentar por email
	if err == domain.ErrUserNotFound {
		user, err = s.userRepo.GetByEmail(req.UsernameOrEmail)
	}

	// Si aún no se encuentra, credenciales inválidas
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verificar la contraseña
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	// Generar token JWT
	token, err := utils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	// Preparar respuesta
	response := &domain.LoginResponse{
		Token: token,
		User:  user.ToUserResponse(),
	}

	return response, nil
}
