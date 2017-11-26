package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/gobuffalo/packr"
	_ "github.com/lib/pq"
	"github.com/metalmatze/godep.org/repository"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	box := packr.NewBox("./assets")

	templateFuncs := template.FuncMap{
		"dateFormat": func(format string, t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format(format)
		},
		"repositoryName": func(url string) string {
			s := strings.Split(url, "/")
			return s[len(s)-1]
		},
	}

	page := template.New("page")
	page.Funcs(templateFuncs)
	page, err := page.Parse(box.String("index.html"))
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	gd, err := repository.NewGoDoc()
	if err != nil {
		log.Fatal(err)
	}

	gh, err := repository.NewGitHubClient(os.Getenv("GITHUB_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	var repositories repository.Storage
	{
		repositories = repository.NewPostgresStorage(db)
	}

	var repositoryService repository.Service
	{
		repositoryService = repository.NewService(repositories, gh, gd)
	}

	var g run.Group
	{
		sig := make(chan os.Signal, 2)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		g.Add(func() error {
			<-sig
			return nil
		}, func(err error) {
			close(sig)
		})
	}
	{
		r := chi.NewRouter()
		r.Get("/index.css", styleHandler(box.Bytes("index.css")))
		r.Get("/flexboxgrid.min.css", styleHandler(box.Bytes("flexboxgrid.min.css")))
		r.Get("/godoc.html", styleHandler(box.Bytes("godoc.html")))
		r.Get("/github.com/{owner}/{name}", packageHandler(repositoryService, page))

		s := http.Server{
			Addr:    ":8000",
			Handler: r,
		}

		g.Add(func() error {
			log.Println("starting http server on :8000")
			return s.ListenAndServe()
		}, func(err error) {
			log.Println("shutting down http server on :8000")
			s.Shutdown(context.Background())
		})
	}
	{
		r := chi.NewRouter()
		r.Handle("/metrics", promhttp.Handler())

		s := http.Server{
			Addr:    ":8001",
			Handler: r,
		}

		g.Add(func() error {
			log.Println("starting internal http server on :8001")
			return s.ListenAndServe()
		}, func(err error) {
			log.Println("shutting down internal http server on :8001")
			s.Shutdown(context.Background())
		})
	}

	if err := g.Run(); err != nil {
		log.Fatal(err)
	}
}

func styleHandler(d []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(d)
	}
}

func packageHandler(repositories repository.Service, page *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := chi.URLParam(r, "owner")
		name := chi.URLParam(r, "name")

		uri, err := url.Parse(fmt.Sprintf("github.com/%s/%s", owner, name))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		repo, err := repositories.Get(r.Context(), uri.String())
		if err == repository.NotFoundErr {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := page.Execute(w, repo); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
