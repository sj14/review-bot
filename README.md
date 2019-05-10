# Review Reminder Bot

[![Build Status](https://dev.azure.com/SimonJuergensmeyer/SimonJuergensmeyer/_apis/build/status/sj14.review-bot?branchName=master)](https://dev.azure.com/SimonJuergensmeyer/SimonJuergensmeyer/_build/latest?definitionId=2&branchName=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/sj14/review-bot)](https://goreportcard.com/report/github.com/sj14/review-bot)

## Installation

```text
go install github.com/sj14/review-bot
```


## Example

### Sample Output

># [Project Name](https://gitlab.com/my_user/my_project)
>
>**How-To**: *Got reminded? Just normally review the given merge request with 👍/👎 or use 😴 if you don't want to receive a reminder about this merge request.*
>
>---
>
>**[Support SHIELD](https://gitlab.com/my_user/my_project/>merge_requests/1940)**  
> 1 💬   3 👍  @hulk
>
>**[Ask Deadpool to join us](https://gitlab.com/my_user/>my_project/merge_requests/1923)**  
> 3 💬   3 👍  @batman
>
>**[Repair the Helicarrier](https://gitlab.com/my_user/>my_project/merge_requests/1777)**  
> 3 💬   @hulk @batman @groot @iron_man
>
>**[Find Kingpin](https://gitlab.com/my_user/my_project/>merge_requests/1099)**  
> 2 💬   7 👍  You got all reviews, @daredevil.

### Configuration

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

### Running

Get all open merge requests from the project with ID `1` and post the resulting reminder to the specified Mattermost channel:

``` console
review-bot -host=$GITLAB_HOST -token=$GITLAB_API_TOKEN -project=1 -webhook=$WEBHOOK_ADDRESS -channel=$MATTERMOST_CHANNEL
```

## Command Line Flags

``` text
  -channel string
        Mattermost channel (e.g. MyChannel) or user (e.g. @AnyUser)
  -host string
        GitLab host address (default "gitlab.com")
  -project int
        GitLab project id (default 1)
  -reviewers string
        path to the reviewers file (default "examples/reviewers.json")
  -template string
        path to the template file
  -token string
        GitLab API token
  -webhook string
        Mattermost webhook URL
```
