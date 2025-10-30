package services

import "users-api/domain"

// UserService define la interface para el servicio de usuarios
type UserService interface {
	// CreateUser crea un nuevo usuario
	CreateUser(req domain.CreateUserRequest) (*domain.UserResponse, error)

	// GetUserByID obtiene un usuario por su ID
	GetUserByID(id uint) (*domain.UserResponse, error)

	// Login autentica un usuario y devuelve un token JWT
	Login(req domain.LoginRequest) (*domain.LoginResponse, error)
}
