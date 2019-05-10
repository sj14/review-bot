# Review Reminder Bot

[![Build Status](https://dev.azure.com/SimonJuergensmeyer/SimonJuergensmeyer/_apis/build/status/sj14.review-bot?branchName=master)](https://dev.azure.com/SimonJuergensmeyer/SimonJuergensmeyer/_build/latest?definitionId=2&branchName=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/sj14/review-bot)](https://goreportcard.com/report/github.com/sj14/review-bot)

## Usage

### Command Line Flags

``` text
  -channel string
        Mattermost channel (e.g. MyChannel) or user (e.g. @AnyUser)
  -host string
        GitLab host address (default "gitlab.com")
  -project int
        GitLab project id (default 1)
  -reviewers string
        file path to the reviewers config file (default "reviewers.json")
  -token string
        GitLab API token
  -webhook string
        Mattermost webhook URL
```

### Examples

The reviewers.json file contains the `gitlab_user_id: "@mattermost_name"`.

```json
{
    "5": "@hulk",
    "17": "@iron_man",
    "92": "@groot",
    "95": "@batman",
    "123": "@daredevil"
}
```

Get all open merge requests from project 1 and post it to the specified Mattermost channel:

``` text
go run main.go -host=$GITLAB_HOST -token=$GITLAB_API_TOKEN -project=1 -webhook=$WEBHOOK_ADDRESS -channel=$MATTERMOST_CHANNEL
```

This will output the merge requests with the number of open discussions (💬) and the number of 👍 and 👎. The missing reviewers will be mentioned.  
Adding the "sleeping" 😴 emoji to a merge request means the user won't review the code and/or doesn't want to be mentioned.  
When all reviewers gave their thumps, the owner of the MR will be informed.

``` markdown
**[Support SHIELD](https://gitlab.com/my_user/my_project/merge_requests/1940)**  
 1 💬   3 :thumbsup:  @hulk

**[Ask Deadpool to join us](https://gitlab.com/my_user/my_project/merge_requests/1923)**  
 3 💬   3 :thumbsup:  @batman

**[Repair the Helicarrier](https://gitlab.com/my_user/my_project/merge_requests/1777)**  
 3 💬   @hulk @batman @groot @iron_man

**[Find Kingpin](https://gitlab.com/my_user/my_project/merge_requests/1099)**  
 2 💬   7 :thumbsup:  You got all reviews, @daredevil.
```