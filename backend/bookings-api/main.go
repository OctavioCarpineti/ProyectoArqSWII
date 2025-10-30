package main

import (
	"bookings-api/clients"
	"bookings-api/config"
	"bookings-api/controllers"
	"bookings-api/middlewares"
	"bookings-api/repositories"
	"bookings-api/services"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// ==========================================
	// CONFIGURACIÓN DE LA BASE DE DATOS
	// ==========================================

	log.Println("🔌 Conectando a MongoDB...")
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatalf("❌ Error al conectar a MongoDB: %v", err)
	}
	log.Println("✅ MongoDB conectado exitosamente")

	// ==========================================
	// INICIALIZACIÓN DE CLIENTES HTTP
	// ==========================================

	log.Println("🌐 Inicializando clientes HTTP...")
	userClient := clients.NewUserClient()
	activitiesClient := clients.NewActivitiesClient()
	log.Println("✅ Clientes HTTP inicializados")

	// ==========================================
	// INICIALIZACIÓN DE REPOSITORIOS
	// ==========================================

	log.Println("💾 Inicializando repositorios...")
	bookingRepo := repositories.NewBookingMongoRepository(db)
	log.Println("✅ Repositorios inicializados")

	// ==========================================
	// INICIALIZACIÓN DE SERVICIOS
	// ==========================================

	log.Println("⚙️  Inicializando servicios...")
	bookingService := services.NewBookingService(bookingRepo, userClient, activitiesClient)
	log.Println("✅ Servicios inicializados")

	// ==========================================
	// INICIALIZACIÓN DE CONTROLADORES
	// ==========================================

	log.Println("🎮 Inicializando controladores...")
	bookingController := controllers.NewBookingController(bookingService)
	log.Println("✅ Controladores inicializados")

	// ==========================================
	// CONFIGURACIÓN DE GIN
	// ==========================================

	router := gin.Default()

	// Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// ==========================================
	// RUTAS PÚBLICAS
	// ==========================================

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status":  "healthy",
			"service": "bookings-api",
		})
	})

	// ==========================================
	// RUTAS PROTEGIDAS (requieren JWT)
	// ==========================================

	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	{
		// Crear reserva
		protected.POST("/bookings", bookingController.CreateBooking)

		// Obtener reserva por ID
		protected.GET("/bookings/:id", bookingController.GetBookingByID)

		// Obtener reservas por usuario (query param: user_id)
		protected.GET("/bookings", bookingController.GetBookingsByUser)

		// Cancelar reserva
		protected.DELETE("/bookings/:id", bookingController.CancelBooking)
	}

	// ==========================================
	// INICIAR SERVIDOR
	// ==========================================

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("🚀 Bookings API iniciando en puerto %s...", port)
	log.Printf("📋 Endpoints disponibles:")
	log.Printf("   POST   /bookings         - Crear reserva (requiere JWT)")
	log.Printf("   GET    /bookings/:id     - Obtener reserva (requiere JWT)")
	log.Printf("   GET    /bookings?user_id - Listar reservas de usuario (requiere JWT)")
	log.Printf("   DELETE /bookings/:id     - Cancelar reserva (requiere JWT)")
	log.Printf("   GET    /health           - Health check")
	log.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ Error al iniciar servidor: %v", err)
	}
}
