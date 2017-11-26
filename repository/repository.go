package repository

import (
	"time"
)

// Repository is a software repository containing Go code.
type (
	Repository struct {
		URL         string
		Description string
		Updated     time.Time

		CurrentVersion Version
		License        License
		Stats          []Stat
		Topics         []Topic
		Versions       []Version
	}
	License struct {
		Name string
	}
	Stat struct {
		Name  string
		Value int
		URL   string
	}
	Topic   struct{}
	Version struct {
		Name      string
		Published time.Time
	}
)
