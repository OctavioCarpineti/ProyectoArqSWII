package domain

import "time"

// CreateScheduleRequest representa la request para crear un horario
type CreateScheduleRequest struct {
	Instructor  string `json:"instructor" binding:"required"`
	DayOfWeek   string `json:"day_of_week" binding:"required,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	StartTime   string `json:"start_time" binding:"required"`
	EndTime     string `json:"end_time" binding:"required"`
	Location    string `json:"location" binding:"required"`
	MaxCapacity int    `json:"max_capacity" binding:"required,min=1"`
}

// UpdateScheduleRequest representa la request para actualizar un horario
type UpdateScheduleRequest struct {
	Instructor  string `json:"instructor"`
	DayOfWeek   string `json:"day_of_week" binding:"omitempty,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Location    string `json:"location"`
	MaxCapacity int    `json:"max_capacity" binding:"omitempty,min=1"`
	Status      string `json:"status" binding:"omitempty,oneof=active cancelled full"`
}

// ScheduleResponse representa la respuesta con datos de un horario
type ScheduleResponse struct {
	ID              string `json:"id"`
	ActivityID      string `json:"activity_id"`
	Instructor      string `json:"instructor"`
	DayOfWeek       string `json:"day_of_week"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Location        string `json:"location"`
	MaxCapacity     int    `json:"max_capacity"`
	CurrentBookings int    `json:"current_bookings"`
	AvailableSpots  int    `json:"available_spots"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// ToScheduleResponse convierte Schedule a ScheduleResponse
func (s *Schedule) ToScheduleResponse() ScheduleResponse {
	return ScheduleResponse{
		ID:              s.ID.Hex(),
		ActivityID:      s.ActivityID,
		Instructor:      s.Instructor,
		DayOfWeek:       s.DayOfWeek,
		StartTime:       s.StartTime,
		EndTime:         s.EndTime,
		Location:        s.Location,
		MaxCapacity:     s.MaxCapacity,
		CurrentBookings: s.CurrentBookings,
		AvailableSpots:  s.MaxCapacity - s.CurrentBookings,
		Status:          s.Status,
		CreatedAt:       s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       s.UpdatedAt.Format(time.RFC3339),
	}
}
