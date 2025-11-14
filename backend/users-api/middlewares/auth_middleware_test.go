package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"users-api/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Generar un token válido
	token, err := utils.GenerateToken(1, "testuser", "test@example.com", "normal")
	assert.NoError(t, err)

	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		// Verificar que los claims estén en el context
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, uint(1), userID)

		username, exists := c.Get("username")
		assert.True(t, exists)
		assert.Equal(t, "testuser", username)

		role, exists := c.Get("role")
		assert.True(t, exists)
		assert.Equal(t, "normal", role)

		c.JSON(200, gin.H{"message": "success"})
	})

	// Act
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, 200, w.Code)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Act
	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header is required")
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Act
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid authorization header format")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Act
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestAdminMiddleware_AdminUser(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Generar un token de admin
	token, err := utils.GenerateToken(1, "admin", "admin@example.com", "admin")
	assert.NoError(t, err)

	router.Use(AuthMiddleware())
	router.Use(AdminMiddleware())
	router.GET("/admin-only", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "admin access granted"})
	})

	// Act
	req, _ := http.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, 200, w.Code)
}

func TestAdminMiddleware_NormalUser(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Generar un token de usuario normal
	token, err := utils.GenerateToken(2, "normaluser", "user@example.com", "normal")
	assert.NoError(t, err)

	router.Use(AuthMiddleware())
	router.Use(AdminMiddleware())
	router.GET("/admin-only", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "admin access granted"})
	})

	// Act
	req, _ := http.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "Admin role required")
}
