package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ActivitiesClient es el cliente para comunicarse con activities-api
type ActivitiesClient struct {
	baseURL    string
	httpClient *http.Client
}

// ScheduleResponse representa la respuesta de activities-api para un schedule
type ScheduleResponse struct {
	ID              string `json:"id"`
	ActivityID      string `json:"activity_id"`
	Instructor      string `json:"instructor"`
	DayOfWeek       string `json:"day_of_week"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Location        string `json:"location"`
	MaxCapacity     int    `json:"max_capacity"`
	CurrentBookings int    `json:"current_bookings"`
	AvailableSpots  int    `json:"available_spots"`
	Status          string `json:"status"`
}

// ActivityResponse representa la respuesta de activities-api para una activity
type ActivityResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
}

// NewActivitiesClient crea una nueva instancia del cliente
func NewActivitiesClient() *ActivitiesClient {
	baseURL := os.Getenv("ACTIVITIES_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	return &ActivitiesClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetScheduleByID obtiene un schedule por su ID
func (c *ActivitiesClient) GetScheduleByID(scheduleID string) (*ScheduleResponse, error) {
	url := fmt.Sprintf("%s/schedules/%s", c.baseURL, scheduleID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call activities-api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("schedule not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("activities-api returned status %d: %s", resp.StatusCode, string(body))
	}

	var schedule ScheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &schedule, nil
}

// GetActivityByID obtiene una activity por su ID
func (c *ActivitiesClient) GetActivityByID(activityID string) (*ActivityResponse, error) {
	url := fmt.Sprintf("%s/activities/%s", c.baseURL, activityID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call activities-api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("activity not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("activities-api returned status %d: %s", resp.StatusCode, string(body))
	}

	var activity ActivityResponse
	if err := json.NewDecoder(resp.Body).Decode(&activity); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &activity, nil
}

// UpdateScheduleBookings actualiza el contador de reservas de un schedule
func (c *ActivitiesClient) UpdateScheduleBookings(scheduleID string, increment bool, token string) error {
	url := fmt.Sprintf("%s/schedules/%s/bookings", c.baseURL, scheduleID)

	body := map[string]bool{"increment": increment}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update schedule bookings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("activities-api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
