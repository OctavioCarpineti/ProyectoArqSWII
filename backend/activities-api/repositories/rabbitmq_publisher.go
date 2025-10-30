package repositories

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ScheduleEvent representa el evento que se publica
type ScheduleEvent struct {
	Operation  string `json:"operation"`   // CREATE, UPDATE, DELETE
	ScheduleID string `json:"schedule_id"` // ID del schedule
	ActivityID string `json:"activity_id"` // ID de la actividad
	Timestamp  int64  `json:"timestamp"`   // Unix timestamp
}

// RabbitMQPublisher publica eventos a RabbitMQ
type RabbitMQPublisher struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	exchangeName string
	routingKey   string
}

// NewRabbitMQPublisher crea una nueva instancia del publisher
func NewRabbitMQPublisher(rabbitURL string, exchangeName string, exchangeType string, routingKey string) (*RabbitMQPublisher, error) {
	// Conectar a RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Crear canal
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declarar exchange
	err = channel.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Printf("✅ RabbitMQ publisher inicializado")
	log.Printf("   Exchange: %s", exchangeName)
	log.Printf("   Type: %s", exchangeType)

	return &RabbitMQPublisher{
		conn:         conn,
		channel:      channel,
		exchangeName: exchangeName,
		routingKey:   routingKey,
	}, nil
}

// PublishScheduleCreated publica evento de creación de schedule
func (p *RabbitMQPublisher) PublishScheduleCreated(scheduleID string, activityID string) error {
	event := ScheduleEvent{
		Operation:  "CREATE",
		ScheduleID: scheduleID,
		ActivityID: activityID,
		Timestamp:  time.Now().Unix(),
	}
	return p.publish(event, "schedule.create")
}

// PublishScheduleUpdated publica evento de actualización de schedule
func (p *RabbitMQPublisher) PublishScheduleUpdated(scheduleID string, activityID string) error {
	event := ScheduleEvent{
		Operation:  "UPDATE",
		ScheduleID: scheduleID,
		ActivityID: activityID,
		Timestamp:  time.Now().Unix(),
	}
	return p.publish(event, "schedule.update")
}

// PublishScheduleDeleted publica evento de eliminación de schedule
func (p *RabbitMQPublisher) PublishScheduleDeleted(scheduleID string, activityID string) error {
	event := ScheduleEvent{
		Operation:  "DELETE",
		ScheduleID: scheduleID,
		ActivityID: activityID,
		Timestamp:  time.Now().Unix(),
	}
	return p.publish(event, "schedule.delete")
}

// publish publica un evento a RabbitMQ
func (p *RabbitMQPublisher) publish(event ScheduleEvent, routingKey string) error {
	// Serializar evento
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publicar mensaje
	err = p.channel.Publish(
		p.exchangeName, // exchange
		routingKey,     // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("📤 Evento publicado: %s para schedule %s", event.Operation, event.ScheduleID)
	return nil
}

// Close cierra la conexión con RabbitMQ
func (p *RabbitMQPublisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	log.Println("🐰 RabbitMQ publisher cerrado")
}
