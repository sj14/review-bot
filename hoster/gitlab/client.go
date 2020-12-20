package gitlab

import (
	"fmt"
	"log"
	"net/http"

	"github.com/xanzy/go-gitlab"
)

//go:generate moq -out client_moq_test.go . clientWrapper
type clientWrapper interface {
	loadProject(repo interface{}) gitlab.Project
	loadMRs(repo interface{}) []*gitlab.MergeRequest
	loadEmojis(repo interface{}, mr *gitlab.MergeRequest) []*gitlab.AwardEmoji
	loadDiscussions(repo interface{}, mr *gitlab.MergeRequest) []*gitlab.Discussion
}

type client struct {
	original *gitlab.Client
}

// newClient returns a new gitlab client.
func newClient(host, token string) *client {
	c, err := gitlab.NewClient(token, gitlab.WithBaseURL(fmt.Sprintf("https://%s/api/v4", host)))
	if err != nil {
		log.Fatalf("failed creating new giltab client: %v\n", err)
	}

	return &client{original: c}
}

func (c *client) loadProject(repo interface{}) gitlab.Project {
	p, resp, err := c.original.Projects.GetProject(repo, nil)
	if err != nil {
		log.Fatalf("failed to get project: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("failed to get project, status code: %v", resp.StatusCode)
	}

	return *p
}

// openMergeRequests returns all open merge requests of the given project.
func (c *client) loadMRs(repo interface{}) []*gitlab.MergeRequest {
	var (
		mergeRequests []*gitlab.MergeRequest
		state         = "opened"
		opts          = &gitlab.ListProjectMergeRequestsOptions{
			State:       &state,
			ListOptions: gitlab.ListOptions{PerPage: 25},
		}
	)

	for {
		pageMRs, resp, err := c.original.MergeRequests.ListProjectMergeRequests(repo, opts)
		if err != nil {
			log.Fatalf("failed to list project merge requests: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("failed to list project merge requests, status code: %v", resp.StatusCode)
		}
		mergeRequests = append(mergeRequests, pageMRs...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return mergeRequests
}

// loadDiscussions of the given MR.
func (c *client) loadDiscussions(repo interface{}, mr *gitlab.MergeRequest) []*gitlab.Discussion {
	var (
		discussions []*gitlab.Discussion
		opts        = &gitlab.ListMergeRequestDiscussionsOptions{PerPage: 25}
	)

	for {
		pageDiscussions, resp, err := c.original.Discussions.ListMergeRequestDiscussions(repo, mr.IID, opts)
		if err != nil {
			log.Fatalf("failed to list emojis for MR %v: %v", mr.IID, err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("failed to list emojis, status code: %v", resp.StatusCode)
		}
		discussions = append(discussions, pageDiscussions...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return discussions
}

// loadEmojis returns all emoji reactions of the particular MR.
func (c *client) loadEmojis(repo interface{}, mr *gitlab.MergeRequest) []*gitlab.AwardEmoji {
	var (
		emojis []*gitlab.AwardEmoji
		opts   = &gitlab.ListAwardEmojiOptions{PerPage: 25}
	)

	for {
		pageEmojis, resp, err := c.original.AwardEmoji.ListMergeRequestAwardEmoji(repo, mr.IID, opts)
		if err != nil {
			log.Fatalf("failed to list emojis for MR %v: %v", mr.IID, err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("failed to list emojis, status code: %v", resp.StatusCode)
		}
		emojis = append(emojis, pageEmojis...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return emojis
}
