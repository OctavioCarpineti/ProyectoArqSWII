package middlewares

import (
	"net/http"
	"strings"
	"users-api/domain"
	"users-api/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware valida el token JWT en cada request
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Obtener el token del header Authorization
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			ctx.Abort()
			return
		}

		// El formato debe ser: "Bearer {token}"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			ctx.Abort()
			return
		}

		tokenString := parts[1]

		// Validar el token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": domain.ErrInvalidToken.Error()})
			ctx.Abort()
			return
		}

		// Guardar los claims en el contexto para usarlos en los handlers
		ctx.Set("user_id", claims.UserID)
		ctx.Set("username", claims.Username)
		ctx.Set("email", claims.Email)
		ctx.Set("role", claims.Role)

		ctx.Next()
	}
}

// AdminOnly middleware que verifica que el usuario sea admin
func AdminOnly() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role, exists := ctx.Get("role")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			ctx.Abort()
			return
		}

		if role != "admin" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
