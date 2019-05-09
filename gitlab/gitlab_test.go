package gitlab

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestAggregateReminder(t *testing.T) {
	mockedClient := &ClientMock{
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
