package types

import "time"

type Venue struct {
	ID        int64
	Name      string
	City      string
	Capacity  int
	CreatedAt time.Time
}
