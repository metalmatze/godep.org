package repository

import (
	"time"
)

type (
	// Repository is a software repository containing Go code.
	Repository struct {
		URL         string
		Description string
		Updated     time.Time

		CurrentVersion Version
		License        License
		Statistics     []Statistic
		Topics         []Topic
		Versions       []Version
	}
	// License of a Repository
	License struct {
		Name string
	}
	// Statistic of a Repository
	Statistic struct {
		Name  string
		Value int
		URL   string
	}
	// Topic describing a Repository
	Topic struct{}
	// Version a Repository was tagged with
	Version struct {
		Name      string
		Published time.Time
	}
)
