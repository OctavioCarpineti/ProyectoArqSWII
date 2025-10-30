package repositories

import (
	"context"
	"time"

	"activities-api/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ActivityMongoRepository implementa ActivityRepository usando MongoDB
type ActivityMongoRepository struct {
	collection *mongo.Collection
}

// NewActivityMongoRepository crea una nueva instancia del repositorio
func NewActivityMongoRepository(db *mongo.Database) ActivityRepository {
	return &ActivityMongoRepository{
		collection: db.Collection("activities"),
	}
}

// Create crea una nueva actividad
func (r *ActivityMongoRepository) Create(activity *domain.Activity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, activity)
	if err != nil {
		return err
	}

	activity.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID obtiene una actividad por su ID
func (r *ActivityMongoRepository) GetByID(id string) (*domain.Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrActivityNotFound
	}

	var activity domain.Activity
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&activity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrActivityNotFound
		}
		return nil, err
	}

	return &activity, nil
}

// GetAll obtiene todas las actividades
func (r *ActivityMongoRepository) GetAll() ([]domain.Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []domain.Activity
	if err = cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

// GetByOwnerID obtiene todas las actividades de un owner
func (r *ActivityMongoRepository) GetByOwnerID(ownerID uint) ([]domain.Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{"owner_id": ownerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []domain.Activity
	if err = cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

// Update actualiza una actividad existente
func (r *ActivityMongoRepository) Update(activity *domain.Activity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	activity.UpdatedAt = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": activity.ID},
		bson.M{"$set": activity},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrActivityNotFound
	}

	return nil
}

// Delete elimina una actividad
func (r *ActivityMongoRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrActivityNotFound
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrActivityNotFound
	}

	return nil
}

// Exists verifica si existe una actividad
func (r *ActivityMongoRepository) Exists(id string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, nil
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": objectID})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
