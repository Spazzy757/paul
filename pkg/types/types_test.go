package types

import (
	"io/ioutil"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Test Loading Config Works", func(t *testing.T) {
		var paulConfig PaulConfig

		yamlFile, err := ioutil.ReadFile("../../PAUL.yaml")
		if err != nil {
			t.Errorf("yamlFile.Get err   #%v ", err)
		}

		paulConfig.LoadConfig(yamlFile)
		if paulConfig.PullRequests.OpenMessage == "" {
			t.Errorf("Expected LoadConfig to Load Data: %v ", paulConfig.PullRequests.OpenMessage)
		}
	})
}
