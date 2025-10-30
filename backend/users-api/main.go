package main

import (
	"log"
	"users-api/config"
	"users-api/controllers"
	"users-api/middlewares"
	"users-api/repositories"
	"users-api/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Conectar a la base de datos
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Inicializar capas (patrón de inyección de dependencias)
	userRepo := repositories.NewUserMySQLRepository(db)
	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	// Configurar Gin router
	router := gin.Default()

	// Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Frontend
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "users-api",
		})
	})

	// Rutas públicas (sin autenticación)
	router.POST("/users", userController.CreateUser)
	router.POST("/login", userController.Login)
	router.GET("/users/:id", userController.GetUserByID)

	// Rutas protegidas (requieren JWT)
	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	{
		// Aquí puedes agregar más rutas protegidas
	}

	// Rutas solo para admins
	admin := router.Group("/admin")
	admin.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
	{
		// Aquí puedes agregar rutas exclusivas para admins
		admin.GET("/users", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin only endpoint"})
		})
	}

	// Iniciar servidor
	port := "8080"
	log.Printf("🚀 Users API running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
