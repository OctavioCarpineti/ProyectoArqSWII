package controllers

import (
	"activities-api/domain"
	"activities-api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ScheduleController maneja las peticiones HTTP relacionadas con horarios
type ScheduleController struct {
	scheduleService services.ScheduleService
}

// NewScheduleController crea una nueva instancia del controlador
func NewScheduleController(scheduleService services.ScheduleService) *ScheduleController {
	return &ScheduleController{
		scheduleService: scheduleService,
	}
}

// CreateSchedule maneja la creación de un nuevo horario
// POST /activities/:activity_id/schedules
func (c *ScheduleController) CreateSchedule(ctx *gin.Context) {
	activityID := ctx.Param("id")

	var req domain.CreateScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	schedule, err := c.scheduleService.CreateSchedule(activityID, req)
	if err != nil {
		status := http.StatusInternalServerError
		message := err.Error()

		switch err {
		case domain.ErrActivityNotFound:
			status = http.StatusNotFound
		case domain.ErrScheduleConflict:
			status = http.StatusConflict
		case domain.ErrInvalidTimeFormat:
			status = http.StatusBadRequest
		}

		ctx.JSON(status, gin.H{"error": message})
		return
	}

	ctx.JSON(http.StatusCreated, schedule)
}

// GetScheduleByID obtiene un horario por su ID
// GET /schedules/:id
func (c *ScheduleController) GetScheduleByID(ctx *gin.Context) {
	id := ctx.Param("id")

	schedule, err := c.scheduleService.GetScheduleByID(id)
	if err != nil {
		if err == domain.ErrScheduleNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

// GetSchedulesByActivityID obtiene todos los horarios de una actividad
// GET /activities/:activity_id/schedules
func (c *ScheduleController) GetSchedulesByActivityID(ctx *gin.Context) {
	activityID := ctx.Param("id")

	schedules, err := c.scheduleService.GetSchedulesByActivityID(activityID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, schedules)
}

// GetAllSchedules obtiene todos los horarios
// GET /schedules
func (c *ScheduleController) GetAllSchedules(ctx *gin.Context) {
	schedules, err := c.scheduleService.GetAllSchedules()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, schedules)
}

// UpdateSchedule actualiza un horario
// PUT /schedules/:id
func (c *ScheduleController) UpdateSchedule(ctx *gin.Context) {
	id := ctx.Param("id")

	userID, _ := ctx.Get("user_id")
	userRole, _ := ctx.Get("role")

	var req domain.UpdateScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	schedule, err := c.scheduleService.UpdateSchedule(id, req, userID.(uint), userRole.(string))
	if err != nil {
		status := http.StatusInternalServerError
		message := err.Error()

		switch err {
		case domain.ErrScheduleNotFound:
			status = http.StatusNotFound
		case domain.ErrUnauthorized:
			status = http.StatusForbidden
		case domain.ErrScheduleConflict:
			status = http.StatusConflict
		}

		ctx.JSON(status, gin.H{"error": message})
		return
	}

	ctx.JSON(http.StatusOK, schedule)
}

// DeleteSchedule elimina un horario
// DELETE /schedules/:id
func (c *ScheduleController) DeleteSchedule(ctx *gin.Context) {
	id := ctx.Param("id")

	userID, _ := ctx.Get("user_id")
	userRole, _ := ctx.Get("role")

	err := c.scheduleService.DeleteSchedule(id, userID.(uint), userRole.(string))
	if err != nil {
		status := http.StatusInternalServerError
		message := err.Error()

		switch err {
		case domain.ErrScheduleNotFound:
			status = http.StatusNotFound
		case domain.ErrUnauthorized:
			status = http.StatusForbidden
		}

		ctx.JSON(status, gin.H{"error": message})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Schedule deleted successfully"})
}

// UpdateBookings actualiza el contador de reservas (endpoint para bookings-api)
// PUT /schedules/:id/bookings
func (c *ScheduleController) UpdateBookings(ctx *gin.Context) {
	id := ctx.Param("id")

	var req struct {
		Increment bool `json:"increment"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := c.scheduleService.UpdateCurrentBookings(id, req.Increment)
	if err != nil {
		if err == domain.ErrScheduleNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Bookings updated successfully"})
}
