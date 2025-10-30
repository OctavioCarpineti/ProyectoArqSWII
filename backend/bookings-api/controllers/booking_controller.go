package controllers

import (
	"bookings-api/domain"
	"bookings-api/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BookingController maneja las peticiones HTTP para bookings
type BookingController struct {
	service services.BookingService
}

// NewBookingController crea una nueva instancia del controller
func NewBookingController(service services.BookingService) *BookingController {
	return &BookingController{
		service: service,
	}
}

// CreateBooking maneja POST /bookings
func (c *BookingController) CreateBooking(ctx *gin.Context) {
	var req domain.CreateBookingRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Error al parsear request: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Obtener el token del header para pasarlo al service
	token := ctx.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Validar que el user_id del JWT coincide con el del request
	userIDFromToken, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if req.UserID != userIDFromToken.(uint) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Cannot create booking for another user"})
		return
	}

	log.Printf("📥 Request: Crear reserva para user_id=%d, schedule_id=%s", req.UserID, req.ScheduleID)

	response, err := c.service.CreateBooking(req, token)
	if err != nil {
		log.Printf("❌ Error al crear reserva: %v", err)

		// Mapear errores a códigos HTTP apropiados
		switch err {
		case domain.ErrInvalidUser:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user"})
		case domain.ErrInvalidSchedule:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule"})
		case domain.ErrScheduleFull:
			ctx.JSON(http.StatusConflict, gin.H{"error": "Schedule is full, no available spots"})
		case domain.ErrBookingAlreadyExists:
			ctx.JSON(http.StatusConflict, gin.H{"error": "You already have a booking for this schedule"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		}
		return
	}

	log.Printf("✅ Reserva creada exitosamente: %s", response.ID)
	ctx.JSON(http.StatusCreated, response)
}

// GetBookingByID maneja GET /bookings/:id
func (c *BookingController) GetBookingByID(ctx *gin.Context) {
	id := ctx.Param("id")

	response, err := c.service.GetBookingByID(id)
	if err != nil {
		if err == domain.ErrBookingNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get booking"})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// GetBookingsByUser maneja GET /bookings?user_id=X
func (c *BookingController) GetBookingsByUser(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id query parameter required"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	// Verificar que el usuario solo pueda ver sus propias reservas (a menos que sea admin)
	userIDFromToken, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userRole, _ := ctx.Get("role")
	if uint(userID) != userIDFromToken.(uint) && userRole != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Cannot view bookings of another user"})
		return
	}

	responses, err := c.service.GetBookingsByUser(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bookings"})
		return
	}

	ctx.JSON(http.StatusOK, responses)
}

// CancelBooking maneja DELETE /bookings/:id
func (c *BookingController) CancelBooking(ctx *gin.Context) {
	id := ctx.Param("id")

	// Obtener datos del JWT
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userRole, _ := ctx.Get("role")
	role := ""
	if userRole != nil {
		role = userRole.(string)
	}

	// Obtener el token para pasarlo al service
	token := ctx.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := c.service.CancelBooking(id, userID.(uint), role, token)
	if err != nil {
		if err == domain.ErrBookingNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}
		if err == domain.ErrUnauthorized {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to cancel this booking"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
}
