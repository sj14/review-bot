package gitlab

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/api/client-go/v2"
)

func TestAggregateReminder(t *testing.T) {
	mockedClient := &clientWrapperMock{
		loadProjectFunc: func(repo interface{}) (gitlab.Project, error) {
			return gitlab.Project{Name: "mocked project"}, nil
		},
		loadMRsFunc: func(repo interface{}) ([]*gitlab.BasicMergeRequest, error) {
			return []*gitlab.BasicMergeRequest{
				{Title: "MR0"},
				{Title: "MR1", Draft: true},
			}, nil
		},
		loadEmojisFunc: func(repo interface{}, mr *gitlab.BasicMergeRequest) ([]*gitlab.AwardEmoji, error) {
			return []*gitlab.AwardEmoji{
				{Name: thumbsup},
			}, nil
		},
		loadDiscussionsFunc: func(repo interface{}, mr *gitlab.BasicMergeRequest) ([]*gitlab.Discussion, error) {
			return []*gitlab.Discussion{
				{ID: "id0", Notes: []*gitlab.Note{{Resolved: false, Resolvable: true}}},
			}, nil
		},
	}

	expP := gitlab.Project{
		Name: "mocked project",
	}

	expR := []reminder{
		{MR: &gitlab.BasicMergeRequest{Title: "MR0"}, Missing: []string{"Spidy"}, Emojis: map[string]int{"thumbsup": 1}, Discussions: 1},
	}

	gotP, gotR, err := aggregate(mockedClient, 2009901, map[string]string{"42": "Spidy"})

	require.NoError(t, err)
	require.Equal(t, expP, gotP)
	require.Equal(t, expR, gotR)
}

func TestResponsiblePerson(t *testing.T) {
	t.Run("author", func(t *testing.T) {
		mr := &gitlab.BasicMergeRequest{Author: &gitlab.BasicUser{Name: "name-of-author"}}
		reviewers := map[string]string{}
		got := responsiblePerson(mr, reviewers)
		require.Equal(t, "name-of-author", got)
	})

	t.Run("@author", func(t *testing.T) {
		mr := &gitlab.BasicMergeRequest{Author: &gitlab.BasicUser{Username: "gitlab_name"}}
		reviewers := map[string]string{"gitlab_name": "@author-of-mr"}
		got := responsiblePerson(mr, reviewers)
		require.Equal(t, "@author-of-mr", got)
	})

	t.Run("assignee", func(t *testing.T) {
		mr := &gitlab.BasicMergeRequest{Assignee: &gitlab.BasicUser{Username: "gitlab_name"}}
		reviewers := map[string]string{"gitlab_name": "assignee-of-mr"}
		got := responsiblePerson(mr, reviewers)
		require.Equal(t, "assignee-of-mr", got)
	})
}

func TestGetReviewed(t *testing.T) {
	mr := &gitlab.BasicMergeRequest{Author: &gitlab.BasicUser{Username: "mr_author"}}

	emojis := []*gitlab.AwardEmoji{
		{Name: thumbsup, User: gitlab.BasicUser{Username: "user0"}},
		{Name: thumbsdown, User: gitlab.BasicUser{Username: "user1"}},
		{Name: sleeping, User: gitlab.BasicUser{Username: "user2"}},
		{Name: "hooray", User: gitlab.BasicUser{Username: "user3"}},
		{Name: thumbsup, User: gitlab.BasicUser{Username: "user3"}},
		{Name: "anyemoji", User: gitlab.BasicUser{Username: "user4"}},
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
