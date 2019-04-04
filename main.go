package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/sj14/review-reminder/gitlab"
	"github.com/sj14/review-reminder/mattermost"
	"github.com/sj14/review-reminder/templates"
)

func main() {
	var (
		host          = flag.String("host", "gitlab.com", "GitLab host address")
		glToken       = flag.String("token", "", "GitLab API token")
		projectID     = flag.Int("project", 1, "GitLab project id")
		reviewersPath = flag.String("reviewers", "reviewers.json", "file path to the reviewers config file")
		webhook       = flag.String("webhook", "", "Mattermost webhook URL")
		channel       = flag.String("channel", "", "Mattermost channel (e.g. MyChannel) or user (e.g. @AnyUser)")
	)
	flag.Parse()

	// setup gitlab client
	git := gitlab.NewClient(*host, *glToken)

	// load reviewers from given json file
	reviewers := loadReviewers(*reviewersPath)

	// get open merge requests
	mergeRequests := gitlab.OpenMergeRequests(git, *projectID)

	// only return merge requests which have no open discussions
	// mergeRequests = filterOpenDiscussions(git, mergeRequests)

	// parse the reminder text template
	template := templates.Get()

	// will contain the reminders of all merge requests
	var reminderText string

	for _, mr := range mergeRequests {
		// dont' check WIP MRs
		if mr.WorkInProgress {
			continue
		}

		// load all emojis awarded to the mr
		emojis := gitlab.LoadEmojis(git, *projectID, mr)

		// check who gave thumbs up/down (or "sleeping")
		reviewedBy := gitlab.GetReviewed(mr, emojis)

		// who is missing thumbs up/down
		missing := gitlab.MissingReviewers(reviewedBy, reviewers)

		// load all discussions of the mr
		discussions := gitlab.LoadDiscussions(git, *projectID, mr)

		// get the number of open discussions
		discussionsCount := gitlab.OpenDiscussionsCount(discussions)

		// get the responsible person of the mr
		owner := gitlab.ResponsiblePerson(mr, reviewers)

		// list each emoji with the usage count
		emojisAggr := gitlab.AggregateEmojis(emojis)

		// generate the reminder text for the current mr
		reminderText += templates.Exec(template, mr, owner, missing, discussionsCount, emojisAggr)
	}

	// print text of all aggregated reminders
	fmt.Println(reminderText)

	if *channel != "" {
		mattermost.Send(*channel, reminderText, *webhook)
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
		log.Fatalf("failed to umarshal reviewers: %v", err)
	}

	return reviewers
}
