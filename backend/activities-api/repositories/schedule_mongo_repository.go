package repositories

import (
	"context"
	"time"

	"activities-api/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ScheduleMongoRepository implementa ScheduleRepository usando MongoDB
type ScheduleMongoRepository struct {
	collection *mongo.Collection
}

// NewScheduleMongoRepository crea una nueva instancia del repositorio
func NewScheduleMongoRepository(db *mongo.Database) ScheduleRepository {
	return &ScheduleMongoRepository{
		collection: db.Collection("schedules"),
	}
}

// Create crea un nuevo horario
func (r *ScheduleMongoRepository) Create(schedule *domain.Schedule) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()
	schedule.CurrentBookings = 0

	result, err := r.collection.InsertOne(ctx, schedule)
	if err != nil {
		return err
	}

	schedule.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID obtiene un horario por su ID
func (r *ScheduleMongoRepository) GetByID(id string) (*domain.Schedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrScheduleNotFound
	}

	var schedule domain.Schedule
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&schedule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrScheduleNotFound
		}
		return nil, err
	}

	return &schedule, nil
}

// GetByActivityID obtiene todos los horarios de una actividad
func (r *ScheduleMongoRepository) GetByActivityID(activityID string) ([]domain.Schedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"activity_id": activityID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var schedules []domain.Schedule
	if err = cursor.All(ctx, &schedules); err != nil {
		return nil, err
	}

	return schedules, nil
}

// GetAll obtiene todos los horarios
func (r *ScheduleMongoRepository) GetAll() ([]domain.Schedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var schedules []domain.Schedule
	if err = cursor.All(ctx, &schedules); err != nil {
		return nil, err
	}

	return schedules, nil
}

// Update actualiza un horario existente
func (r *ScheduleMongoRepository) Update(schedule *domain.Schedule) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	schedule.UpdatedAt = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": schedule.ID},
		bson.M{"$set": schedule},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrScheduleNotFound
	}

	return nil
}

// Delete elimina un horario
func (r *ScheduleMongoRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrScheduleNotFound
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrScheduleNotFound
	}

	return nil
}

// CheckConflict verifica si hay conflicto de horario/sala
func (r *ScheduleMongoRepository) CheckConflict(dayOfWeek, startTime, endTime, location string, excludeID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"day_of_week": dayOfWeek,
		"location":    location,
		"$or": []bson.M{
			{
				"start_time": bson.M{"$lt": endTime},
				"end_time":   bson.M{"$gt": startTime},
			},
		},
	}

	// Excluir el horario actual si estamos actualizando
	if excludeID != "" {
		objectID, err := primitive.ObjectIDFromHex(excludeID)
		if err == nil {
			filter["_id"] = bson.M{"$ne": objectID}
		}
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// IncrementBookings incrementa el contador de reservas
func (r *ScheduleMongoRepository) IncrementBookings(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrScheduleNotFound
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$inc": bson.M{"current_bookings": 1},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrScheduleNotFound
	}

	return nil
}

// DecrementBookings decrementa el contador de reservas
func (r *ScheduleMongoRepository) DecrementBookings(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrScheduleNotFound
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$inc": bson.M{"current_bookings": -1},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrScheduleNotFound
	}

	return nil
}
