package repositories

import (
	"errors"
	"users-api/domain"

	"gorm.io/gorm"
)

// UserMySQLRepository implementa UserRepository usando MySQL y GORM
type UserMySQLRepository struct {
	db *gorm.DB
}

// NewUserMySQLRepository crea una nueva instancia del repositorio
func NewUserMySQLRepository(db *gorm.DB) UserRepository {
	return &UserMySQLRepository{db: db}
}

// Create crea un nuevo usuario en la base de datos
func (r *UserMySQLRepository) Create(user *domain.User) error {
	result := r.db.Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByID obtiene un usuario por su ID
func (r *UserMySQLRepository) GetByID(id uint) (*domain.User, error) {
	var user domain.User
	result := r.db.First(&user, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}

// GetByUsername obtiene un usuario por su username
func (r *UserMySQLRepository) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("username = ?", username).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}

// GetByEmail obtiene un usuario por su email
func (r *UserMySQLRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}

// Update actualiza un usuario existente
func (r *UserMySQLRepository) Update(user *domain.User) error {
	result := r.db.Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete elimina un usuario por su ID
func (r *UserMySQLRepository) Delete(id uint) error {
	result := r.db.Delete(&domain.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// Exists verifica si existe un usuario con el username o email dado
func (r *UserMySQLRepository) Exists(username, email string) (bool, error) {
	var count int64
	result := r.db.Model(&domain.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}
