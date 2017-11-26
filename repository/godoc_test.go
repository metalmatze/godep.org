package repository

import (
	"strings"
	"testing"
	"time"
)

func TestParsePkgInfo(t *testing.T) {
	pkgs := map[string]GoDocInfo{
		"Package sarama imports 30 packages (graph) and is imported by 433 packages.\nUpdated 2017-11-20.\nRefresh now.\nTools for package owners.": {
			Imports:   30,
			Importers: 433,
			Updated:   time.Date(2017, 11, 20, 0, 0, 0, 0, time.UTC),
		},
		"Package jwt imports 17 packages (graph) and is imported by 1343 packages.\nUpdated 2017-11-20.\nRefresh now.\nTools for package owners.": {
			Imports:   17,
			Importers: 1343,
			Updated:   time.Date(2017, 11, 20, 0, 0, 0, 0, time.UTC),
		},
		"Updated 2017-11-22.\nRefresh now.\nTools for package owners.": {
			Updated: time.Date(2017, 11, 22, 0, 0, 0, 0, time.UTC),
		},
		"Package gin imports 32 packages (graph) and is imported by 3869 packages.\nUpdated 2017-11-21.\nRefresh now.\nTools for package owners.": {
			Imports:   32,
			Importers: 3869,
			Updated:   time.Date(2017, 11, 21, 0, 0, 0, 0, time.UTC),
		},
		"Package mysql imports 22 packages (graph) and is imported by 3889 packages.\nUpdated 2017-11-17.\nRefresh now.\nTools for package owners.": {
			Imports:   22,
			Importers: 3889,
			Updated:   time.Date(2017, 11, 17, 0, 0, 0, 0, time.UTC),
		},
		"Updated 2017-11-13.\nRefresh now.\nTools for package owners.": {
			Updated: time.Date(2017, 11, 13, 0, 0, 0, 0, time.UTC),
		},
		"Package websocket imports 22 packages (graph) and is imported by 2952 packages.\nUpdated 2017-11-19.\nRefresh now.\nTools for package owners.": {
			Imports:   22,
			Importers: 2952,
			Updated:   time.Date(2017, 11, 19, 0, 0, 0, 0, time.UTC),
		},
	}

	for s, i := range pkgs {
		info, err := parseInfo(strings.NewReader(s))
		if err != nil {
			t.Error(err)
		}
		if !i.Updated.Equal(info.Updated) || i.Imports != info.Imports || i.Importers != info.Importers {
			t.Errorf("expected pkginfo doesn't match the actual: \n%+v\n%+v\n", i, info)
		}
	}
}

func BenchmarkParseInfo(b *testing.B) {
	in := "Package sarama imports 30 packages (graph) and is imported by 433 packages.\nUpdated 2017-11-20.\nRefresh now.\nTools for package owners."
	for i := 0; i < b.N; i++ {
		parseInfo(strings.NewReader(in))
	}
}
