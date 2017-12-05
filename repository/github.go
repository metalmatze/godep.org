package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/shurcooL/githubql"
	"golang.org/x/oauth2"
)

type (
	// GitHub makes API calls to GitHub
	GitHub struct {
		client   *githubql.Client
		apiCalls metrics.Histogram
	}
)

// NewGitHubClient initializes a new GitHub client from a token
func NewGitHubClient(token string, apiCalls metrics.Histogram) (*GitHub, error) {
	gh := &GitHub{
		client: githubql.NewClient(oauth2.NewClient(
			context.TODO(),
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
		)),
		apiCalls: apiCalls.With("service", "github"),
	}

	// Initialize metric with a zero value
	gh.apiCalls.Observe(0)

	return gh, nil
}

// Get a repository's data from its urlPath
func (gh *GitHub) Get(ctx context.Context, urlPath string) (Repository, error) {
	defer func(start time.Time) {
		gh.apiCalls.Observe(time.Since(start).Seconds())
	}(time.Now())

	urlParts := strings.Split(urlPath, "/")
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
		URL:         urlPath,
		Description: string(q.Repository.Description),
		Updated:     time.Now(),
		Statistics: []Statistic{{
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

	if len(q.Repository.Releases.Edges) > 0 {
		for _, r := range q.Repository.Releases.Edges {
			repo.Versions = append(repo.Versions, Version{
				Name:      string(r.Node.Tag.Name),
				Published: r.Node.PublishedAt.Time,
			})
		}
	} else {
		for _, r := range q.Repository.Refs.Edges {
			repo.Versions = append(repo.Versions, Version{
				Name: string(r.Node.Name),
			})
		}
	}

	return repo, nil
}
