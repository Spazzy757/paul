package types

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	var paulConfig PaulConfig

	yamlFile, err := ioutil.ReadFile("../../.github/PAUL.yaml")
	assert.Equal(t, nil, err)

	err = paulConfig.LoadConfig(yamlFile)
	assert.Equal(t, nil, err)
	t.Run("Test Loading Config - OpenMessage", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.PullRequests.OpenMessage, "")
	})
	t.Run("Test Loading Config - CatsEnabled", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.PullRequests.CatsEnabled, false)
	})
	t.Run("Test Loading Config - DogsEnabled", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.PullRequests.DogsEnabled, false)
	})
	t.Run("Test Loading Config - AllowApproval", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.PullRequests.AllowApproval, false)
	})
	t.Run("Test Loading Config - Labels", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.Labels, false)
	})
	t.Run("Test Loading Config - LimitPullRequests", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.PullRequests.LimitPullRequests.MaxNumber, nil)
	})
	t.Run("Test Loading Config - EmptyDescriptionCheck", func(t *testing.T) {
		assert.Equal(t, paulConfig.EmptyDescriptionCheck.Enabled, true)
		assert.Equal(t, paulConfig.EmptyDescriptionCheck.Enforced, true)
	})
	t.Run("Test Loading Config - Branch Destroyer", func(t *testing.T) {
		assert.Equal(t, true, paulConfig.BranchDestroyer.Enabled)
		assert.Equal(t, []string{"main"}, paulConfig.BranchDestroyer.ProtectedBranches)
	})
	t.Run("Test Loading Config - Stale Time", func(t *testing.T) {
		assert.NotEqual(t, 0, paulConfig.PullRequests.StaleTime)
	})
	t.Run("Test Loading Config - Automated Merge", func(t *testing.T) {
		assert.Equal(t, true, paulConfig.PullRequests.AutomatedMerge)
	})
	t.Run("Test Loading Config - GiphyEnabled", func(t *testing.T) {
		assert.Equal(t, true, paulConfig.PullRequests.GiphyEnabled)
	})
	t.Run("Test Loading Config - DCOCheckEnabled", func(t *testing.T) {
		assert.Equal(t, true, paulConfig.PullRequests.DCOCheck)
	})
	t.Run("Test Loading Config - VerifiedCommitCheckEnabled", func(t *testing.T) {
		assert.Equal(t, true, paulConfig.PullRequests.VerifiedCommitCheck)
	})
	t.Run("Test Loading Config - Assign", func(t *testing.T) {
		assert.Equal(t, true, paulConfig.PullRequests.Assign)
	})
}

func TestLoadConfigFails(t *testing.T) {
	// Only run actual test in subprocess
	var paulConfig PaulConfig
	err := paulConfig.LoadConfig([]byte(`% ^ & HHH`))
	assert.NotEqual(t, nil, err)
}
