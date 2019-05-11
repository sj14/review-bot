package gitlab

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/sj14/review-bot/review"
	"github.com/xanzy/go-gitlab"
)

type reminder struct {
	MR          *gitlab.MergeRequest
	Missing     []string
	Discussions int
	Owner       string
	Emojis      map[string]int
}

// AggregateReminder will generate the reminder message.
func AggregateReminder(host, token string, projectID int, reviewers map[string]string, template *template.Template) string {
	// setup gitlab client
	git := newClient(host, token)

	project, reminders := aggregate(git, projectID, reviewers)

	// prevent from sending the header only
	if len(reminders) == 0 {
		return ""
	}

	return execTemplate(template, project, reminders)
}

// helper functions for easier testability (mocked gitlab client)
func aggregate(git client, projectID int, reviewers map[string]string) (gitlab.Project, []reminder) {
	project := git.projectInfo(projectID)

	// get open merge requests
	mergeRequests := git.openMergeRequests(projectID)

	// TODO: add option
	// only return merge requests which have no open discussions
	// mergeRequests = filterOpenDiscussions(git, mergeRequests)

	// will contain the reminders of all merge requests
	var reminders []reminder

	for _, mr := range mergeRequests {
		// don't check WIP MRs
		if mr.WorkInProgress {
			continue
		}

		// load all emojis awarded to the mr
		emojis := git.loadEmojis(projectID, mr)

		// check who gave thumbs up/down (or "sleeping")
		reviewedBy := getReviewed(mr, emojis)

		// who is missing thumbs up/down
		missing := review.MissingReviewers(reviewedBy, reviewers)

		// load all discussions of the mr
		discussions := git.loadDiscussions(projectID, mr)

		// get the number of open discussions
		discussionsCount := openDiscussionsCount(discussions)

		// get the responsible person of the mr
		owner := responsiblePerson(mr, reviewers)

		// list each emoji with the usage count
		emojisAggr := aggregateEmojis(emojis)

		reminders = append(reminders, reminder{mr, missing, discussionsCount, owner, emojisAggr})
	}

	return project, reminders
}

//go:generate moq -out client_moq_test.go . client
type client interface {
	projectInfo(id int) gitlab.Project
	openMergeRequests(projectID int) []*gitlab.MergeRequest
	loadEmojis(projectID int, mr *gitlab.MergeRequest) []*gitlab.AwardEmoji
	loadDiscussions(projectID int, mr *gitlab.MergeRequest) []*gitlab.Discussion
}

type clientWrapper struct {
	original *gitlab.Client
}

// newClient returns a new gitlab client.
func newClient(host, token string) *clientWrapper {
	client := gitlab.NewClient(nil, token)
	if err := client.SetBaseURL(fmt.Sprintf("https://%s/api/v4", host)); err != nil {
		log.Fatalf("failed to set gitlab host: %v", err)
	}
	return &clientWrapper{original: client}
}

func (cw *clientWrapper) projectInfo(id int) gitlab.Project {
	p, resp, err := cw.original.Projects.GetProject(id, nil)
	if err != nil {
		log.Fatalf("failed to get project: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("failed to get project, status code: %v", resp.StatusCode)
	}

	return *p
}

// responsiblePerson returns the mattermost name of the assignee or author of the MR
// (fallback: gitlab author name)
func responsiblePerson(mr *gitlab.MergeRequest, reviewers map[string]string) string {
	if mr.Assignee.ID != 0 {
		if assignee, ok := reviewers[strconv.Itoa(mr.Assignee.ID)]; ok {
			return assignee
		}
	}

	if author, ok := reviewers[strconv.Itoa(mr.Author.ID)]; ok {
		return author
	}

	return mr.Author.Name
}

// openMergeRequests returns all open merge requests of the given project.
func (cw *clientWrapper) openMergeRequests(projectID int) []*gitlab.MergeRequest {
	// options
	var (
		mergeRequests []*gitlab.MergeRequest
		state         = "opened"
		opts          = &gitlab.ListProjectMergeRequestsOptions{
			State:       &state,
			ListOptions: gitlab.ListOptions{PerPage: 100},
		}
	)

	for {
		pageMRs, resp, err := cw.original.MergeRequests.ListProjectMergeRequests(projectID, opts)
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
func (cw *clientWrapper) loadDiscussions(projectID int, mr *gitlab.MergeRequest) []*gitlab.Discussion {

	var (
		discussions []*gitlab.Discussion
		opts        = &gitlab.ListMergeRequestDiscussionsOptions{PerPage: 100}
	)

	for {
		pageDiscussions, resp, err := cw.original.Discussions.ListMergeRequestDiscussions(projectID, mr.IID, opts)
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

// openDiscussionsCount returns the number of open discussions.
func openDiscussionsCount(discussions []*gitlab.Discussion) int {
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

// filterOpenDiscussions returns only merge requests which have no open discussions.
func filterOpenDiscussions(mergeRequests []*gitlab.MergeRequest, discussions []*gitlab.Discussion) []*gitlab.MergeRequest {
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

// loadEmojis returns all emoji reactions of the particular MR.
func (cw *clientWrapper) loadEmojis(projectID int, mr *gitlab.MergeRequest) []*gitlab.AwardEmoji {

	var (
		emojis []*gitlab.AwardEmoji
		opts   = &gitlab.ListAwardEmojiOptions{PerPage: 100}
	)

	for {
		pageEmojis, resp, err := cw.original.AwardEmoji.ListMergeRequestAwardEmoji(projectID, mr.IID, opts)
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

// getReviewed returns the gitlab user id of the people who have already reviewed the MR.
// The emojis "thumbsup" 👍 and "thumbsdown" 👎 signal the user reviewed the merge request and won't receive a reminder.
// The emoji "sleeping" 😴 means the user won't review the code and/or doesn't want to be reminded.
func getReviewed(mr *gitlab.MergeRequest, emojis []*gitlab.AwardEmoji) []string {
	var reviewedBy = []string{strconv.Itoa(mr.Author.ID)}

	for _, emoji := range emojis {
		if emoji.Name == "thumbsup" ||
			emoji.Name == "thumbsdown" ||
			emoji.Name == "sleeping" {
			reviewedBy = append(reviewedBy, strconv.Itoa(emoji.User.ID))
		}
	}

	return reviewedBy
}

// aggregateEmojis lists all emojis with their usage count.
func aggregateEmojis(emojis []*gitlab.AwardEmoji) map[string]int {
	var aggregate = make(map[string]int)

	for _, emoji := range emojis {
		count := aggregate[emoji.Name]
		count++
		aggregate[emoji.Name] = count
	}

	return aggregate
}
