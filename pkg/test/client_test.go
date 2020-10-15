package test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockClient(t *testing.T) {
	mclient := GetMockClient()
	t.Run("Test Mock GitService", func(t *testing.T) {
		expected := &MockGitService{}
		assert.Equal(t,
			reflect.TypeOf(mclient.GitService),
			reflect.TypeOf(expected),
		)
	})
	t.Run("Test Mock RepoService", func(t *testing.T) {
		expected := &MockRepoService{}
		assert.Equal(t,
			reflect.TypeOf(mclient.RepoService),
			reflect.TypeOf(expected),
		)
	})
	t.Run("Test Mock PullRequestService", func(t *testing.T) {
		expected := &MockPullRequestService{}
		assert.Equal(t,
			reflect.TypeOf(mclient.PullRequestService),
			reflect.TypeOf(expected),
		)
	})
	t.Run("Test Mock IssueService", func(t *testing.T) {
		expected := &MockIssueService{}
		assert.Equal(t,
			reflect.TypeOf(mclient.IssueService),
			reflect.TypeOf(expected),
		)
	})
}
