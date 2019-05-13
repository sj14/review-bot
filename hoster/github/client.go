package github

import (
	"context"
	"log"
	"net/http"

	"github.com/google/go-github/v25/github"
	"golang.org/x/oauth2"
)

//go:generate moq -out client_moq_test.go . clientWrapper
type clientWrapper interface {
	loadRepository(owner, repo string) *github.Repository
	loadPRs(owner, repo string) []*github.PullRequest
	loadReviews(owner, repo string, number int) []*github.PullRequestReview
}

type client struct {
	original *github.Client
	ctx      context.Context
}

// newClient returns a new github client.
func newClient(token string) *client {
	ctx := context.Background()

	c := github.NewClient(nil)
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		c = github.NewClient(tc)
	}

	return &client{original: c, ctx: ctx}
}

func (c *client) loadRepository(owner, repo string) *github.Repository {
	repository, resp, err := c.original.Repositories.Get(c.ctx, owner, repo)
	if err != nil {
		log.Fatalf("failed loading repo: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("failed loading repo, status code: %v", resp.StatusCode)
	}
	return repository
}

func (c *client) loadPRs(owner, repo string) []*github.PullRequest {
	var (
		pullRequests []*github.PullRequest
		opts         = &github.PullRequestListOptions{ListOptions: github.ListOptions{PerPage: 25}}
	)

	for {
		pagePRs, resp, err := c.original.PullRequests.List(c.ctx, owner, repo, opts)
		if err != nil {
			log.Fatalf("failed loading pull requests: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("failed loading pull requests, status code: %v", resp.StatusCode)
		}
		pullRequests = append(pullRequests, pagePRs...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return pullRequests
}

func (c *client) loadReviews(owner, repo string, number int) []*github.PullRequestReview {
	var (
		reviews []*github.PullRequestReview
		opts    = &github.ListOptions{PerPage: 25}
	)

	for {
		pageReviews, resp, err := c.original.PullRequests.ListReviews(c.ctx, owner, repo, number, opts)
		if err != nil {
			log.Fatalf("failed loading reviews: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("failed loading reviews, status code: %v", resp.StatusCode)
		}
		reviews = append(reviews, pageReviews...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return reviews
}
