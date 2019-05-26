package github

import (
	"text/template"

	"github.com/google/go-github/v25/github"
)

type reminder struct {
	PR          *github.PullRequest
	Missing     []string
	Discussions int
	Owner       string
	Emojis      map[string]int
}

// AggregateReminder will generate the reminder message.
func AggregateReminder(token, owner, repo string, reviewers map[string]string, template *template.Template) string {
	var (
		reminders    []reminder
		git          = newClient(token)
		repository   = git.loadRepository(owner, repo)
		pullRequests = git.loadPRs(owner, repo)
	)

	for _, pr := range pullRequests {
		if pr.GetDraft() {
			continue
		}

		reviews := git.loadReviews(owner, repo, pr.GetNumber())

		reviewedBy := getReviewed(pr, reviews)

		missing := missingReviewers(pr.RequestedReviewers, reviewedBy, reviewers)

		// TODO: comments not working
		// fmt.Printf("comments: %v, review comments: %v\n", pr.GetComments(), pr.GetReviewComments())

		owner := responsiblePerson(pr, reviewers)

		// TODO: reactions/emojis
		reminders = append(reminders, reminder{pr, missing, pr.GetComments(), owner, nil})
	}
	return execTemplate(template, repository, reminders)
}

const (
	approved  = "APPROVED"
	dismissed = "DISMISSED"
)

func getReviewed(pr *github.PullRequest, reviews []*github.PullRequestReview) []string {
	var reviewedBy []string
	for _, rev := range reviews {
		if rev.GetState() == approved || rev.GetState() == dismissed {
			reviewedBy = append(reviewedBy, rev.GetUser().GetLogin())
		}
	}
	return reviewedBy
}

func prepareReactions(reactions *github.Reactions) map[string]int {
	result := make(map[string]int)

	if i := reactions.GetConfused(); i > 0 {
		result[":confused:"] = i
	}
	if i := reactions.GetHeart(); i > 0 {
		result[":heart:"] = i
	}
	if i := reactions.GetHooray(); i > 0 {
		result[":hooray:"] = i
	}
	if i := reactions.GetLaugh(); i > 0 {
		result[":laugh:"] = i
	}
	if i := reactions.GetMinusOne(); i > 0 {
		result[":-1:"] = i
	}
	if i := reactions.GetPlusOne(); i > 0 {
		result[":+1:"] = i
	}
	return result
}

func missingReviewers(requested []*github.User, reviewedBy []string, mapping map[string]string) []string {
	var missing []string

	for _, requested := range requested {
		approved := false
		added := false

		for userID, userName := range mapping {
			if requested.GetLogin() != userID {
				continue
			}
			for _, approverID := range reviewedBy {
				if userID == approverID {
					approved = true
					break
				}
			}
			if !approved {
				missing = append(missing, userName)
				added = true
			}
			// we found the requested user/mapping,
			// don't check further mappings for this user
			break
		}
		// missing chat name mapping, use github login as fallback
		if !approved && !added {
			missing = append(missing, requested.GetLogin())
		}
	}
	return missing
}

func isRequestedReviewer(reviewers []*github.User, requested *github.User) bool {
	for _, r := range reviewers {
		if r.GetLogin() == requested.GetLogin() {
			return true
		}
	}
	return false
}

func responsiblePerson(pr *github.PullRequest, reviewers map[string]string) string {
	// corresponding mattermost name
	if author, ok := reviewers[pr.GetUser().GetLogin()]; ok {
		return author
	}

	// fallback
	return pr.GetUser().GetLogin()
}
