package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/sj14/review-bot/gitlab"
	"github.com/sj14/review-bot/mattermost"
)

func main() {
	var (
		host          = flag.String("host", "gitlab.com", "GitLab host address")
		token         = flag.String("token", "", "GitLab API token")
		projectID     = flag.Int("project", 1, "GitLab project id")
		reviewersPath = flag.String("reviewers", "reviewers.json", "file path to the reviewers config file")
		webhook       = flag.String("webhook", "", "Mattermost webhook URL")
		channel       = flag.String("channel", "", "Mattermost channel (e.g. MyChannel) or user (e.g. @AnyUser)")
	)
	flag.Parse()

	// setup
	reviewers := loadReviewers(*reviewersPath)

	// aggregate
	reminder := gitlab.AggregateReminder(*host, *token, *projectID, reviewers, gitlab.DefaultTemplate())
	if reminder == "" {
		return
	}

	// output
	fmt.Println(reminder)

	if *channel != "" {
		mattermost.Send(*channel, reminder, *webhook)
	}
}

// load reviewers from given json file
// formatting: "GitLab UserID":"Mattermost Username"
// e.g. {"3":"@john.doe","5":"@max"}
func loadReviewers(reviewersPath string) map[int]string {
	b, err := ioutil.ReadFile(reviewersPath)
	if err != nil {
		log.Fatalf("failed to read reviewers file: %v", err)
	}

	// 'GitLab UserID': 'Mattermost Username'
	reviewers := map[int]string{}

	if err := json.Unmarshal(b, &reviewers); err != nil {
		log.Fatalf("failed to unmarshal reviewers: %v", err)
	}

	return reviewers
}
