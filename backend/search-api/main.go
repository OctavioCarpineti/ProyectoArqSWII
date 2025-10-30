package main

import (
	"log"
	"os"
	"os/signal"
	"search-api/clients"
	"search-api/config"
	"search-api/controllers"
	"search-api/repositories"
	"search-api/services"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 Iniciando Search API...")

	// ==========================================
	// CONFIGURACIÓN
	// ==========================================

	solrURL := config.GetSolrURL()
	memcachedHost := config.GetMemcachedHost()
	rabbitURL := config.GetRabbitMQURL()

	// ==========================================
	// INICIALIZACIÓN DE CLIENTES
	// ==========================================

	log.Println("🌐 Inicializando cliente de activities-api...")
	activitiesClient := clients.NewActivitiesClient()

	// ==========================================
	// INICIALIZACIÓN DE REPOSITORIOS
	// ==========================================

	log.Println("🔍 Inicializando repositorio SolR...")
	solrRepo := repositories.NewSolrRepository(solrURL)

	// Verificar conexión con SolR
	if err := solrRepo.Ping(); err != nil {
		log.Printf("⚠️  Warning: No se pudo conectar a SolR: %v", err)
		log.Println("   Continuando de todas formas...")
	} else {
		log.Println("✅ SolR conectado correctamente")
	}

	log.Println("💾 Inicializando repositorio de caché...")
	cacheRepo := repositories.NewCacheRepository(memcachedHost)

	// ==========================================
	// INICIALIZACIÓN DE RABBITMQ CONSUMER
	// ==========================================

	log.Println("🐰 Inicializando RabbitMQ consumer...")
	consumer, err := repositories.NewRabbitMQConsumer(
		rabbitURL,
		solrRepo,
		cacheRepo,
		activitiesClient,
	)
	if err != nil {
		log.Printf("⚠️  Warning: No se pudo inicializar RabbitMQ consumer: %v", err)
		log.Println("   Search API funcionará sin sincronización automática")
	} else {
		// Iniciar consumer en background
		if err := consumer.Start(); err != nil {
			log.Printf("⚠️  Warning: No se pudo iniciar consumer: %v", err)
		} else {
			log.Println("✅ RabbitMQ consumer iniciado correctamente")

			// Cerrar consumer al terminar
			defer consumer.Close()
		}
	}

	// ==========================================
	// INICIALIZACIÓN DE SERVICIOS
	// ==========================================

	log.Println("⚙️  Inicializando servicios...")
	searchService := services.NewSearchService(solrRepo, cacheRepo)
	log.Println("✅ Servicios inicializados")

	// ==========================================
	// INICIALIZACIÓN DE CONTROLADORES
	// ==========================================

	log.Println("🎮 Inicializando controladores...")
	searchController := controllers.NewSearchController(searchService)
	log.Println("✅ Controladores inicializados")

	// ==========================================
	// CONFIGURACIÓN DE GIN
	// ==========================================

	router := gin.Default()

	// Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// ==========================================
	// RUTAS PÚBLICAS
	// ==========================================

	// Health check
	router.GET("/health", searchController.Health)

	// Búsqueda
	router.GET("/search", searchController.Search)

	// Obtener schedule por ID
	router.GET("/search/:id", searchController.GetScheduleByID)

	// ==========================================
	// MANEJO DE SEÑALES DE CIERRE GRACEFUL
	// ==========================================

	// Canal para recibir señales del sistema
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar servidor en goroutine
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	go func() {
		log.Printf("🚀 Search API iniciando en puerto %s...", port)
		log.Printf("📋 Endpoints disponibles:")
		log.Printf("   GET /search              - Búsqueda de horarios")
		log.Printf("   GET /search/:id          - Obtener horario por ID")
		log.Printf("   GET /health              - Health check")
		log.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")

		if err := router.Run(":" + port); err != nil {
			log.Fatalf("❌ Error al iniciar servidor: %v", err)
		}
	}()

	// Esperar señal de cierre
	<-quit
	log.Println("🛑 Señal de cierre recibida, apagando gracefully...")

	log.Println("✅ Search API cerrado correctamente")
}
