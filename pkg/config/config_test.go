package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/Spazzy757/paul/pkg/test"
	"github.com/stretchr/testify/assert"
)

const (
	// baseURLPath is a non-empty Client.BaseURL path to use during tests,
	// to ensure relative URLs are used for all endpoints. See issue #752.
	baseURLPath = "/api-v3"
)

func TestGetPaulConfig(t *testing.T) {
	t.Run("Test Read Paul Config Returns Valid Paul Config", func(t *testing.T) {
		yamlFile, err := ioutil.ReadFile("../../PAUL.yaml")
		assert.Equal(t, nil, err)

		mClient, mux, serverURL, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/contents/",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				fmt.Fprint(w, `[{
		            "type": "file",
		            "name": "PAUL.yaml",
		            "download_url": "`+serverURL+baseURLPath+`/download/PAUL.yaml"
		        }]`)
			},
		)
		mux.HandleFunc("/download/PAUL.yaml", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, string(yamlFile))
		})

		owner := "Spazzy757"
		repo := "paul"
		cfg, err := GetPaulConfig(
			context.Background(),
			&owner, &repo,
			"PAUL.yaml",
			"main",
			mClient,
		)
		assert.Equal(t, nil, err)
		assert.NotEqual(t, "", cfg.PullRequests.OpenMessage)
		assert.NotEqual(t, cfg.PullRequests.CatsEnabled, false)
		assert.NotEqual(t, cfg.PullRequests.DogsEnabled, false)
	})
}
