package clients

import (
	"encoding/json"
	"fmt"
	"io"
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

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call users-api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("users-api returned status %d: %s", resp.StatusCode, string(body))
	}

	var user UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

// ValidateUser valida que un usuario existe
func (c *UserClient) ValidateUser(userID uint) error {
	_, err := c.GetUserByID(userID)
	return err
}
