package repository

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
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

func (gd *GoDoc) Get(ctx context.Context, urlPath string) error {
	defer func(start time.Time) {
		log.Println("godoc", time.Since(start))
	}(time.Now())

	u, err := url.Parse("https://godoc.org")
	if err != nil {
		return errors.Wrap(err, "failed to parse base url")
	}
	u.Path = urlPath

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	resp, err := gd.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to do the request")
	}

	if resp.StatusCode != http.StatusOK {
		return NotFoundErr
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return errors.Wrap(err, "failed to open the godoc document from response")
	}

	s := doc.Find("#x-pkginfo p").First()
	fmt.Println(strings.TrimSpace(s.Text()))

	return nil
}

type PkgInfo struct {
	Imports  int
	Imported int
	Updated  time.Time
}

func parsePkgInfo(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Package") {
			println(line)
		} else {
			println(line)
		}
	}
	return nil
}
