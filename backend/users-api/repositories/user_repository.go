package repositories

import "users-api/domain"

// UserRepository define la interface para el repositorio de usuarios
type UserRepository interface {
	// Create crea un nuevo usuario
	Create(user *domain.User) error

	// GetByID obtiene un usuario por su ID
	GetByID(id uint) (*domain.User, error)

	// GetByUsername obtiene un usuario por su username
	GetByUsername(username string) (*domain.User, error)

	// GetByEmail obtiene un usuario por su email
	GetByEmail(email string) (*domain.User, error)

	// Update actualiza un usuario existente
	Update(user *domain.User) error

	// Delete elimina un usuario
	Delete(id uint) error

	// Exists verifica si existe un usuario con el username o email dado
	Exists(username, email string) (bool, error)
}
