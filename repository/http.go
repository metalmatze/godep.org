package repository

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
)

func GithubHandler(repositories Service, tmpl *template.Template, notfoundTmpl *template.Template) http.HandlerFunc {
	type Page struct {
		Title      string
		Repository Repository
	}

	return func(w http.ResponseWriter, r *http.Request) {
		owner := chi.URLParam(r, "owner")
		name := chi.URLParam(r, "name")

		uri, err := url.Parse(fmt.Sprintf("github.com/%s/%s", owner, name))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		repo, err := repositories.Get(r.Context(), uri.String())
		if err == NotFoundErr {
			w.WriteHeader(http.StatusNotFound)
			notfoundTmpl.ExecuteTemplate(w, "layout", nil)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p := Page{
			Title:      fmt.Sprintf("%s - ", name),
			Repository: repo,
		}

		if err := tmpl.ExecuteTemplate(w, "layout", p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
