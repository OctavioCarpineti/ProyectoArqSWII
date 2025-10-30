package domain

import "time"

// CreateActivityRequest representa la request para crear una actividad
type CreateActivityRequest struct {
	OwnerID     uint    `json:"owner_id" binding:"required"`
	Name        string  `json:"name" binding:"required,min=3"`
	Description string  `json:"description"`
	Category    string  `json:"category" binding:"required"`
	Duration    int     `json:"duration" binding:"required,min=30,max=180"`
	Price       float64 `json:"price" binding:"required,min=0"`
	ImageURL    string  `json:"image_url"`
}

// UpdateActivityRequest representa la request para actualizar una actividad
type UpdateActivityRequest struct {
	Name        string  `json:"name" binding:"omitempty,min=3"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Duration    int     `json:"duration" binding:"omitempty,min=30,max=180"`
	Price       float64 `json:"price" binding:"omitempty,min=0"`
	ImageURL    string  `json:"image_url"`
	Status      string  `json:"status" binding:"omitempty,oneof=active inactive"`
}

// ActivityResponse representa la respuesta con datos de una actividad
type ActivityResponse struct {
	ID          string  `json:"id"`
	OwnerID     uint    `json:"owner_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Duration    int     `json:"duration"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url,omitempty"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// ToActivityResponse convierte Activity a ActivityResponse
func (a *Activity) ToActivityResponse() ActivityResponse {
	return ActivityResponse{
		ID:          a.ID.Hex(),
		OwnerID:     a.OwnerID,
		Name:        a.Name,
		Description: a.Description,
		Category:    a.Category,
		Duration:    a.Duration,
		Price:       a.Price,
		ImageURL:    a.ImageURL,
		Status:      a.Status,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   a.UpdatedAt.Format(time.RFC3339),
	}
}
