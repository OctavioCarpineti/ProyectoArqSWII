package config

import (
	"fmt"
	"log"
	"os"
	"time"
	"users-api/domain"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetDBConfig obtiene la configuración de la base de datos desde variables de entorno
func GetDBConfig() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "3307")
	user := getEnv("DB_USER", "gym_user")
	password := getEnv("DB_PASSWORD", "gym_pass")
	dbname := getEnv("DB_NAME", "gym_users")

	// Formato DSN de MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	return dsn
}

// ConnectDatabase conecta a la base de datos MySQL usando GORM
func ConnectDatabase() (*gorm.DB, error) {
	dsn := GetDBConfig()

	// Configuración de logger para GORM
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Conectar a la base de datos
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configurar el pool de conexiones
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migrar el schema (crea/actualiza tablas)
	err = db.AutoMigrate(&domain.User{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("✅ Database connected successfully")
	return db, nil
}

// getEnv obtiene una variable de entorno o devuelve un valor por defecto
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
