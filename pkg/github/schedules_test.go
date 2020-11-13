package github

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/Spazzy757/paul/pkg/test"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

func TestPullRequestsScheduledJobs(t *testing.T) {
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
		PullRequestsScheduledJobs(ctx, mClient)
	})
}

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
	cfg := types.PaulConfig{
		PullRequests: types.PullRequests{
			StaleTime: 15,
		},
	}
	t.Run("Test Marking Stale", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
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
		markPullRequestsStale(
			ctx,
			mClient,
			[]*ScehduledJobInformation{
				{
					Cfg:          cfg,
					Repo:         &repo,
					PullRequests: []*github.PullRequest{&stalePullRequest, &notStalePullRequest},
				},
			})
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

func TestCheckTimeStamp(t *testing.T) {
	now := time.Now()
	stale := now.Add(-300 * time.Hour * 24)
	stalePullRequest := github.PullRequest{
		ID:        github.Int64(1),
		UpdatedAt: &stale,
	}
	notStalePullRequest := github.PullRequest{
		ID:        github.Int64(2),
		UpdatedAt: &now,
	}
	pullRequestList := []*github.PullRequest{
		&stalePullRequest,
		&notStalePullRequest,
	}
	cfg := types.PaulConfig{
		PullRequests: types.PullRequests{
			StaleTime: 15,
		},
	}
	t.Run("Test Returns only Stale Pull Requests", func(t *testing.T) {
		stalePullRequests := checkTimeStamps(cfg, pullRequestList)
		assert.Equal(t, []*github.PullRequest{
			&stalePullRequest,
		}, stalePullRequests)
	})
}

func TestGetScehduledJobInformationList(t *testing.T) {
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
	cfg := types.PaulConfig{
		PullRequests: types.PullRequests{
			StaleTime: 15,
		},
	}
	cfgBytes, err := yaml.Marshal(cfg)
	assert.Equal(t, nil, err)
	t.Run("Test Unknown Webhook Payload is Handled correctly", func(t *testing.T) {
		mClient, mux, serverURL, teardown := test.GetMockClient()
		defer teardown()
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
		mux.HandleFunc("/download/PAUL.yaml", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, string(cfgBytes))
		})
		infoList, err := getScehduledJobInformationList(ctx, mClient)
		assert.Equal(t, nil, err)
		assert.Equal(t, 1, len(infoList))
		assert.Equal(t, 2, len(infoList[0].PullRequests))
	})
}
