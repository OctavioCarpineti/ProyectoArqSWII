package config

import (
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

// GetRabbitMQURL obtiene la URL de RabbitMQ desde variables de entorno
func GetRabbitMQURL() string {
	url := os.Getenv("RABBITMQ_URL")

	// DEBUG: Mostrar qué URL está leyendo
	log.Printf("🔍 DEBUG: Variable RABBITMQ_URL leída: '%s'", url)

	if url == "" {
		log.Println("⚠️  WARNING: RABBITMQ_URL vacía, usando default localhost")
		return "amqp://guest:guest@localhost:5672/"
	}

	log.Printf("✅ Usando RABBITMQ_URL: %s", url)
	return url
}

// ConnectRabbitMQ conecta a RabbitMQ y devuelve el canal
func ConnectRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
	url := GetRabbitMQURL()

	log.Printf("🔌 Intentando conectar a RabbitMQ...")

	// Conectar a RabbitMQ
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("❌ Error de conexión: %v", err)
		return nil, nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	log.Println("✅ Conexión establecida, creando canal...")

	// Crear un canal
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declarar el exchange (por si no existe)
	err = ch.ExchangeDeclare(
		"schedules_exchange", // name
		"topic",              // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Println("✅ RabbitMQ connected successfully")
	return conn, ch, nil
}

// RabbitMQConfig contiene la configuración del publisher
type RabbitMQConfig struct {
	ExchangeName string
	ExchangeType string
	RoutingKey   string
}

// GetSchedulesPublisherConfig retorna la configuración del publisher
func GetSchedulesPublisherConfig() RabbitMQConfig {
	return RabbitMQConfig{
		ExchangeName: "schedules_exchange",
		ExchangeType: "topic",
		RoutingKey:   "schedule", // Se le agrega .create, .update, .delete
	}
}
