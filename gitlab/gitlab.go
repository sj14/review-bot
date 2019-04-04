package gitlab

import (
	"fmt"
	"log"
	"net/http"

	"github.com/xanzy/go-gitlab"
)

// NewClient returns a new gitlab client.
func NewClient(host, token string) *gitlab.Client {
	client := gitlab.NewClient(nil, token)
	if err := client.SetBaseURL(fmt.Sprintf("https://%s/api/v4", host)); err != nil {
		log.Fatalf("failed to set gitlab host: %v", err)
	}
	return client
}

// ResponsiblePerson returns the mattermost name of the assignee or author of the MR
// (fallback: gitlab author name)
func ResponsiblePerson(mr *gitlab.MergeRequest, reviewers map[int]string) string {
	if mr.Assignee.ID != 0 {
		assignee, ok := reviewers[mr.Assignee.ID]
		if ok {
			return assignee
		}
	}

	author, ok := reviewers[mr.Author.ID]
	if ok {
		return author
	}

	return mr.Author.Name
}

// OpenMergeRequests returns all open merge requests of the given project.
func OpenMergeRequests(git *gitlab.Client, projectID int) []*gitlab.MergeRequest {
	// options
	state := "opened"
	opts := &gitlab.ListProjectMergeRequestsOptions{State: &state, ListOptions: gitlab.ListOptions{PerPage: 100}}

	// first page
	mergeRequests, resp, err := git.MergeRequests.ListProjectMergeRequests(projectID, opts)
	if err != nil {
		log.Fatalf("failed to list project merge requests: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("failed to list project merge requests, status code: %v", resp.StatusCode)
	}

	// following pages
	for page := 2; page <= resp.TotalPages; page++ {
		opts.Page = page

		pageMRs, resp, err := git.MergeRequests.ListProjectMergeRequests(projectID, opts)
		if err != nil {
			log.Fatalf("failed to list project merge requests: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("failed to list project merge requests, status code: %v", resp.StatusCode)
		}
		mergeRequests = append(mergeRequests, pageMRs...)
	}

	return mergeRequests
}

// LoadDiscussions of the given MR.
func LoadDiscussions(git *gitlab.Client, projectID int, mr *gitlab.MergeRequest) []*gitlab.Discussion {
	opts := &gitlab.ListMergeRequestDiscussionsOptions{PerPage: 100}

	// first page
	discussions, resp, err := git.Discussions.ListMergeRequestDiscussions(projectID, mr.IID, opts)
	if err != nil {
		log.Fatalf("failed to get discussions for mr %v: %v", mr.IID, err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("failed to list emojis, status code: %v", resp.StatusCode)
	}

	// following pages
	for page := 2; page <= resp.TotalPages; page++ {
		opts.Page = page

		pageDiscussions, resp, err := git.Discussions.ListMergeRequestDiscussions(projectID, mr.IID, opts)
		if err != nil {
			log.Fatalf("failed to list emojis for MR %v: %v", mr.IID, err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("failed to list emojis, status code: %v", resp.StatusCode)
		}
		discussions = append(discussions, pageDiscussions...)
	}

	return discussions
}

// OpenDiscussionsCount returns the number of open discussions.
func OpenDiscussionsCount(discussions []*gitlab.Discussion) int {
	// check if any of the discussions are unresolved
	count := 0
	for _, d := range discussions {
		for _, n := range d.Notes {
			if !n.Resolved && n.Resolvable {
				count++
			}
		}
	}
	return count
}

// FilterOpenDiscussions returns only merge requests which have no open discussions.
func FilterOpenDiscussions(mergeRequests []*gitlab.MergeRequest, discussions []*gitlab.Discussion) []*gitlab.MergeRequest {
	result := []*gitlab.MergeRequest{}

	for _, mr := range mergeRequests {
		// check if any of the discussions are unresolved
		anyUnresolved := false
	LoopDiscussions:
		for _, d := range discussions {
			for _, n := range d.Notes {
				if !n.Resolved && n.Resolvable {
					anyUnresolved = true
					break LoopDiscussions
				}
			}
		}

		// don't add merge request with unresolved discussion
		if !anyUnresolved {
			result = append(result, mr)
		}
	}
	return result
}

// LoadEmojis returns all emoji reactions of the particular MR.
func LoadEmojis(git *gitlab.Client, projectID int, mr *gitlab.MergeRequest) []*gitlab.AwardEmoji {
	opts := &gitlab.ListAwardEmojiOptions{PerPage: 100}

	// first page
	emojis, resp, err := git.AwardEmoji.ListMergeRequestAwardEmoji(projectID, mr.IID, opts)
	if err != nil {
		log.Fatalf("failed to list emojis for MR %v: %v", mr.IID, err)
	}

	// following pages
	for page := 2; page <= resp.TotalPages; page++ {
		opts.Page = page

		pageEmojis, resp, err := git.AwardEmoji.ListMergeRequestAwardEmoji(projectID, mr.IID, opts)
		if err != nil {
			log.Fatalf("failed to list emojis for MR %v: %v", mr.IID, err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("failed to list emojis, status code: %v", resp.StatusCode)
		}
		emojis = append(emojis, pageEmojis...)
	}

	return emojis
}

// GetReviewed returns the gitlab user id of the people who have already reviewed the MR.
// The emojis "thumbsup" 👍 and "thumbsdown" 👎 signal the user reviewed the merge request and won't receive a reminder.
// The emoji "sleeping" 😴 means the user won't review the code and/or doesn't want to be reminded.
func GetReviewed(mr *gitlab.MergeRequest, emojis []*gitlab.AwardEmoji) []int {
	var reviewedBy []int
	reviewedBy = append(reviewedBy, mr.Author.ID)
	for _, emoji := range emojis {
		if emoji.Name == "thumbsup" || emoji.Name == "thumbsdown" || emoji.Name == "sleeping" {
			reviewedBy = append(reviewedBy, emoji.User.ID)
		}
	}

	return reviewedBy
}

// AggregateEmojis lists all emojis with their usage count.
func AggregateEmojis(emojis []*gitlab.AwardEmoji) map[string]int {
	var aggregate map[string]int
	aggregate = make(map[string]int)

	for _, emoji := range emojis {
		count := aggregate[emoji.Name]
		count++
		aggregate[emoji.Name] = count
	}

	return aggregate
}

// MissingReviewers returns all reviewers which have not reacted with 👍, 👎 or 😴.
func MissingReviewers(reviewedBy []int, approvers map[int]string) []string {
	var result []string
	for userID, userName := range approvers {
		approved := false
		for _, approverID := range reviewedBy {
			if userID == approverID {
				approved = true
				break
			}
		}
		if !approved {
			result = append(result, userName)
		}
	}

	return result
}
