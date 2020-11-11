package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/Spazzy757/paul/pkg/test"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

func TestMarkPullRequestsAsStale(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	stale := now.Add(-300 * time.Hour * 24)
	repo := github.Repository{
		Owner: &github.User{
			Login: github.String("Spazzy757"),
		},
		Name:          github.String("paul"),
		DefaultBranch: github.String("main"),
	}
	stalePullRequest := github.PullRequest{
		ID:        github.Int64(1),
		Number:    github.Int(1),
		UpdatedAt: &stale,
		Base: &github.PullRequestBranch{
			Repo: &github.Repository{
				Name: github.String("paul"),
				Owner: &github.User{
					Login: github.String("Spazzy757"),
				},
			},
		},
	}
	notStalePullRequest := github.PullRequest{
		ID:        github.Int64(2),
		UpdatedAt: &now,
	}
	t.Run("Test Unknown Webhook Payload is Handled correctly", func(t *testing.T) {
		mClient, mux, serverURL, teardown := test.GetMockClient()
		defer teardown()
		yamlFile, err := ioutil.ReadFile("../../PAUL.yaml")
		assert.Equal(t, nil, err)
		mux.HandleFunc(
			"/installation/repositories",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				jsonBytes, _ := json.Marshal(repo)
				response := fmt.Sprintf(`{"repositories": [%v]}`, string(jsonBytes))
				fmt.Fprint(w, response)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				jsonBytesStale, _ := json.Marshal(stalePullRequest)
				jsonBytesNotStale, _ := json.Marshal(notStalePullRequest)
				response := fmt.Sprintf(
					`[%v, %v]}`,
					string(jsonBytesStale),
					string(jsonBytesNotStale),
				)
				fmt.Fprint(w, response)
			},
		)
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
		input := []string{"stale"}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/issues/1/labels",
			func(w http.ResponseWriter, r *http.Request) {
				var v []string
				json.NewDecoder(r.Body).Decode(&v)
				assert.Equal(t, v, input)
				fmt.Fprint(w, `[{"url":"u"}]`)
			},
		)
		mux.HandleFunc("/download/PAUL.yaml", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, string(yamlFile))
		})
		MarkPullRequestsAsStale(ctx, mClient)
	})
}

func TestHandleError(t *testing.T) {

	t.Run("Test Error return true", func(t *testing.T) {
		check := handleError(fmt.Errorf("test"))
		assert.Equal(t, true, check)
	})
	t.Run("Test No Error return false", func(t *testing.T) {
		check := handleError(nil)
		assert.Equal(t, false, check)
	})
}
