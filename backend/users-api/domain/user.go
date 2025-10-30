package domain

import (
	"time"
)

// User representa la entidad de usuario en la base de datos
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"` // El "-" oculta el password en JSON
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `gorm:"default:'normal'" json:"role"` // "normal" o "admin"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName especifica el nombre de la tabla en la base de datos
func (User) TableName() string {
	return "users"
}
