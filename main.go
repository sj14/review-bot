package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"text/template"

	"github.com/sj14/review-bot/hoster/github"
	"github.com/sj14/review-bot/hoster/gitlab"
	"github.com/sj14/review-bot/mattermost"
)

func main() {
	var (
		host          = flag.String("host", "", "host address (e.g. github.com, gitlab.com or self-hosted gitlab url")
		token         = flag.String("token", "", "host API token")
		repo          = flag.String("repo", "", "github repository (format: 'owner/repo')")
		projectID     = flag.Int("project", 0, "gitlab project id")
		reviewersPath = flag.String("reviewers", "examples/reviewers.json", "path to the reviewers file")
		templatePath  = flag.String("template", "", "path to the template file")
		webhook       = flag.String("webhook", "", "Mattermost webhook URL")
		channel       = flag.String("channel", "", "Mattermost channel (e.g. MyChannel) or user (e.g. @AnyUser)")
	)
	flag.Parse()

	if *host == "" {
		log.Fatalln("missing host")
	}

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
		if *repo == "" {
			log.Fatalln("missing github repository")
		}

		ownerRespo := strings.SplitN(*repo, "/", 2)
		if len(ownerRespo) != 2 {
			log.Fatalln("wrong repo format (use 'owner/repo')")
		}
		reminder = github.AggregateReminder(*token, ownerRespo[0], ownerRespo[1], reviewers, tmpl)
	} else {
		if *projectID == 0 {
			log.Fatalln("missing gitlab project id")
		}
		reminder = gitlab.AggregateReminder(*host, *token, *projectID, reviewers, tmpl)
	}

	if reminder == "" {
		return
	}

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
