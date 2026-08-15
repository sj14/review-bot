package github

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v90/github"
)

const httpTimeout = 30 * time.Second

//go:generate go tool moq -out client_moq_test.go . clientWrapper
type clientWrapper interface {
	loadRepository(owner, repo string) (*github.Repository, error)
	loadPRs(owner, repo string) ([]*github.PullRequest, error)
	loadReviews(owner, repo string, number int) ([]*github.PullRequestReview, error)
}

type client struct {
	original *github.Client
	ctx      context.Context
}

// newClient returns a new github client.
func newClient(token string) (*client, error) {
	ctx := context.Background()

	opts := []github.ClientOptionsFunc{github.WithTimeout(httpTimeout)}
	if token != "" {
		opts = append(opts, github.WithAuthToken(token))
	}

	c, err := github.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed creating new github client: %w", err)
	}

	return &client{original: c, ctx: ctx}, nil
}

func (c *client) loadRepository(owner, repo string) (*github.Repository, error) {
	repository, resp, err := c.original.Repositories.Get(c.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed loading repo: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed loading repo, status code: %v", resp.StatusCode)
	}
	return repository, nil
}

func (c *client) loadPRs(owner, repo string) ([]*github.PullRequest, error) {
	var (
		pullRequests []*github.PullRequest
		opts         = &github.PullRequestListOptions{ListOptions: github.ListOptions{PerPage: 25}}
	)

	for {
		pagePRs, resp, err := c.original.PullRequests.List(c.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed loading pull requests: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed loading pull requests, status code: %v", resp.StatusCode)
		}
		pullRequests = append(pullRequests, pagePRs...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return pullRequests, nil
}

func (c *client) loadReviews(owner, repo string, number int) ([]*github.PullRequestReview, error) {
	var (
		reviews []*github.PullRequestReview
		opts    = &github.ListOptions{PerPage: 25}
	)

	for {
		pageReviews, resp, err := c.original.PullRequests.ListReviews(c.ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("failed loading reviews: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed loading reviews, status code: %v", resp.StatusCode)
		}
		reviews = append(reviews, pageReviews...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return reviews, nil
}
