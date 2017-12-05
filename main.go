package main

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"
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

	var rs repository.Service
	{
		rs = repository.NewService(repositories, gh, gd)
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
		notFoundTmpl, err := loadTemplates(box, "_layout.html", "404.html")
		if err != nil {
			level.Warn(logger).Log("msg", "failed to load templates", "err", err)
			os.Exit(2)
		}

		homeTmpl, err := loadTemplates(box, "_layout.html", "home.html")
		if err != nil {
			level.Warn(logger).Log("msg", "failed to load templates", "err", err)
			os.Exit(2)
		}

		faqTmpl, err := loadTemplates(box, "_layout.html", "faq.html")
		if err != nil {
			level.Warn(logger).Log("msg", "failed to load templates", "err", err)
			os.Exit(2)
		}

		repositoryTmpl, err := loadTemplates(box, "_layout.html", "repository.html")
		if err != nil {
			level.Warn(logger).Log("msg", "failed to load templates", "err", err)
			os.Exit(2)
		}

		r := chi.NewRouter()
		r.Get("/", homeHandler(rs, homeTmpl))
		r.Get("/faq", faqHandler(faqTmpl))
		r.Get("/main.css", styleHandler(box.Bytes("main.css")))
		r.Get("/github.com/{owner}/{name}", repository.GitHubHandler(rs, repositoryTmpl, notFoundTmpl))
		r.NotFound(notFoundHandler(notFoundTmpl))

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

func loadTemplates(box packr.Box, templates ...string) (*template.Template, error) {
	tmpl := template.New("page")
	tmpl.Funcs(template.FuncMap{
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
	})

	var err error
	for _, t := range templates {
		tmpl, err = tmpl.Parse(box.String(t))
		if err != nil {
			return tmpl, err
		}
	}

	return tmpl, nil
}

func notFoundHandler(tmpl *template.Template) http.HandlerFunc {
	type Page struct{ Title string }

	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		tmpl.ExecuteTemplate(w, "layout", Page{Title: "Not Found"})
	}
}

func homeHandler(rs repository.Service, tmpl *template.Template) http.HandlerFunc {
	type Page struct {
		Title   string
		Popular []string
		Latest  []string
		Random  []string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		homepage, err := rs.Homepage(r.Context())
		if err != nil {
			http.Error(w, "failed to retrieve homepage", http.StatusInternalServerError)
			return
		}

		p := Page{
			Popular: homepage.Popular,
			Latest:  homepage.Latest,
			Random:  homepage.Random,
		}

		if err := tmpl.ExecuteTemplate(w, "layout", p); err != nil {
			http.Error(w, "failed to execute homepage layout", http.StatusInternalServerError)
			return
		}
	}
}

func faqHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "layout", nil)
	}
}

func styleHandler(d []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(d)
	}
}
