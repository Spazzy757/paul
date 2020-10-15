package github

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/Spazzy757/paul/pkg/test"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

func getMockPayload() []byte {
	file, _ := ioutil.ReadFile("../../mocks/pr.json")
	return []byte(file)
}

func TestCreateReview(t *testing.T) {
	t.Run("Test Webhook is Handled correctly", func(t *testing.T) {
		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")
		mClient := test.GetMockClient()

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := reviewComment(e.PullRequest, mClient, "test")
		assert.Equal(t, nil, err)
	})
}

func TestFirstPRCheck(t *testing.T) {
	t.Run("Test First PR - no message", func(t *testing.T) {
		firstPR := firstPRCheck("", "opened")
		assert.Equal(t, false, firstPR)
	})
	t.Run("Test First PR - should be true", func(t *testing.T) {
		firstPR := firstPRCheck("Test", "opened")
		assert.Equal(t, true, firstPR)
	})
	t.Run("Test First PR - should Be false", func(t *testing.T) {
		firstPR := firstPRCheck("Test", "closed")
		assert.Equal(t, false, firstPR)
	})
}

func TestBranchDestroyer(t *testing.T) {
	t.Run("Test Webhook is Handled correctly", func(t *testing.T) {
		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")
		mClient := test.GetMockClient()

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := branchDestroyer(e.PullRequest, mClient, "test")
		assert.Equal(t, nil, err)
	})
}

func TestBranchDestroyerCheck(t *testing.T) {
	protected := []string{"main", "master"}
	cfg := &types.BranchDestroyer{
		Enabled:           true,
		ProtectedBranches: protected,
	}
	t.Run("Test Branch Destroyer - not enabled", func(t *testing.T) {
		disabledCfg := &types.BranchDestroyer{
			Enabled:           false,
			ProtectedBranches: protected,
		}
		destroyBranch := branchDestroyerCheck(
			disabledCfg,
			"completed",
			"main",
			"feature",
		)
		assert.Equal(t, false, destroyBranch)
	})
	t.Run("Test Branch Destroyer - not completed", func(t *testing.T) {
		destroyBranch := branchDestroyerCheck(
			cfg,
			"opened",
			"main",
			"feature",
		)
		assert.Equal(t, false, destroyBranch)
	})
	t.Run("Test Branch Destroyer - default branch", func(t *testing.T) {
		destroyBranch := branchDestroyerCheck(
			cfg,
			"completed",
			"main",
			"main",
		)
		assert.Equal(t, false, destroyBranch)
	})
	t.Run("Test Branch Destroyer - protected branch", func(t *testing.T) {
		destroyBranch := branchDestroyerCheck(
			cfg,
			"completed",
			"main",
			"master",
		)
		assert.Equal(t, false, destroyBranch)
	})
	t.Run("Test Branch Destroyer - valid", func(t *testing.T) {
		destroyBranch := branchDestroyerCheck(
			cfg,
			"completed",
			"main",
			"feature",
		)
		assert.Equal(t, true, destroyBranch)
	})

}
