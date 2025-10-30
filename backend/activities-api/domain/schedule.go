package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Schedule representa un horario específico de una actividad
type Schedule struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ActivityID      string             `bson:"activity_id" json:"activity_id"`
	Instructor      string             `bson:"instructor" json:"instructor"`
	DayOfWeek       string             `bson:"day_of_week" json:"day_of_week"`
	StartTime       string             `bson:"start_time" json:"start_time"`
	EndTime         string             `bson:"end_time" json:"end_time"`
	Location        string             `bson:"location" json:"location"`
	MaxCapacity     int                `bson:"max_capacity" json:"max_capacity"`
	CurrentBookings int                `bson:"current_bookings" json:"current_bookings"`
	Status          string             `bson:"status" json:"status"` // "active", "cancelled", "full"
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

// ValidDaysOfWeek son los días válidos
var ValidDaysOfWeek = []string{
	"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday",
}
