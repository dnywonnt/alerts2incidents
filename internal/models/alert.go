package models // dnywonnt.me/alerts2incidents/internal/models

import "time"

// Alert represents the data structure for an alert.
type Alert struct {
	Summary   string    `json:"summary"`    // Brief description of the alert
	CreatedAt time.Time `json:"created_at"` // Time when the alert was created
}
