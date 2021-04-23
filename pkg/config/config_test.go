package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/Spazzy757/paul/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// baseURLPath is a non-empty Client.BaseURL path to use during tests,
	// to ensure relative URLs are used for all endpoints. See issue #752.
	baseURLPath = "/api-v3"
)

func TestGetPaulConfig(t *testing.T) {
	t.Run("Test Read Paul Config Returns Valid Paul Config", func(t *testing.T) {
		assertions := require.New(t)
		yamlFile, err := ioutil.ReadFile("../../.github/PAUL.yaml")
		assertions.NoError(err)

		mClient, mux, serverURL, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/contents/",
			func(w http.ResponseWriter, r *http.Request) {
				assertions.Equal(r.Method, "GET")
				fmt.Fprint(w, `[{
		            "type": "file",
		            "name": "PAUL.yaml",
		            "download_url": "`+serverURL+baseURLPath+`/download/PAUL.yaml"
		        }]`)
			},
		)
		mux.HandleFunc("/download/PAUL.yaml", func(w http.ResponseWriter, r *http.Request) {
			assertions.Equal(r.Method, "GET")
			fmt.Fprint(w, string(yamlFile))
		})

		owner := "Spazzy757"
		repo := "paul"
		cfg, err := GetPaulConfig(
			context.Background(),
			owner, repo,
			"main",
			mClient,
		)
		assertions.NoError(err)
		assertions.NotEqual("", cfg.PullRequests.OpenMessage)
		assertions.NotEqual(cfg.PullRequests.CatsEnabled, false)
		assertions.NotEqual(cfg.PullRequests.DogsEnabled, false)
	})
	t.Run("Test Read Paul Config Returns Valid Paul Config if its in .github directory", func(t *testing.T) {
		assertions := require.New(t)
		yamlFile, err := ioutil.ReadFile("../../.github/PAUL.yaml")
		assertions.NoError(err)

		mClient, mux, serverURL, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/contents/.github",
			func(w http.ResponseWriter, r *http.Request) {
				assertions.Equal(r.Method, "GET")
				fmt.Fprint(w, `[{
		            "type": "dir",
		            "name": ".github",
		            "path": ".github"
		           },{
		            "type": "file",
		            "name": "PAUL.yaml",
		            "download_url": "`+serverURL+baseURLPath+`/download/PAUL.yaml"
		        }]`)
			},
		)
		mux.HandleFunc("/download/PAUL.yaml", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			log.Println("here")

			fmt.Fprint(w, string(yamlFile))
		})

		owner := "Spazzy757"
		repo := "paul"
		cfg, err := GetPaulConfig(
			context.Background(),
			owner, repo,
			"main",
			mClient,
		)
		assertions.NoError(err)
		assertions.NotEqual("", cfg.PullRequests.OpenMessage)
		assertions.NotEqual(cfg.PullRequests.CatsEnabled, false)
		assertions.NotEqual(cfg.PullRequests.DogsEnabled, false)
	})
	t.Run("Test Read Paul Error Returns an Empty Config but No Error", func(t *testing.T) {
		assertions := require.New(t)
		mClient, _, _, teardown := test.GetMockClient()
		defer teardown()

		owner := "Spazzy757"
		repo := "paul"
		cfg, err := GetPaulConfig(
			context.Background(),
			owner, repo,
			"main",
			mClient,
		)
		assertions.NoError(err)
		assertions.Equal("", cfg.PullRequests.OpenMessage)
		assertions.Equal(cfg.PullRequests.CatsEnabled, false)
		assertions.Equal(cfg.PullRequests.DogsEnabled, false)
	})
	t.Run("Test Read Paul Error Returns an Empty Config with Error", func(t *testing.T) {
		assertions := require.New(t)
		mClient, mux, serverURL, teardown := test.GetMockClient()
		defer teardown()

		mux.HandleFunc(
			"/repos/Spazzy757/paul/contents/",
			func(w http.ResponseWriter, r *http.Request) {
				assertions.Equal(r.Method, "GET")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `[{
		            "type": "file",
		            "name": "PAUL.yaml",
		            "download_url": "`+serverURL+baseURLPath+`/download/PAUL.yaml"
		        }]`)
			},
		)

		owner := "Spazzy757"
		repo := "paul"
		cfg, err := GetPaulConfig(
			context.Background(),
			owner, repo,
			"main",
			mClient,
		)
		assertions.Error(err)
		assertions.Equal("", cfg.PullRequests.OpenMessage)
		assertions.Equal(cfg.PullRequests.CatsEnabled, false)
		assertions.Equal(cfg.PullRequests.DogsEnabled, false)
	})
}
