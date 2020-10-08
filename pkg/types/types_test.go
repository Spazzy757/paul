package types

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	var paulConfig PaulConfig

	yamlFile, err := ioutil.ReadFile("../../PAUL.yaml")
	if err != nil {
		t.Errorf("yamlFile.Get err   #%v ", err)
	}

	paulConfig.LoadConfig(yamlFile)
	t.Run("Test Loading Config - OpenMessage", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.PullRequests.OpenMessage, "")
	})
	t.Run("Test Loading Config - CatsEnabled", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.PullRequests.CatsEnabled, false)
	})
	t.Run("Test Loading Config - DogsEnabled", func(t *testing.T) {
		assert.NotEqual(t, paulConfig.PullRequests.DogsEnabled, false)
	})

}
