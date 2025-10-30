package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Activity representa una actividad del gimnasio
type Activity struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerID     uint               `bson:"owner_id" json:"owner_id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Category    string             `bson:"category" json:"category"`
	Duration    int                `bson:"duration" json:"duration"` // en minutos
	Price       float64            `bson:"price" json:"price"`
	ImageURL    string             `bson:"image_url,omitempty" json:"image_url,omitempty"`
	Status      string             `bson:"status" json:"status"` // "active", "inactive"
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// ValidCategories son las categorías válidas
var ValidCategories = []string{
	"Yoga", "Spinning", "Pilates", "Funcional", "CrossFit", "Zumba", "Boxing", "Natacion",
}
