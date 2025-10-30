package controllers

import (
	"net/http"
	"strconv"
	"users-api/domain"
	"users-api/services"

	"github.com/gin-gonic/gin"
)

// UserController maneja las peticiones HTTP relacionadas con usuarios
type UserController struct {
	userService services.UserService
}

// NewUserController crea una nueva instancia del controlador
func NewUserController(userService services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// CreateUser maneja la creación de un nuevo usuario
// POST /users
func (c *UserController) CreateUser(ctx *gin.Context) {
	var req domain.CreateUserRequest

	// Validar el body de la request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Llamar al servicio
	user, err := c.userService.CreateUser(req)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Internal server error"

		// Mapear errores a códigos HTTP
		switch err {
		case domain.ErrUserAlreadyExists:
			status = http.StatusConflict
			message = err.Error()
		case domain.ErrInvalidEmail, domain.ErrInvalidPassword, domain.ErrInvalidUsername:
			status = http.StatusBadRequest
			message = err.Error()
		}

		ctx.JSON(status, gin.H{"error": message})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

// GetUserByID obtiene un usuario por su ID
// GET /users/:id
func (c *UserController) GetUserByID(ctx *gin.Context) {
	// Obtener ID del parámetro de la URL
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Llamar al servicio
	user, err := c.userService.GetUserByID(uint(id))
	if err != nil {
		if err == domain.ErrUserNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// Login autentica un usuario y devuelve un token JWT
// POST /login
func (c *UserController) Login(ctx *gin.Context) {
	var req domain.LoginRequest

	// Validar el body de la request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Llamar al servicio
	response, err := c.userService.Login(req)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
