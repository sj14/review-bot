package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"text/template"

	"github.com/sj14/review-bot/mattermost"
	"github.com/sj14/review-bot/review/github"
	"github.com/sj14/review-bot/review/gitlab"
)

func main() {
	var (
		host          = flag.String("host", "gitlab.com", "GitLab host address")
		token         = flag.String("token", "", "host API token")
		projectID     = flag.Int("project", 1, "GitLab project id")
		reviewersPath = flag.String("reviewers", "examples/reviewers.json", "path to the reviewers file")
		templatePath  = flag.String("template", "", "path to the template file")
		webhook       = flag.String("webhook", "", "Mattermost webhook URL")
		channel       = flag.String("channel", "", "Mattermost channel (e.g. MyChannel) or user (e.g. @AnyUser)")
	)
	flag.Parse()

	// setup
	reviewers := loadReviewers(*reviewersPath)

	var tmpl *template.Template
	if *templatePath != "" {
		tmpl = loadTemplate(*templatePath)
	} else if *host == "github.com" {
		tmpl = github.DefaultTemplate()
	} else {
		tmpl = gitlab.DefaultTemplate()
	}

	var reminder string
	if *host == "github.com" {
		reminder = github.AggregateReminder(*token, "kubernetes", "enhancements", reviewers, tmpl)
	} else {
		reminder = gitlab.AggregateReminder(*host, *token, *projectID, reviewers, tmpl)
	}

	if reminder == "" {
		return
	}

	// output
	fmt.Println(reminder)

	if *channel != "" {
		mattermost.Send(*channel, reminder, *webhook)
	}
}

func loadTemplate(path string) *template.Template {
	t, err := template.ParseFiles(path)
	if err != nil {
		log.Fatalf("failed to read template file: %v", err)
	}
	return t
}

// load reviewers from given json file
// formatting:
// "GitLab UserID":"Mattermost Username"
// e.g. {"3":"@john.doe","5":"@max"}
// or
// 'Github LoginName': 'Mattermost Username'
// e.g. {"sj14":"@simon","john":"@john"}
func loadReviewers(path string) map[string]string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read reviewers file: %v", err)
	}

	reviewers := map[string]string{}
	if err := json.Unmarshal(b, &reviewers); err != nil {
		log.Fatalf("failed to unmarshal reviewers: %v", err)
	}

	return reviewers
}
