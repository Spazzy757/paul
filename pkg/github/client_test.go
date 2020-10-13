package github

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

type mockRepoClient struct {
	resp io.ReadCloser
}

func (m *mockRepoClient) DownloadContents(
	ctx context.Context,
	owner, repo, filepath string,
	opt *github.RepositoryContentGetOptions,
) (io.ReadCloser, error) {
	return m.resp, nil
}

func TestGetPaulConfig(t *testing.T) {
	t.Run("Test Read Paul Config Returns Valid Paul Config", func(t *testing.T) {
		yamlFile, err := ioutil.ReadFile("../../PAUL.yaml")
		assert.Equal(t, nil, err)
		ctx := context.Background()

		mc := &mockRepoClient{
			resp: ioutil.NopCloser(bytes.NewReader(yamlFile)),
		}
		rc := &repoClient{ctx: ctx, client: mc}

		owner := "test"
		repo := "test"
		cfg, err := getPaulConfig(&owner, &repo, "example.com", rc)
		assert.Equal(t, nil, err)
		assert.NotEqual(t, "", cfg.PullRequests.OpenMessage)
		assert.NotEqual(t, cfg.PullRequests.CatsEnabled, false)
		assert.NotEqual(t, cfg.PullRequests.DogsEnabled, false)
	})
}
