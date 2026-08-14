package gitlab

import (
	"gitlab.com/gitlab-org/api/client-go/v2"
)

type reminder struct {
	MR          *gitlab.BasicMergeRequest
	Missing     []string
	Discussions int
	Owner       string
	Emojis      map[string]int
}

// AggregateReminder will generate the reminder message.
func AggregateReminder(host, token string, repo interface{}, reviewers map[string]string) (gitlab.Project, []reminder) {
	// setup gitlab client
	git := newClient(host, token)

	project, reminders := aggregate(git, repo, reviewers)

	// prevent from sending the header only
	// if len(reminders) == 0 {
	// 	return ""
	// }

	return project, reminders
}

// func AggregateUsersReminder(host, token string, repo interface{}, reviewers map[string]string, template *template.Template) string {
// 	// setup gitlab client
// 	git := newClient(host, token)

// 	project, reminders := aggregate(git, repo, reviewers)

// 	// prevent from sending the header only
// 	if len(reminders) == 0 {
// 		return ""
// 	}

// 	return ExecTemplate(template, project, reminders)
// }

// helper functions for easier testability (mocked gitlab client)
func aggregate(git clientWrapper, repo interface{}, reviewers map[string]string) (gitlab.Project, []reminder) {
	project := git.loadProject(repo)

	// get open merge requests
	mergeRequests := git.loadMRs(repo)

	// TODO: add option
	// only return merge requests which have no open discussions
	// mergeRequests = filterOpenDiscussions(git, mergeRequests)

	// will contain the reminders of all merge requests
	var reminders []reminder

	for _, mr := range mergeRequests {
		// don't check WIP MRs
		if mr.Draft {
			continue
		}

		// load all emojis awarded to the mr
		emojis := git.loadEmojis(repo, mr)

		// check who gave thumbs up/down (or "sleeping")
		reviewedBy := getReviewed(mr, emojis)

		// who is missing thumbs up/down
		missing := missingReviewers(reviewedBy, reviewers)

		// load all discussions of the mr
		discussions := git.loadDiscussions(repo, mr)

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

// responsiblePerson returns the mattermost name of the assignee or author of the MR
// (fallback: gitlab author name)
func responsiblePerson(mr *gitlab.BasicMergeRequest, reviewers map[string]string) string {
	if mr == nil {
		return ""
	}

	if mr.Assignee != nil && mr.Assignee.Username != "" {
		if assignee, ok := reviewers[mr.Assignee.Username]; ok {
			return assignee
		}
	}

	if mr.Author == nil {
		return ""
	}

	if author, ok := reviewers[mr.Author.Username]; ok {
		return author
	}

	return mr.Author.Name
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
func filterOpenDiscussions(mergeRequests []*gitlab.BasicMergeRequest, discussions []*gitlab.Discussion) []*gitlab.BasicMergeRequest {
	result := []*gitlab.BasicMergeRequest{}

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

const (
	thumbsup   = "thumbsup"
	thumbsdown = "thumbsdown"
	sleeping   = "sleeping"
)

// getReviewed returns the gitlab user id of the people who have already reviewed the MR.
// The emojis "thumbsup" 👍 and "thumbsdown" 👎 signal the user reviewed the merge request and won't receive a reminder.
// The emoji "sleeping" 😴 means the user won't review the code and/or doesn't want to be reminded.
func getReviewed(mr *gitlab.BasicMergeRequest, emojis []*gitlab.AwardEmoji) []string {
	var reviewedBy []string

	if mr != nil && mr.Author != nil {
		reviewedBy = append(reviewedBy, mr.Author.Username)
	}

	for _, emoji := range emojis {
		if emoji.Name == thumbsup ||
			emoji.Name == thumbsdown ||
			emoji.Name == sleeping {
			reviewedBy = append(reviewedBy, emoji.User.Username)
		}
	}

	return reviewedBy
}

func missingReviewers(reviewedBy []string, approvers map[string]string) []string {
	var missing []string
	for userID, userName := range approvers {
		approved := false
		for _, approverID := range reviewedBy {
			if userID == approverID {
				approved = true
				break
			}
		}
		if !approved {
			missing = append(missing, userName)
		}
	}

	return missing
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
