package github

import (
	"testing"

	"github.com/google/go-github/v25/github"
	"github.com/stretchr/testify/require"
)

func stringp(s string) *string {
	return &s
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
