# Review Reminder Bot

For an enhanced SaaS version, visit https://www.review-bot.com/

---

![Action](https://github.com/sj14/dbbench/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/sj14/review-bot)](https://goreportcard.com/report/github.com/sj14/review-bot)

`review-bot` sends a reminder message to Mattermost or Slack with all open pull/merge requests which need an approval. Well suitable for running as a cron-job, e.g. for daily reminders.

This tool is still **beta**. The usage with Gitlab and Mattermost is more mature while the Github and Slack usage is an early preview.

## Installation

```text
go get -u github.com/sj14/review-bot
```

## Example

### Sample Output for Gitlab and Mattermost

># [Project Name](https://gitlab.com/my_user/my_project)
>
>**How-To**: *Got reminded? Just normally review the given merge request with 👍/👎 or use 😴 if you don't want to receive a reminder about this merge request.*
>
>---
>
>**[Support SHIELD](https://gitlab.com/my_user/my_project/merge_requests/1940)**  
> 1 💬   3 👍  @hulk
>
>**[Ask Deadpool to join us](https://gitlab.com/my_user/my_project/merge_requests/1923)**  
> 3 💬   3 👍  @batman
>
>**[Repair the Helicarrier](https://gitlab.com/my_user/my_project/merge_requests/1777)**  
> 3 💬   @hulk @batman @groot @iron_man
>
>**[Find Kingpin](https://gitlab.com/my_user/my_project/merge_requests/1099)**  
> 2 💬   4 👍  You got all reviews, @daredevil.

### Configuration

The `reviewers.json` file contains the gitlab/github user name as key and the mattermost name or slack [user id](https://api.slack.com/methods/users.identity) as value.

**Example 1**: github/gitlab username and mattermost name

```json
{
    "hulk51": "@hulk",
    "tonystark": "@iron_man",
    "groot": "@groot",
    "darkknight": "@batman",
    "lawyer": "@daredevil"
}
```

**Example 2**: github/gitlab username and slack id

```json
{
    "hulk51": "@U024BE7LH",
    "tonystark": "U0G9QF9C6",
    "groot": "@U0JA38A",
    "darkknight": "@U0QM9L4",
    "lawyer": "@U0JMB8O1"
}
```

### Running

Get all open merge requests from the Gitlab project `owner/repo` and post the resulting reminder to the specified Mattermost channel:

``` text
review-bot -host=$GITLAB_HOST -token=$GITLAB_API_TOKEN -repo=owner/repo -webhook=$WEBHOOK_ADDRESS -channel=$MATTERMOST_CHANNEL
```

## Command Line Flags

``` text
  -channel string
        mattermost channel (e.g. MyChannel) or user (e.g. @AnyUser)
  -host string
        host address (e.g. github.com, gitlab.com or self-hosted gitlab url)
  -repo string
        repository (format: 'owner/repo'), or project id (only gitlab)
  -reviewers string
        path to the reviewers file (default "examples/reviewers.json")
  -template string
        path to the template file
  -token string
        host API token
  -webhook string
        slack/mattermost webhook URL
```

## Templates

We use the Go [template](https://golang.org/pkg/text/template/) package for parsing.
Depending on which backend you use, there are different fields you can use. Check the [examples](https://github.com/sj14/review-bot/tree/master/examples) folder for a quick overview.

### Gitlab

Accessing `{{.Project}}` gives you access to these [fields](https://godoc.org/github.com/xanzy/go-gitlab#Project).  
While `{{range .Reminders}}` gives you access to `{{.MR}}` which is the [merge request](https://godoc.org/github.com/xanzy/go-gitlab#MergeRequest). `{{.Missing}}` is the Slack/Mattermost handle of the missing reviewer. `{{.Discussions}}` is the number of open discussion. `{{.Owner}}` is the Mattermost name of the assignee or otherwise the creator of the merge request. `{{.Emojis}}` is a map with the reacted emoji's and their count on this merge request.

The corresponding Go structs:

```go
type data struct {
      Project   gitlab.Project
      Reminders []reminder
}

type reminder struct {
      MR          *gitlab.MergeRequest
      Missing     []string
      Discussions int
      Owner       string
      Emojis      map[string]int
}
```


### Github

Accessing `{{.Repository}}` gives you access to these [fields](https://godoc.org/github.com/google/go-github/github#Repository).  
While `{{range .Reminders}}` gives you access to `{{.PR}}` which is the [pull request](https://godoc.org/github.com/google/go-github/github#PullRequest). `{{.Owner}}` the Mattermost name of the PR creator or the Github login as fallback. `{{.Missing}}` is the Slack/Mattermost handle of the missing reviewer.

```go
type data struct {
      Repository *github.Repository
      Reminders  []reminder
}

type reminder struct {
      PR          *github.PullRequest
      Missing     []string
      Owner       string
}
```
