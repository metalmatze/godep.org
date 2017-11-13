package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/go-chi/chi"
	"github.com/gobuffalo/packr"
	"github.com/shurcooL/githubql"
	"golang.org/x/oauth2"
)

func main() {
	box := packr.NewBox("./assets")

	templateFuncs := template.FuncMap{
		"dateFormat": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("on Jan 02, 2006")
		},
	}

	page := template.New("page")
	page.Funcs(templateFuncs)
	page, err := page.Parse(box.String("index.html"))
	if err != nil {
		log.Fatal(err)
	}

	ghClient := githubql.NewClient(oauth2.NewClient(
		context.TODO(),
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")}),
	))

	r := chi.NewRouter()
	r.Get("/index.css", styleHandler(box.Bytes("index.css")))
	r.Get("/github.com/{owner}/{name}", packageHandler(ghClient, page))

	log.Println("starting http server on :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func styleHandler(d []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(d)
	}
}

type (
	dataStats struct {
		Stars        int
		Watchers     int
		Forks        int
		Issues       int
		PullRequests int
	}
	dataTopic struct {
		Name string
		URL  *url.URL
	}
	licenseData struct {
		Name string
		URL  *url.URL
	}
	versionData struct {
		Name       string
		Draft      bool
		Prerelease bool
		Published  time.Time
		URL        *url.URL
	}
	data struct {
		Owner          string
		Name           string
		Description    string
		License        licenseData
		Stats          dataStats
		Topics         []dataTopic
		Versions       []versionData
		CurrentVersion versionData
	}
)

func packageHandler(client *githubql.Client, page *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := chi.URLParam(r, "owner")
		name := chi.URLParam(r, "name")

		var q struct {
			Repository struct {
				Description      githubql.String
				Forks            struct{ TotalCount githubql.Int }
				Stargazers       struct{ TotalCount githubql.Int }
				Watchers         struct{ TotalCount githubql.Int }
				Issues           struct{ TotalCount githubql.Int }
				PullRequests     struct{ TotalCount githubql.Int }
				RepositoryTopics struct {
					Edges []struct {
						Node struct {
							Topic struct {
								Name githubql.String
							}
							URL githubql.URI
						}
					}
				} `graphql:"repositoryTopics(first: 100)"`
				LicenseInfo struct {
					SpdxID githubql.String
					URL    githubql.URI
				}
				Releases struct {
					Edges []struct {
						Node struct {
							IsDraft      githubql.Boolean
							IsPrerelease githubql.Boolean
							PublishedAt  githubql.DateTime
							URL          githubql.URI
							Tag          struct {
								Name githubql.String
							}
						}
					}
				} `graphql:"releases(last:100)"`
				Refs struct {
					Edges []struct {
						Node struct {
							Name githubql.String
						}
					}
				} `graphql:"refs(refPrefix: \"refs/tags/\", first: 100)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		vars := map[string]interface{}{
			"owner": githubql.String(owner),
			"name":  githubql.String(name),
		}

		if err := client.Query(r.Context(), &q, vars); err != nil {
			log.Println(err)
		}

		d := data{
			Owner:       owner,
			Name:        name,
			Description: string(q.Repository.Description),
			License: licenseData{
				Name: string(q.Repository.LicenseInfo.SpdxID),
				URL:  q.Repository.LicenseInfo.URL.URL,
			},
			Stats: dataStats{
				Stars:        int(q.Repository.Stargazers.TotalCount),
				Watchers:     int(q.Repository.Watchers.TotalCount),
				Forks:        int(q.Repository.Forks.TotalCount),
				Issues:       int(q.Repository.Issues.TotalCount),
				PullRequests: int(q.Repository.PullRequests.TotalCount),
			},
		}

		for _, e := range q.Repository.RepositoryTopics.Edges {
			d.Topics = append(d.Topics, dataTopic{
				Name: string(e.Node.Topic.Name),
				URL:  e.Node.URL.URL,
			})
		}

		if len(q.Repository.Releases.Edges) > 0 {
			for _, r := range q.Repository.Releases.Edges {
				d.Versions = append(d.Versions, versionData{
					Name:       string(r.Node.Tag.Name),
					Draft:      bool(r.Node.IsDraft),
					Prerelease: bool(r.Node.IsPrerelease),
					URL:        r.Node.URL.URL,
					Published:  r.Node.PublishedAt.Time,
				})
			}

			sort.Slice(d.Versions, func(i, j int) bool {
				return d.Versions[i].Published.After(d.Versions[j].Published)
			})
		} else {
			for _, r := range q.Repository.Refs.Edges {
				u, _ := url.Parse(fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", owner, name, r.Node.Name))
				d.Versions = append(d.Versions, versionData{
					Name: string(r.Node.Name),
					URL:  u,
				})
			}
		}

		if len(d.Versions) > 0 {
			d.CurrentVersion = d.Versions[0]
		} else {
			//d.CurrentVersion = versionData{
			//	Name:
			//}
		}

		if err := page.Execute(w, d); err != nil {
			panic(err) // TODO
		}
	}
}
