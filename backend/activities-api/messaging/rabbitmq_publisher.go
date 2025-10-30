package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"activities-api/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher publica mensajes a RabbitMQ
type RabbitMQPublisher struct {
	channel *amqp.Channel
}

// NewRabbitMQPublisher crea una nueva instancia del publicador
func NewRabbitMQPublisher(channel *amqp.Channel) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		channel: channel,
	}
}

// PublishScheduleEvent publica un evento de schedule
func (p *RabbitMQPublisher) PublishScheduleEvent(event domain.ScheduleEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Serializar el evento a JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Determinar routing key basado en la operación
	routingKey := fmt.Sprintf("schedule.%s", event.Operation)

	// Publicar el mensaje
	err = p.channel.PublishWithContext(
		ctx,
		"schedules_exchange", // exchange
		routingKey,           // routing key
		false,                // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
