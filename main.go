package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/sj14/review-bot/hoster/github"
	"github.com/sj14/review-bot/hoster/gitlab"
	"github.com/sj14/review-bot/slackermost"
	"github.com/spf13/pflag"
)

func main() {
	var (
		// Repository flags, available for all subcommands.
		repoFlags     = pflag.NewFlagSet("repo", pflag.ExitOnError)
		githost       = repoFlags.String("host", "", "host address (e.g. github.com, gitlab.com or self-hosted gitlab url)")
		token         = repoFlags.String("token", "", "host API token")
		repo          = repoFlags.String("repo", "", "repository (format: 'owner/repo'), or project id (only gitlab)")
		reviewersPath = repoFlags.String("reviewers", "examples/reviewers.json", "path to the reviewers file")
		templatePath  = repoFlags.String("template", "", "path to the template file")

		// Slack/Mattermost flags
		slackermostFlags = pflag.NewFlagSet("slackermost", pflag.ExitOnError)
		webhook          = slackermostFlags.String("webhook", "", "slack/mattermost webhook URL")
		channelOrUser    = slackermostFlags.String("channel", "", "mattermost channel (e.g. MyChannel) or user (e.g. @AnyUser)")

		// E-Mail flags
		mailFlags = pflag.NewFlagSet("mail", pflag.ExitOnError)
		// mailhost  = mailFlags.String("mailhost", "", "...")
		// user      = mailFlags.String("user", "", "...")
		// pass      = mailFlags.String("pass", "", "...")
	)

	repoFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Available subcommands:\n\tslackermost | mail\n")
		fmt.Fprintf(os.Stderr, "\tUse 'subcommand --help' for all flags of the specified command.\n")
		fmt.Fprintf(os.Stderr, "Generic flags for all subcommands:\n")
		repoFlags.PrintDefaults()
	}

	log.Printf("got arg: %v", os.Args[1])

	switch os.Args[1] {
	case "slackermost":
		slackermostFlags.AddFlagSet(repoFlags)
		if err := slackermostFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse slackermost flags: %v", err)
		}
		log.Println("called slackermost subcommand")
	case "mail":
		mailFlags.AddFlagSet(repoFlags)
		if err := mailFlags.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse mail flags: %v", err)
		}
		log.Println("called mail subcommand")
	default:
		if err := repoFlags.Parse(os.Args[1:]); err != nil {
			log.Fatalf("failed to parse default flags: %v", err)
		}

		// Command not recognized. Print usage help and exit.
		repoFlags.Usage()
		os.Exit(1)
	}

	// No comamnd given. Print usage help and exit.
	if len(os.Args) < 2 {
		repoFlags.Usage()
		os.Exit(1)
	}

	if *githost == "" {
		log.Fatalln("missing host")
	}
	if *repo == "" {
		log.Fatalln("missing repository")
	}

	reviewers := loadReviewers(*reviewersPath)

	var tmpl *template.Template
	if *templatePath != "" {
		tmpl = loadTemplate(*templatePath)
	} else if *githost == "github.com" {
		tmpl = github.DefaultTemplate()
	} else {
		tmpl = gitlab.DefaultTemplate()
	}

	var reminder string
	if *githost == "github.com" {
		ownerRespo := strings.SplitN(*repo, "/", 2)
		if len(ownerRespo) != 2 {
			log.Fatalln("wrong repo format (use 'owner/repo')")
		}
		repo, reminders := github.AggregateReminder(*token, ownerRespo[0], ownerRespo[1], reviewers)
		if len(reminders) == 0 {
			// prevent from sending the header only
			return
		}
		reminder = github.ExecTemplate(tmpl, repo, reminders)

	} else {
		project, reminders := gitlab.AggregateReminder(*githost, *token, *repo, reviewers)
		if len(reminders) == 0 {
			// prevent from sending the header only
			return
		}
		reminder = gitlab.ExecTemplate(tmpl, project, reminders)
	}

	if reminder == "" {
		return
	}

	fmt.Println(reminder)

	if *webhook != "" {
		if err := slackermost.Send(*channelOrUser, reminder, *webhook); err != nil {
			log.Fatalf("failed sending slackermost message: %v", err)
		}
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
