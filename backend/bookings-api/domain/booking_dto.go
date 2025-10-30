package domain

import "time"

// CreateBookingRequest representa la request para crear una reserva
type CreateBookingRequest struct {
	UserID     uint   `json:"user_id" binding:"required"`
	ScheduleID string `json:"schedule_id" binding:"required"`
}

// BookingResponse representa la respuesta con datos de una reserva
type BookingResponse struct {
	ID               string  `json:"id"`
	UserID           uint    `json:"user_id"`
	ScheduleID       string  `json:"schedule_id"`
	ActivityName     string  `json:"activity_name"`
	ActivityCategory string  `json:"activity_category"`
	Instructor       string  `json:"instructor"`
	DayOfWeek        string  `json:"day_of_week"`
	StartTime        string  `json:"start_time"`
	EndTime          string  `json:"end_time"`
	Location         string  `json:"location"`
	Price            float64 `json:"price"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// ToBookingResponse convierte Booking a BookingResponse
func (b *Booking) ToBookingResponse() BookingResponse {
	return BookingResponse{
		ID:               b.ID.Hex(),
		UserID:           b.UserID,
		ScheduleID:       b.ScheduleID,
		ActivityName:     b.ActivityName,
		ActivityCategory: b.ActivityCategory,
		Instructor:       b.Instructor,
		DayOfWeek:        b.DayOfWeek,
		StartTime:        b.StartTime,
		EndTime:          b.EndTime,
		Location:         b.Location,
		Price:            b.Price,
		Status:           b.Status,
		CreatedAt:        b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        b.UpdatedAt.Format(time.RFC3339),
	}
}
