package domain

import "time"

// ScheduleSearch representa un horario indexado en SolR
// Combina información de Schedule + Activity para búsqueda eficiente
type ScheduleSearch struct {
	// IDs
	ID         string `json:"id"`          // ID del schedule (MongoDB)
	ScheduleID string `json:"schedule_id"` // Mismo que ID (para claridad)
	ActivityID string `json:"activity_id"` // ID de la actividad asociada

	// Información de la actividad (desnormalizada)
	ActivityName     string  `json:"activity_name"`
	ActivityCategory string  `json:"activity_category"`
	ActivityPrice    float64 `json:"price"` // Changed from "activity_price" to match Solr schema

	// Información del horario
	Instructor string `json:"instructor"`
	DayOfWeek  string `json:"day_of_week"` // monday, tuesday, etc.
	StartTime  string `json:"start_time"`  // HH:MM
	EndTime    string `json:"end_time"`    // HH:MM
	Location   string `json:"location"`

	// Capacidad y disponibilidad
	MaxCapacity     int `json:"max_capacity"`
	CurrentBookings int `json:"current_bookings"`
	AvailableSpots  int `json:"available_spots"`

	// Estado
	Status string `json:"status"` // active, cancelled, full

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// IsAvailable verifica si el horario tiene cupos disponibles
func (s *ScheduleSearch) IsAvailable() bool {
	return s.Status == "active" && s.AvailableSpots > 0
}

// CalculateAvailableSpots calcula los cupos disponibles
func (s *ScheduleSearch) CalculateAvailableSpots() {
	s.AvailableSpots = s.MaxCapacity - s.CurrentBookings
	if s.AvailableSpots < 0 {
		s.AvailableSpots = 0
	}
}
