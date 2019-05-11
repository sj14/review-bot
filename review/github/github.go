package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/google/go-github/v25/github"
	"github.com/sj14/review-bot/review"
	"golang.org/x/oauth2"
)

type reminder struct {
	PR      *github.PullRequest
	Missing []string
	// Discussions int
	// Owner github.User // already present in PR
	// Emojis      map[string]int
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

		// fmt.Printf("checking pr: %v\n", pr.GetTitle())
		reviews, resp, err := client.PullRequests.ListReviews(ctx, owner, repo, pr.GetNumber(), &github.ListOptions{1, 10}) // TODO: pagination
		if err != nil {
			log.Fatalf("failed loading reactions: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("failed loading reactions, status code: %v", resp.StatusCode)
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
		missing := review.MissingReviewers(reviewedBy, reviewers)
		// fmt.Printf("%v, %v\n", pr.GetTitle(), missing)
		reminders = append(reminders, reminder{pr, missing})
	}
	// fmt.Printf("reminders: %v\n", reminders)
	return execTemplate(template, repository, reminders)
}

func isRequestedReviewer(reviewers []*github.User, requested *github.User) bool {
	for _, r := range reviewers {
		if r.GetLogin() == requested.GetLogin() {
			return true
		}
	}
	return false
}
