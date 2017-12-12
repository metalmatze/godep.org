package repository

import (
	"sort"
	"time"
)

type (
	// Repository is a software repository containing Go code.
	Repository struct {
		URL         string
		Description string
		Updated     time.Time

		License    License
		Statistics []Statistic
		Topics     []Topic
		Versions   []Version
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

func (r Repository) CurrentVersion() Version {
	if len(r.Versions) == 0 {
		return Version{}
	}

	if r.Versions[0].Published.IsZero() {
		return r.Versions[0]
	}

	sort.Slice(r.Versions, func(i, j int) bool {
		return r.Versions[i].Published.After(r.Versions[j].Published)
	})
	return r.Versions[0]
}
