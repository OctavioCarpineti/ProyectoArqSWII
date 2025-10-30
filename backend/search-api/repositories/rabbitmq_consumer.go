package repositories

import (
	"encoding/json"
	"fmt"
	"log"
	"search-api/clients"
	"search-api/config"
	"search-api/domain"
	"time"

	"github.com/streadway/amqp"
)

// ScheduleEvent representa un evento de schedule publicado por activities-api
type ScheduleEvent struct {
	Operation  string `json:"operation"`   // CREATE, UPDATE, DELETE
	ScheduleID string `json:"schedule_id"` // ID del schedule
	ActivityID string `json:"activity_id"` // ID de la actividad
	Timestamp  int64  `json:"timestamp"`   // Unix timestamp
}

// RabbitMQConsumer consume mensajes de RabbitMQ y sincroniza con SolR
type RabbitMQConsumer struct {
	conn             *amqp.Connection
	channel          *amqp.Channel
	solrRepo         SolrRepository
	cacheRepo        CacheRepository
	activitiesClient *clients.ActivitiesClient
	config           config.RabbitMQConfig
}

// NewRabbitMQConsumer crea una nueva instancia del consumidor
func NewRabbitMQConsumer(
	rabbitURL string,
	solrRepo SolrRepository,
	cacheRepo CacheRepository,
	activitiesClient *clients.ActivitiesClient,
) (*RabbitMQConsumer, error) {
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

	queueConfig := config.GetSchedulesQueueConfig()

	// Declarar exchange
	err = channel.ExchangeDeclare(
		queueConfig.ExchangeName, // name
		queueConfig.ExchangeType, // type
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declarar cola
	queue, err := channel.QueueDeclare(
		queueConfig.QueueName, // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		queue.Name,               // queue name
		queueConfig.RoutingKey,   // routing key
		queueConfig.ExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Printf("✅ RabbitMQ consumer inicializado correctamente")
	log.Printf("   Exchange: %s", queueConfig.ExchangeName)
	log.Printf("   Queue: %s", queueConfig.QueueName)
	log.Printf("   Routing Key: %s", queueConfig.RoutingKey)

	return &RabbitMQConsumer{
		conn:             conn,
		channel:          channel,
		solrRepo:         solrRepo,
		cacheRepo:        cacheRepo,
		activitiesClient: activitiesClient,
		config:           queueConfig,
	}, nil
}

// Start inicia el consumo de mensajes en una goroutine
func (c *RabbitMQConsumer) Start() error {
	// Configurar QoS
	err := c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Consumir mensajes
	msgs, err := c.channel.Consume(
		c.config.QueueName, // queue
		"search-api",       // consumer tag
		false,              // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Println("🐰 RabbitMQ consumer esperando mensajes...")

	// Procesar mensajes en goroutine
	go func() {
		for msg := range msgs {
			if err := c.processMessage(msg); err != nil {
				log.Printf("❌ Error procesando mensaje: %v", err)
				msg.Nack(false, true) // Requeue message
			} else {
				msg.Ack(false) // Acknowledge message
			}
		}
	}()

	return nil
}

// processMessage procesa un mensaje de RabbitMQ
func (c *RabbitMQConsumer) processMessage(msg amqp.Delivery) error {
	log.Printf("📨 Mensaje recibido: %s", string(msg.Body))

	// Parsear evento
	var event ScheduleEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	log.Printf("🔄 Procesando evento: %s para schedule %s", event.Operation, event.ScheduleID)

	switch event.Operation {
	case "CREATE", "UPDATE":
		return c.handleCreateOrUpdate(event)
	case "DELETE":
		return c.handleDelete(event)
	default:
		log.Printf("⚠️  Operación desconocida: %s", event.Operation)
		return nil
	}
}

// handleCreateOrUpdate maneja eventos de creación o actualización
func (c *RabbitMQConsumer) handleCreateOrUpdate(event ScheduleEvent) error {
	// 1. Obtener schedule completo desde activities-api
	log.Printf("📥 Obteniendo schedule %s desde activities-api...", event.ScheduleID)
	schedule, err := c.activitiesClient.GetScheduleByID(event.ScheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	// 2. Obtener activity completa
	log.Printf("📥 Obteniendo activity %s desde activities-api...", event.ActivityID)
	activity, err := c.activitiesClient.GetActivityByID(event.ActivityID)
	if err != nil {
		return fmt.Errorf("failed to get activity: %w", err)
	}

	// 3. Crear documento para SolR
	scheduleSearch := &domain.ScheduleSearch{
		ID:               schedule.ID,
		ScheduleID:       schedule.ID,
		ActivityID:       schedule.ActivityID,
		ActivityName:     activity.Name,
		ActivityCategory: activity.Category,
		ActivityPrice:    activity.Price,
		Instructor:       schedule.Instructor,
		DayOfWeek:        schedule.DayOfWeek,
		StartTime:        schedule.StartTime,
		EndTime:          schedule.EndTime,
		Location:         schedule.Location,
		MaxCapacity:      schedule.MaxCapacity,
		CurrentBookings:  schedule.CurrentBookings,
		AvailableSpots:   schedule.AvailableSpots,
		Status:           schedule.Status,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 4. Indexar en SolR
	log.Printf("📇 Indexando en SolR...")
	if err := c.solrRepo.Index(scheduleSearch); err != nil {
		return fmt.Errorf("failed to index in solr: %w", err)
	}

	// 5. Invalidar caché relacionada
	log.Printf("🗑️  Invalidando caché para schedule %s...", event.ScheduleID)
	c.cacheRepo.InvalidatePattern(fmt.Sprintf("*schedule_id=%s*", event.ScheduleID))
	c.cacheRepo.InvalidatePattern(fmt.Sprintf("*activity_id=%s*", event.ActivityID))

	log.Printf("✅ Schedule %s indexado exitosamente", event.ScheduleID)
	return nil
}

// handleDelete maneja eventos de eliminación
func (c *RabbitMQConsumer) handleDelete(event ScheduleEvent) error {
	// 1. Eliminar de SolR
	log.Printf("🗑️  Eliminando schedule %s de SolR...", event.ScheduleID)
	if err := c.solrRepo.Delete(event.ScheduleID); err != nil {
		return fmt.Errorf("failed to delete from solr: %w", err)
	}

	// 2. Invalidar caché
	log.Printf("🗑️  Invalidando caché para schedule %s...", event.ScheduleID)
	c.cacheRepo.InvalidatePattern(fmt.Sprintf("*schedule_id=%s*", event.ScheduleID))

	log.Printf("✅ Schedule %s eliminado exitosamente", event.ScheduleID)
	return nil
}

// Close cierra la conexión con RabbitMQ
func (c *RabbitMQConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	log.Println("🐰 RabbitMQ consumer cerrado")
}
