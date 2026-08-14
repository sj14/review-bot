package gitlab

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.com/gitlab-org/api/client-go/v2"
)

const httpTimeout = 30 * time.Second

//go:generate moq -out client_moq_test.go . clientWrapper
type clientWrapper interface {
	loadProject(repo interface{}) (gitlab.Project, error)
	loadMRs(repo interface{}) ([]*gitlab.BasicMergeRequest, error)
	loadEmojis(repo interface{}, mr *gitlab.BasicMergeRequest) ([]*gitlab.AwardEmoji, error)
	loadDiscussions(repo interface{}, mr *gitlab.BasicMergeRequest) ([]*gitlab.Discussion, error)
}

type client struct {
	original *gitlab.Client
}

// newClient returns a new gitlab client.
func newClient(host, token string) (*client, error) {
	c, err := gitlab.NewClient(
		token,
		gitlab.WithBaseURL(fmt.Sprintf("https://%s/api/v4", host)),
		gitlab.WithHTTPClient(&http.Client{Timeout: httpTimeout}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating new gitlab client: %w", err)
	}

	return &client{original: c}, nil
}

func (c *client) loadProject(repo interface{}) (gitlab.Project, error) {
	p, resp, err := c.original.Projects.GetProject(repo, nil)
	if err != nil {
		return gitlab.Project{}, fmt.Errorf("failed to get project: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return gitlab.Project{}, fmt.Errorf("failed to get project, status code: %v", resp.StatusCode)
	}

	return *p, nil
}

// openMergeRequests returns all open merge requests of the given project.
func (c *client) loadMRs(repo interface{}) ([]*gitlab.BasicMergeRequest, error) {
	var (
		mergeRequests []*gitlab.BasicMergeRequest
		state         = "opened"
		opts          = &gitlab.ListProjectMergeRequestsOptions{
			State:       &state,
			ListOptions: gitlab.ListOptions{PerPage: 25},
		}
	)

	for {
		pageMRs, resp, err := c.original.MergeRequests.ListProjectMergeRequests(repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list project merge requests: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to list project merge requests, status code: %v", resp.StatusCode)
		}
		mergeRequests = append(mergeRequests, pageMRs...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return mergeRequests, nil
}

// loadDiscussions of the given MR.
func (c *client) loadDiscussions(repo interface{}, mr *gitlab.BasicMergeRequest) ([]*gitlab.Discussion, error) {
	var (
		discussions []*gitlab.Discussion
		opts        = &gitlab.ListMergeRequestDiscussionsOptions{ListOptions: gitlab.ListOptions{PerPage: 25}}
	)

	for {
		pageDiscussions, resp, err := c.original.Discussions.ListMergeRequestDiscussions(repo, mr.IID, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list discussions for MR %v: %w", mr.IID, err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to list discussions, status code: %v", resp.StatusCode)
		}
		discussions = append(discussions, pageDiscussions...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return discussions, nil
}

// loadEmojis returns all emoji reactions of the particular MR.
func (c *client) loadEmojis(repo interface{}, mr *gitlab.BasicMergeRequest) ([]*gitlab.AwardEmoji, error) {
	var (
		emojis []*gitlab.AwardEmoji
		opts   = &gitlab.ListAwardEmojiOptions{ListOptions: gitlab.ListOptions{PerPage: 25}}
	)

	for {
		pageEmojis, resp, err := c.original.AwardEmoji.ListMergeRequestAwardEmoji(repo, mr.IID, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list emojis for MR %v: %w", mr.IID, err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to list emojis, status code: %v", resp.StatusCode)
		}
		emojis = append(emojis, pageEmojis...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return emojis, nil
}
