package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/google/go-github/v25/github"
	"golang.org/x/oauth2"
)

type reminder struct {
	PR          *github.PullRequest
	Missing     []string
	Discussions int
	// Owner github.User // already present in PR
	Emojis map[string]int
}

func AggregateReminder(token, owner, repo string, reviewers map[string]string, template *template.Template) string {
	ctx := context.Background()

	client := github.NewClient(nil)
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}

	repository, resp, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		log.Fatalf("failed loading repo: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("failed loading repo, status code: %v", resp.StatusCode)
	}

	fmt.Println(repository.GetName())
	pullRequests, resp, err := client.PullRequests.List(ctx, owner, repo, nil)
	if err != nil {
		log.Fatalf("failed loading pull requests: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("failed loading pull requests, status code: %v", resp.StatusCode)
	}

	var reminders []reminder

	for _, pr := range pullRequests {
		if pr.GetDraft() {
			continue
		}

		reviews, resp, err := client.PullRequests.ListReviews(ctx, owner, repo, pr.GetNumber(), &github.ListOptions{1, 10}) // TODO: pagination
		if err != nil {
			log.Fatalf("failed loading reviews: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("failed loading reviews, status code: %v", resp.StatusCode)
		}

		var reviewedBy []string
		for _, rev := range reviews {
			if ok := isRequestedReviewer(pr.RequestedReviewers, rev.GetUser()); !ok {
				continue
			}
			fmt.Printf("check review from %v\n", rev.GetUser().GetLogin())

			if rev.GetState() == "APPROVED" || rev.GetState() == "DISMISSED" {
				reviewedBy = append(reviewedBy, rev.GetUser().GetLogin()) // TODO: fix casting to int
				fmt.Printf("Added as reviewed \n")
			}
		}
		missing := missingReviewers(pr.RequestedReviewers, reviewedBy, reviewers)

		// TODO: comments not working
		// fmt.Printf("comments: %v, review comments: %v\n", pr.GetComments(), pr.GetReviewComments())

		// TODO: reactions/emojis
		reminders = append(reminders, reminder{pr, missing, pr.GetComments(), nil})
	}
	return execTemplate(template, repository, reminders)
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
		}
		// missing mapping
		if !added {
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
