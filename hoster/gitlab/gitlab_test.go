package gitlab

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestAggregateReminder(t *testing.T) {
	mockedClient := &clientWrapperMock{
		loadProjectFunc: func(repo interface{}) gitlab.Project {
			return gitlab.Project{Name: "mocked project"}
		},
		loadMRsFunc: func(repo interface{}) []*gitlab.MergeRequest {
			return []*gitlab.MergeRequest{
				{Title: "MR0"},
				{Title: "MR1", WorkInProgress: true},
			}
		},
		loadEmojisFunc: func(repo interface{}, mr *gitlab.MergeRequest) []*gitlab.AwardEmoji {
			return []*gitlab.AwardEmoji{
				{Name: thumbsup},
			}
		},
		loadDiscussionsFunc: func(repo interface{}, mr *gitlab.MergeRequest) []*gitlab.Discussion {
			return []*gitlab.Discussion{
				{ID: "id0", Notes: []*gitlab.Note{{Resolved: false, Resolvable: true}}},
			}
		},
	}

	expP := gitlab.Project{
		Name: "mocked project",
	}

	expR := []reminder{
		{MR: &gitlab.MergeRequest{Title: "MR0"}, Missing: []string{"Spidy"}, Emojis: map[string]int{"thumbsup": 1}, Discussions: 1},
	}

	gotP, gotR := aggregate(mockedClient, 2009901, map[string]string{"42": "Spidy"})

	require.Equal(t, expP, gotP)
	require.Equal(t, expR, gotR)
}

func TestResponsiblePerson(t *testing.T) {
	t.Run("author", func(t *testing.T) {
		mr := &gitlab.MergeRequest{Author: &gitlab.BasicUser{Name: "name-of-author"}}
		reviewers := map[string]string{}
		got := responsiblePerson(mr, reviewers)
		require.Equal(t, "name-of-author", got)
	})

	t.Run("@author", func(t *testing.T) {
		mr := &gitlab.MergeRequest{Author: &gitlab.BasicUser{Username: "gitlab_name"}}
		reviewers := map[string]string{"gitlab_name": "@author-of-mr"}
		got := responsiblePerson(mr, reviewers)
		require.Equal(t, "@author-of-mr", got)
	})

	t.Run("assignee", func(t *testing.T) {
		mr := &gitlab.MergeRequest{Assignee: &gitlab.BasicUser{Username: "gitlab_name"}}
		reviewers := map[string]string{"gitlab_name": "assignee-of-mr"}
		got := responsiblePerson(mr, reviewers)
		require.Equal(t, "assignee-of-mr", got)
	})
}

func TestGetReviewed(t *testing.T) {
	mr := &gitlab.MergeRequest{Author: &gitlab.BasicUser{Username: "mr_author"}}

	type user struct {
		Name      string `json:"name"`
		Username  string `json:"username"`
		ID        int    `json:"id"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	}

	emojis := []*gitlab.AwardEmoji{
		{Name: thumbsup, User: user{Username: "user0"}},
		{Name: thumbsdown, User: user{Username: "user1"}},
		{Name: sleeping, User: user{Username: "user2"}},
		{Name: "hooray", User: user{Username: "user3"}},
		{Name: thumbsup, User: user{Username: "user3"}},
		{Name: "anyemoji", User: user{Username: "user4"}},
	}

	got := getReviewed(mr, emojis)

	want := []string{"mr_author", "user0", "user1", "user2", "user3"}
	require.Equal(t, want, got)
}

func TestMissingReviewers(t *testing.T) {
	reviewedBy := []string{"user1", "user2"}

	approvers := map[string]string{
		"user0": "@user0",
		"user1": "@user1",
		"user2": "@user2",
		"user3": "@user3",
	}

	got := missingReviewers(reviewedBy, approvers)

	want := []string{"@user0", "@user3"}
	require.ElementsMatch(t, want, got)
}

func TestAggregateEmojis(t *testing.T) {
	input := []*gitlab.AwardEmoji{
		{Name: "emoji0"},
		{Name: "emoji0"},
		{Name: "emoji0"},
		{Name: "emoji1"},
		{Name: "emoji1"},
		{Name: "emoji2"},
	}

	got := aggregateEmojis(input)

	want := map[string]int{"emoji0": 3, "emoji1": 2, "emoji2": 1}
	require.Equal(t, want, got)
}
