package repository

import "time"

// Repository is a software repository containing Go code.
type Repository struct {
	URL         string
	Description string
	Updated     time.Time
}
