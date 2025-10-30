package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetMongoURI obtiene la URI de MongoDB desde variables de entorno
func GetMongoURI() string {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return "mongodb://admin:admin123@localhost:27017"
	}
	return uri
}

// GetDatabaseName obtiene el nombre de la base de datos
func GetDatabaseName() string {
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		return "gym_db"
	}
	return dbName
}

// ConnectDatabase conecta a MongoDB
func ConnectDatabase() (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Configurar opciones de cliente
	clientOptions := options.Client().ApplyURI(GetMongoURI())

	// Conectar a MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// Verificar la conexión
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	db := client.Database(GetDatabaseName())
	log.Println("✅ MongoDB connected successfully")
	return db, nil
}
