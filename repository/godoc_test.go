package repository

import (
	"strings"
	"testing"
)

func TestParsePkgInfo(t *testing.T) {
	pkgs := []string{
		"Package sarama imports 30 packages (graph) and is imported by 433 packages.\nUpdated 2017-11-20.\nRefresh now.\nTools for package owners.",
		"Package jwt imports 17 packages (graph) and is imported by 1343 packages.\nUpdated 2017-11-20.\nRefresh now.\nTools for package owners.",
		"Updated 2017-11-22.\nRefresh now.\nTools for package owners.",
		"Package gin imports 32 packages (graph) and is imported by 3869 packages.\nUpdated 2017-11-21.\nRefresh now.\nTools for package owners.",
		"Package mysql imports 22 packages (graph) and is imported by 3889 packages.\nUpdated 2017-11-17.\nRefresh now.\nTools for package owners.",
		"Updated 2017-11-13.\nRefresh now.\nTools for package owners.",
		"Package websocket imports 22 packages (graph) and is imported by 2952 packages.\nUpdated 2017-11-19.\nRefresh now.\nTools for package owners.",
	}

	for _, p := range pkgs {
		parsePkgInfo(strings.NewReader(p))
	}
}
