package github

import (
	"testing"

	"github.com/google/go-github/v25/github"
	"github.com/stretchr/testify/require"
)

func stringp(s string) *string {
	return &s
}

func TestGetReviewed(t *testing.T) {
	pr := &github.PullRequest{User: &github.User{Login: stringp("owner")},
		RequestedReviewers: []*github.User{
			{Login: stringp("reviewer0")},
			{Login: stringp("reviewer1")},
			{Login: stringp("reviewer2")},
		}}

	reviews := []*github.PullRequestReview{
		{User: &github.User{Login: stringp("reviewer0")}, State: stringp(approved)},
		{User: &github.User{Login: stringp("reviewer1")}, State: stringp(dismissed)},
		{User: &github.User{Login: stringp("reviewer2")}},                           // requested but missing
		{User: &github.User{Login: stringp("reviewer3")}, State: stringp(approved)}, // not requested
	}

	want := []string{"reviewer0", "reviewer1", "reviewer3"}
	got := getReviewed(pr, reviews)
	require.Equal(t, want, got)
}

func TestMissingReviewers(t *testing.T) {
	requested := []*github.User{
		{Login: stringp("user0")},
		{Login: stringp("user1")},
		{Login: stringp("user2")},
		{Login: stringp("user3")},
	}

	reviewedBy := []string{"user1", "user2"}

	mapping := map[string]string{
		"user0": "@user0",
		"user1": "@user1",
		"user2": "@user2",
		// "user3": "@user3", // Test for fallback to github login name on missing mapping
	}

	got := missingReviewers(requested, reviewedBy, mapping)

	want := []string{"@user0", "user3"}
	require.ElementsMatch(t, want, got)
}

func TestIsRequestedReviewer(t *testing.T) {
	reviewers := []*github.User{
		{Login: stringp("user0")},
		{Login: stringp("user1")},
		{Login: stringp("user2")},
		{Login: stringp("user3")},
	}

	t.Run("mapping", func(t *testing.T) {
		requested := &github.User{Login: stringp("user2")}
		got := isRequestedReviewer(reviewers, requested)
		require.Equal(t, true, got)
	})
	t.Run("fallback", func(t *testing.T) {
		requested := &github.User{Login: stringp("unknown")}
		got := isRequestedReviewer(reviewers, requested)
		require.Equal(t, false, got)
	})
}

func TestResponsiblePerson(t *testing.T) {
	reviewers := map[string]string{
		"user0": "@user0",
		"user1": "@user1",
		"user2": "@user2",
		"user3": "@user3",
	}

	t.Run("mapping", func(t *testing.T) {
		pr := &github.PullRequest{User: &github.User{
			Login: stringp("user2"),
		}}
		got := responsiblePerson(pr, reviewers)
		require.Equal(t, "@user2", got)
	})
	t.Run("fallback", func(t *testing.T) {
		pr := &github.PullRequest{User: &github.User{
			Login: stringp("unkown"),
		}}
		got := responsiblePerson(pr, reviewers)
		require.Equal(t, "unkown", got)
	})
}
