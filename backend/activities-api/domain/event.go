package domain

import "time"

// ScheduleEvent representa un evento de schedule para RabbitMQ
type ScheduleEvent struct {
	Operation  string `json:"operation"` // "CREATE", "UPDATE", "DELETE"
	ScheduleID string `json:"schedule_id"`
	ActivityID string `json:"activity_id"`
	Timestamp  int64  `json:"timestamp"`
}

// NewScheduleEvent crea un nuevo evento
func NewScheduleEvent(operation, scheduleID, activityID string) ScheduleEvent {
	return ScheduleEvent{
		Operation:  operation,
		ScheduleID: scheduleID,
		ActivityID: activityID,
		Timestamp:  time.Now().Unix(),
	}
}
