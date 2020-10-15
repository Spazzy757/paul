package config

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/Spazzy757/paul/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestGetPaulConfig(t *testing.T) {
	t.Run("Test Read Paul Config Returns Valid Paul Config", func(t *testing.T) {
		yamlFile, err := ioutil.ReadFile("../../PAUL.yaml")
		assert.Equal(t, nil, err)

		mClient := test.GetMockClient()
		mClient.RepoService = &test.MockRepoService{
			DownloadContentsResp: ioutil.NopCloser(bytes.NewReader(yamlFile)),
		}

		owner := "test"
		repo := "test"
		cfg, err := GetPaulConfig(&owner, &repo, "example.com", "main", mClient)
		assert.Equal(t, nil, err)
		assert.NotEqual(t, "", cfg.PullRequests.OpenMessage)
		assert.NotEqual(t, cfg.PullRequests.CatsEnabled, false)
		assert.NotEqual(t, cfg.PullRequests.DogsEnabled, false)
	})
}
