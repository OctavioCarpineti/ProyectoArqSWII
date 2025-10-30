package config

import (
	"log"
	"os"
)

// GetRabbitMQURL obtiene la URL de RabbitMQ desde variables de entorno
func GetRabbitMQURL() string {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	log.Printf("🐰 RabbitMQ URL: %s", url)
	return url
}

// RabbitMQConfig contiene la configuración de colas y exchanges
type RabbitMQConfig struct {
	ExchangeName string
	ExchangeType string
	QueueName    string
	RoutingKey   string
}

// GetSchedulesQueueConfig retorna la configuración para la cola de schedules
func GetSchedulesQueueConfig() RabbitMQConfig {
	return RabbitMQConfig{
		ExchangeName: "schedules_exchange",
		ExchangeType: "topic",
		QueueName:    "schedules_queue",
		RoutingKey:   "schedule.*", // schedule.create, schedule.update, schedule.delete
	}
}
