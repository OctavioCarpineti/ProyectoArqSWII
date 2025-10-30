package repositories

import (
	"bookings-api/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// BookingMongoRepository implementa BookingRepository usando MongoDB
type BookingMongoRepository struct {
	collection *mongo.Collection
}

// NewBookingMongoRepository crea una nueva instancia del repositorio
func NewBookingMongoRepository(db *mongo.Database) BookingRepository {
	return &BookingMongoRepository{
		collection: db.Collection("bookings"),
	}
}

// Create crea una nueva reserva
func (r *BookingMongoRepository) Create(booking *domain.Booking) error {
	booking.ID = primitive.NewObjectID()
	booking.CreatedAt = time.Now()
	booking.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, booking)
	return err
}

// GetByID obtiene una reserva por su ID
func (r *BookingMongoRepository) GetByID(id string) (*domain.Booking, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrBookingNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var booking domain.Booking
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&booking)
	if err == mongo.ErrNoDocuments {
		return nil, domain.ErrBookingNotFound
	}
	if err != nil {
		return nil, err
	}

	return &booking, nil
}

// GetByUserID obtiene todas las reservas de un usuario
func (r *BookingMongoRepository) GetByUserID(userID uint) ([]domain.Booking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id": userID,
		"status":  bson.M{"$ne": "cancelled"}, // Excluir reservas canceladas
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []domain.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}

// GetByScheduleID obtiene todas las reservas de un horario
func (r *BookingMongoRepository) GetByScheduleID(scheduleID string) ([]domain.Booking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"schedule_id": scheduleID,
		"status":      "confirmed",
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []domain.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}

// CheckDuplicateBooking verifica si ya existe una reserva del usuario para ese schedule
func (r *BookingMongoRepository) CheckDuplicateBooking(userID uint, scheduleID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":     userID,
		"schedule_id": scheduleID,
		"status":      "confirmed",
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Update actualiza una reserva
func (r *BookingMongoRepository) Update(booking *domain.Booking) error {
	booking.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": booking.ID}
	update := bson.M{"$set": booking}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrBookingNotFound
	}

	return nil
}

// Delete elimina una reserva (soft delete cambiando status)
func (r *BookingMongoRepository) Delete(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrBookingNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"status":     "cancelled",
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrBookingNotFound
	}

	return nil
}
