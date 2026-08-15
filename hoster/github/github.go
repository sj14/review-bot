package github

import (
	"github.com/google/go-github/v90/github"
)

type reminder struct {
	PR          *github.PullRequest
	Missing     []string
	Discussions int
	Owner       string
	Emojis      map[string]int
}

// AggregateReminder will generate the reminder message.
func AggregateReminder(token, owner, repo string, reviewers map[string]string) (*github.Repository, []reminder, error) {
	git, err := newClient(token)
	if err != nil {
		return nil, nil, err
	}

	repository, err := git.loadRepository(owner, repo)
	if err != nil {
		return nil, nil, err
	}

	pullRequests, err := git.loadPRs(owner, repo)
	if err != nil {
		return nil, nil, err
	}

	var reminders []reminder

	for _, pr := range pullRequests {
		if pr.GetDraft() {
			continue
		}

		reviews, err := git.loadReviews(owner, repo, pr.GetNumber())
		if err != nil {
			return nil, nil, err
		}

		reviewedBy := getReviewed(pr, reviews)

		missing := missingReviewers(pr.RequestedReviewers, reviewedBy, reviewers)

		// TODO: comments not working
		// fmt.Printf("comments: %v, review comments: %v\n", pr.GetComments(), pr.GetReviewComments())

		owner := responsiblePerson(pr, reviewers)

		// TODO: reactions/emojis
		reminders = append(reminders, reminder{pr, missing, pr.GetComments(), owner, nil})
	}
	return repository, reminders, nil
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
