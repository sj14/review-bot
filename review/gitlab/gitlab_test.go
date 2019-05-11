package gitlab

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestAggregateReminder(t *testing.T) {
	mockedClient := &clientMock{
		projectInfoFunc: func(id int) gitlab.Project {
			return gitlab.Project{Name: "mocked project"}
		},
		openMergeRequestsFunc: func(projectID int) []*gitlab.MergeRequest {
			return []*gitlab.MergeRequest{
				{Title: "MR0"},
			}
		},
		loadEmojisFunc: func(projectID int, mr *gitlab.MergeRequest) []*gitlab.AwardEmoji {
			return []*gitlab.AwardEmoji{
				{Name: ":thumbsup:"},
			}
		},
		loadDiscussionsFunc: func(projectID int, mr *gitlab.MergeRequest) []*gitlab.Discussion {
			return []*gitlab.Discussion{
				{ID: "id0", Notes: []*gitlab.Note{{Resolved: false, Resolvable: true}}},
			}
		},
	}

	expP := gitlab.Project{
		Name: "mocked project",
	}

	expR := []reminder{
		{MR: &gitlab.MergeRequest{Title: "MR0"}, Missing: []string{"Spidy"}, Emojis: map[string]int{":thumbsup:": 1}, Discussions: 1},
	}

	gotP, gotR := aggregate(mockedClient, 2009901, map[int]string{42: "Spidy"})

	require.Equal(t, expP, gotP)
	require.Equal(t, expR, gotR)
}

func TestResponsiblePerson(t *testing.T) {
	t.Run("author", func(t *testing.T) {
		mr := &gitlab.MergeRequest{
			Author: struct {
				ID        int        `json:"id"`
				Username  string     `json:"username"`
				Name      string     `json:"name"`
				State     string     `json:"state"`
				CreatedAt *time.Time `json:"created_at"`
			}{
				Name: "name-of-author",
			},
		}

		reviewers := map[int]string{}
		got := responsiblePerson(mr, reviewers)
		require.Equal(t, "name-of-author", got)
	})

	t.Run("@author", func(t *testing.T) {
		mr := &gitlab.MergeRequest{
			Author: struct {
				ID        int        `json:"id"`
				Username  string     `json:"username"`
				Name      string     `json:"name"`
				State     string     `json:"state"`
				CreatedAt *time.Time `json:"created_at"`
			}{
				ID: 5,
			},
		}

		reviewers := map[int]string{5: "@author-of-mr"}
		got := responsiblePerson(mr, reviewers)
		require.Equal(t, "@author-of-mr", got)
	})

	t.Run("assignee", func(t *testing.T) {
		mr := &gitlab.MergeRequest{
			Assignee: struct {
				ID        int        `json:"id"`
				Username  string     `json:"username"`
				Name      string     `json:"name"`
				State     string     `json:"state"`
				CreatedAt *time.Time `json:"created_at"`
			}{
				ID: 5,
			},
		}

		reviewers := map[int]string{5: "assignee-of-mr"}
		got := responsiblePerson(mr, reviewers)
		require.Equal(t, "assignee-of-mr", got)
	})
}
