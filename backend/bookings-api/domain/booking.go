package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Booking representa una reserva de un usuario a un horario
type Booking struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     uint               `bson:"user_id" json:"user_id"`
	ScheduleID string             `bson:"schedule_id" json:"schedule_id"`

	// Información desnormalizada (snapshot al momento de reservar)
	ActivityName     string  `bson:"activity_name" json:"activity_name"`
	ActivityCategory string  `bson:"activity_category" json:"activity_category"`
	Instructor       string  `bson:"instructor" json:"instructor"`
	DayOfWeek        string  `bson:"day_of_week" json:"day_of_week"`
	StartTime        string  `bson:"start_time" json:"start_time"`
	EndTime          string  `bson:"end_time" json:"end_time"`
	Location         string  `bson:"location" json:"location"`
	Price            float64 `bson:"price" json:"price"`

	Status    string    `bson:"status" json:"status"` // "confirmed", "cancelled", "attended"
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
