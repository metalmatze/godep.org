package repository

import (
	"bufio"
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

var (
	rePackage = regexp.MustCompile(`imports (\d+) packages \(graph\) and is imported by (\d+)`)
	reUpdated = regexp.MustCompile(`Updated (.{10})\.`)
)

type GoDoc struct {
	client *http.Client
}

func NewGoDoc() (*GoDoc, error) {
	return &GoDoc{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (gd *GoDoc) Get(ctx context.Context, urlPath string) (*GoDocInfo, error) {
	defer func(start time.Time) {
		log.Println("godoc", time.Since(start))
	}(time.Now())

	u, err := url.Parse("https://godoc.org")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse base url")
	}
	u.Path = urlPath

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	resp, err := gd.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do the request")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NotFoundErr
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open the godoc document from response")
	}

	s := doc.Find("#x-pkginfo p").First()
	return parseInfo(strings.NewReader(s.Text()))
}

type GoDocInfo struct {
	Imports   int
	Importers int
	Updated   time.Time
}

func parseInfo(r io.Reader) (*GoDocInfo, error) {
	p := &GoDocInfo{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Package") {
			e := rePackage.FindStringSubmatch(line)
			if len(e) == 3 {
				imports, err := strconv.ParseInt(e[1], 10, 64)
				if err != nil {
					return p, err
				}
				importers, err := strconv.ParseInt(e[2], 10, 64)
				if err != nil {
					return p, err
				}
				p.Imports = int(imports)
				p.Importers = int(importers)
			}
		}
		if strings.HasPrefix(line, "Updated") {
			e := reUpdated.FindStringSubmatch(line)
			if len(e) == 2 {
				t, err := time.Parse("2006-01-02", e[1])
				if err != nil {
					return p, err
				}
				p.Updated = t
			}
		}
	}
	return p, nil
}
