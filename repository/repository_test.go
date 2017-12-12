package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRepository_SetCurrentVersion(t *testing.T) {
	v1 := Version{Name: "v1"}
	v2 := Version{Name: "v2"}

	r := Repository{}
	assert.Equal(t, Version{}, r.CurrentVersion())

	r = Repository{}
	r.Versions = []Version{v1}
	assert.Equal(t, v1, r.CurrentVersion())

	r = Repository{}
	r.Versions = []Version{v1, v2}
	assert.Equal(t, v2, r.CurrentVersion())
	assert.Equal(t, v1, r.Versions[0])
	assert.Equal(t, v2, r.Versions[1])

	v3 := Version{Name: "v3", Published: time.Now().Add(-2 * time.Minute)}
	v4 := Version{Name: "v4", Published: time.Now().Add(-1 * time.Minute)}

	r = Repository{}
	r.Versions = []Version{v3, v4}
	assert.Equal(t, v4, r.CurrentVersion())
	assert.Equal(t, v3, r.Versions[0])
	assert.Equal(t, v4, r.Versions[1])
}
