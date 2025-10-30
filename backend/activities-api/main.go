package main

import (
	"activities-api/clients"
	"activities-api/config"
	"activities-api/controllers"
	"activities-api/messaging"
	"activities-api/middlewares"
	"activities-api/repositories"
	"activities-api/services"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Conectar a MongoDB
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Conectar a RabbitMQ
	rabbitConn, rabbitChannel, err := config.ConnectRabbitMQ()
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	// Inicializar clientes
	userClient := clients.NewUserClient()

	// Inicializar RabbitMQ Publisher
	publisher := messaging.NewRabbitMQPublisher(rabbitChannel)

	// Inicializar repositorios
	activityRepo := repositories.NewActivityMongoRepository(db)
	scheduleRepo := repositories.NewScheduleMongoRepository(db)

	// Inicializar servicios
	activityService := services.NewActivityService(activityRepo, scheduleRepo, userClient)
	scheduleService := services.NewScheduleService(scheduleRepo, activityRepo, userClient, publisher)

	// Inicializar controladores
	activityController := controllers.NewActivityController(activityService)
	scheduleController := controllers.NewScheduleController(scheduleService)

	// Configurar Gin router
	router := gin.Default()

	// Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "activities-api",
		})
	})

	// ============================================
	// RUTAS PÚBLICAS
	// ============================================

	router.GET("/activities", activityController.GetAllActivities)
	router.GET("/activities/:id", activityController.GetActivityByID)
	router.GET("/activities/:id/schedules", scheduleController.GetSchedulesByActivityID)

	router.GET("/schedules", scheduleController.GetAllSchedules)
	router.GET("/schedules/:id", scheduleController.GetScheduleByID)

	// ============================================
	// RUTAS PROTEGIDAS
	// ============================================

	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	{
		protected.POST("/activities", middlewares.AdminOnly(), activityController.CreateActivity)
		protected.PUT("/activities/:id", activityController.UpdateActivity)
		protected.DELETE("/activities/:id", activityController.DeleteActivity)

		protected.POST("/activities/:id/schedules", scheduleController.CreateSchedule)
		protected.PUT("/schedules/:id", scheduleController.UpdateSchedule)
		protected.DELETE("/schedules/:id", scheduleController.DeleteSchedule)

		protected.PUT("/schedules/:id/bookings", scheduleController.UpdateBookings)
	}

	// Iniciar servidor
	port := "8081"
	log.Printf("🚀 Activities API running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
