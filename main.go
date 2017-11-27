package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/gobuffalo/packr"
	_ "github.com/lib/pq"
	"github.com/metalmatze/godep.org/repository"
	"github.com/oklog/run"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	config := struct {
		DSN         string
		GithubToken string
	}{
		DSN:         os.Getenv("DSN"),
		GithubToken: os.Getenv("GITHUB_TOKEN"),
	}

	if config.DSN == "" {
		config.DSN = "postgres://postgres:postgres@localhost:5432?sslmode=disable"
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.WithPrefix(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

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
		logger.Log("msg", "failed to parse index.html template", "err", err)
		os.Exit(2)
	}

	db, err := sql.Open("postgres", config.DSN)
	if err != nil {
		logger.Log("msg", "failed to open sql connection to postgres", "err", err)
		os.Exit(2)
	}
	defer db.Close()

	apiCalls := prometheus.NewHistogramFrom(prom.HistogramOpts{
		Namespace: "godep",
		Name:      "api_calls",
		Help:      "API calls made to other services",
		Buckets:   []float64{.025, .05, .075, .1, .2, .3, .4, .5, .6, .7, .8, .9, 1, 1.5, 2, 3, 4, 5},
	}, []string{"service"})

	gd, err := repository.NewGoDoc(apiCalls)
	if err != nil {
		logger.Log("msg", "failed to create godoc client", "err", err)
		os.Exit(2)
	}

	gh, err := repository.NewGitHubClient(config.GithubToken, apiCalls)
	if err != nil {
		logger.Log("msg", "failed to create github client", "err", err)
		os.Exit(2)
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
		r.Get("/github.com/{owner}/{name}", packageHandler(repositoryService, page))

		s := http.Server{
			Addr:    ":8000",
			Handler: r,
		}

		g.Add(func() error {
			level.Info(logger).Log("msg", "starting http server on :8000")
			return s.ListenAndServe()
		}, func(err error) {
			level.Info(logger).Log("msg", "shutting down http server on :8000")
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
			level.Info(logger).Log("msg", "starting internal http server on :8001")
			return s.ListenAndServe()
		}, func(err error) {
			level.Info(logger).Log("msg", "shutting down internal http server on :8001")
			s.Shutdown(context.Background())
		})
	}

	if err := g.Run(); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
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
