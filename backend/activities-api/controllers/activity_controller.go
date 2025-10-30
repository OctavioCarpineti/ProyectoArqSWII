package controllers

import (
	"activities-api/domain"
	"activities-api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ActivityController maneja las peticiones HTTP relacionadas con actividades
type ActivityController struct {
	activityService services.ActivityService
}

// NewActivityController crea una nueva instancia del controlador
func NewActivityController(activityService services.ActivityService) *ActivityController {
	return &ActivityController{
		activityService: activityService,
	}
}

// CreateActivity maneja la creación de una nueva actividad
// POST /activities
func (c *ActivityController) CreateActivity(ctx *gin.Context) {
	var req domain.CreateActivityRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	activity, err := c.activityService.CreateActivity(req)
	if err != nil {
		status := http.StatusInternalServerError
		message := err.Error()

		switch err {
		case domain.ErrInvalidOwner:
			status = http.StatusBadRequest
		case domain.ErrInvalidCategory:
			status = http.StatusBadRequest
		}

		ctx.JSON(status, gin.H{"error": message})
		return
	}

	ctx.JSON(http.StatusCreated, activity)
}

// GetActivityByID obtiene una actividad por su ID
// GET /activities/:id
func (c *ActivityController) GetActivityByID(ctx *gin.Context) {
	id := ctx.Param("id")

	activity, err := c.activityService.GetActivityByID(id)
	if err != nil {
		if err == domain.ErrActivityNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Activity not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, activity)
}

// GetAllActivities obtiene todas las actividades
// GET /activities
func (c *ActivityController) GetAllActivities(ctx *gin.Context) {
	activities, err := c.activityService.GetAllActivities()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, activities)
}

// UpdateActivity actualiza una actividad
// PUT /activities/:id
func (c *ActivityController) UpdateActivity(ctx *gin.Context) {
	id := ctx.Param("id")

	// Obtener user_id y role del contexto (inyectados por middleware)
	userID, _ := ctx.Get("user_id")
	userRole, _ := ctx.Get("role")

	var req domain.UpdateActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	activity, err := c.activityService.UpdateActivity(id, req, userID.(uint), userRole.(string))
	if err != nil {
		status := http.StatusInternalServerError
		message := err.Error()

		switch err {
		case domain.ErrActivityNotFound:
			status = http.StatusNotFound
		case domain.ErrUnauthorized:
			status = http.StatusForbidden
		case domain.ErrInvalidCategory:
			status = http.StatusBadRequest
		}

		ctx.JSON(status, gin.H{"error": message})
		return
	}

	ctx.JSON(http.StatusOK, activity)
}

// DeleteActivity elimina una actividad
// DELETE /activities/:id
func (c *ActivityController) DeleteActivity(ctx *gin.Context) {
	id := ctx.Param("id")

	userID, _ := ctx.Get("user_id")
	userRole, _ := ctx.Get("role")

	err := c.activityService.DeleteActivity(id, userID.(uint), userRole.(string))
	if err != nil {
		status := http.StatusInternalServerError
		message := err.Error()

		switch err {
		case domain.ErrActivityNotFound:
			status = http.StatusNotFound
		case domain.ErrUnauthorized:
			status = http.StatusForbidden
		}

		ctx.JSON(status, gin.H{"error": message})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Activity deleted successfully"})
}
