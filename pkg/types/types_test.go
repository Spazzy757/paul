package types

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	var paulConfig PaulConfig

	yamlFile, err := ioutil.ReadFile("../../PAUL.yaml")
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
	t.Run("Test Loading Config - Labels", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.Labels, false)
	})
	t.Run("Test Loading Config - Branch Destroyer", func(t *testing.T) {
		assert.Equal(t, true, paulConfig.BranchDestroyer.Enabled)
		assert.Equal(t, []string{"main"}, paulConfig.BranchDestroyer.ProtectedBranches)
	})

}

func TestLoadConfigFails(t *testing.T) {
	// Only run actual test in subprocess
	var paulConfig PaulConfig
	err := paulConfig.LoadConfig([]byte(`% ^ & HHH`))
	assert.NotEqual(t, nil, err)
}
