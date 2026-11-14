package types

import "time"

type Venue struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	City      string    `json:"city"`
	Capacity  int       `json:"capacity"`
	CreatedAt time.Time `json:"created_at"`
}
