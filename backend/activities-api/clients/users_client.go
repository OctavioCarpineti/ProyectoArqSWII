package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// UserClient es el cliente para comunicarse con users-api
type UserClient struct {
	baseURL    string
	httpClient *http.Client
}

// UserResponse representa la respuesta de users-api
type UserResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// NewUserClient crea una nueva instancia del cliente
func NewUserClient() *UserClient {
	baseURL := os.Getenv("USERS_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &UserClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetUserByID obtiene un usuario por su ID desde users-api
func (c *UserClient) GetUserByID(userID uint) (*UserResponse, error) {
	url := fmt.Sprintf("%s/users/%d", c.baseURL, userID)

	log.Printf("🔍 DEBUG: Llamando a users-api: %s", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Printf("❌ Error al llamar a users-api: %v", err)
		return nil, fmt.Errorf("failed to call users-api: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("📡 Respuesta de users-api: Status %d", resp.StatusCode)

	if resp.StatusCode == http.StatusNotFound {
		log.Println("⚠️  Usuario no encontrado en users-api")
		return nil, fmt.Errorf("user not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ users-api devolvió error: %s", string(body))
		return nil, fmt.Errorf("users-api returned status %d: %s", resp.StatusCode, string(body))
	}

	var user UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Printf("❌ Error al decodificar respuesta: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("✅ Usuario obtenido: ID=%d, Username=%s", user.ID, user.Username)
	return &user, nil
}

// ValidateUser valida que un usuario existe
func (c *UserClient) ValidateUser(userID uint) error {
	_, err := c.GetUserByID(userID)
	return err
}
