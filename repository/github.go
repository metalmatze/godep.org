package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shurcooL/githubql"
	"golang.org/x/oauth2"
)

type (
	GitHub struct {
		client *githubql.Client
	}
)

func NewGitHubClient(token string) (*GitHub, error) {
	return &GitHub{
		client: githubql.NewClient(oauth2.NewClient(
			context.TODO(),
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
		)),
	}, nil
}

func (gh *GitHub) Get(ctx context.Context, url string) (Repository, error) {
	urlParts := strings.Split(url, "/")
	owner, name := urlParts[1], urlParts[2]

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

	if err := gh.client.Query(ctx, &q, vars); err != nil {
		return Repository{}, err
	}

	repo := Repository{
		URL:         url,
		Description: string(q.Repository.Description),
		Updated:     time.Now(),
		Stats: []Stat{{
			Name:  "Forks",
			Value: int(q.Repository.Forks.TotalCount),
			URL:   fmt.Sprintf("https://github.com/%s/%s/network", owner, name),
		}, {
			Name:  "Issues",
			Value: int(q.Repository.Issues.TotalCount),
			URL:   fmt.Sprintf("https://github.com/%s/%s/issues", owner, name),
		}, {
			Name:  "PullRequests",
			Value: int(q.Repository.PullRequests.TotalCount),
			URL:   fmt.Sprintf("https://github.com/%s/%s/pulls", owner, name),
		}, {
			Name:  "Stars",
			Value: int(q.Repository.Stargazers.TotalCount),
			URL:   fmt.Sprintf("https://github.com/%s/%s/stargazers", owner, name),
		}, {
			Name:  "Watchers",
			Value: int(q.Repository.Watchers.TotalCount),
			URL:   fmt.Sprintf("https://github.com/%s/%s/watchers", owner, name),
		}},
	}

	return repo, nil

	//d := data{
	//	Owner:       owner,
	//	Name:        name,
	//	Description: string(q.Repository.Description),
	//	License: licenseData{
	//		Name: string(q.Repository.LicenseInfo.SpdxID),
	//		URL:  q.Repository.LicenseInfo.URL.URL,
	//	},
	//	Stats: dataStats{
	//		Stars:        int(q.Repository.Stargazers.TotalCount),
	//		Watchers:     int(q.Repository.Watchers.TotalCount),
	//		Forks:        int(q.Repository.Forks.TotalCount),
	//		Issues:       int(q.Repository.Issues.TotalCount),
	//		PullRequests: int(q.Repository.PullRequests.TotalCount),
	//	},
	//}
	//
	//for _, e := range q.Repository.RepositoryTopics.Edges {
	//	d.Topics = append(d.Topics, dataTopic{
	//		Name: string(e.Node.Topic.Name),
	//		URL:  e.Node.URL.URL,
	//	})
	//}
	//
	//if len(q.Repository.Releases.Edges) > 0 {
	//	for _, r := range q.Repository.Releases.Edges {
	//		d.Versions = append(d.Versions, versionData{
	//			Name:       string(r.Node.Tag.Name),
	//			Draft:      bool(r.Node.IsDraft),
	//			Prerelease: bool(r.Node.IsPrerelease),
	//			URL:        r.Node.URL.URL,
	//			Published:  r.Node.PublishedAt.Time,
	//		})
	//	}
	//
	//	sort.Slice(d.Versions, func(i, j int) bool {
	//		return d.Versions[i].Published.After(d.Versions[j].Published)
	//	})
	//} else {
	//	for _, r := range q.Repository.Refs.Edges {
	//		u, _ := url.Parse(fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", owner, name, r.Node.Name))
	//		d.Versions = append(d.Versions, versionData{
	//			Name: string(r.Node.Name),
	//			URL:  u,
	//		})
	//	}
	//}
	//
	//if len(d.Versions) > 0 {
	//	d.CurrentVersion = d.Versions[0]
	//} else {
	//	//d.CurrentVersion = versionData{
	//	//	Name:
	//	//}
	//}
}
