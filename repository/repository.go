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

// CurrentVersion returns the latest and current version of all versions
func (r Repository) CurrentVersion() Version {
	if len(r.Versions) == 0 {
		return Version{}
	}

	if r.Versions[len(r.Versions)-1].Published.IsZero() {
		return r.Versions[len(r.Versions)-1]
	}

	// Copy the original slice to not sort in place
	versions := make([]Version, len(r.Versions))
	copy(versions, r.Versions)

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Published.After(versions[j].Published)
	})
	return versions[0]
}
