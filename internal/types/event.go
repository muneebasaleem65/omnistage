package types

import (
	"time"
)

type Event struct {
	ID          int64     `json:"id"`
	VenueID     int64     `json:"venue_id"`
	Title       string    `json:"title"`
	Artist      string    `json:"artist"`
	Description string    `json:"description"`
	StartsAt    time.Time `json:"starts_at"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
