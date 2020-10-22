package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/Spazzy757/paul/pkg/test"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

func getMockPayload() []byte {
	file, _ := ioutil.ReadFile("../../mocks/opened-pr.json")
	return []byte(file)
}

func TestCreateReview(t *testing.T) {
	t.Run("Test Webhook is Handled correctly", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()

		webhookPayload := getMockPayload()
		input := &github.PullRequestReviewRequest{
			Body:  github.String("test"),
			Event: github.String("COMMENT"),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/reviews",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.PullRequestReviewRequest)
				json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, r.Method, "POST")
				assert.Equal(t, input, v)
				fmt.Fprint(w, `{"id":1}`)
			},
		)

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := reviewComment(context.Background(), e.PullRequest, mClient, "test")
		assert.Equal(t, nil, err)
	})
}

func TestFirstPRCheck(t *testing.T) {
	t.Run("Test First PR - no message", func(t *testing.T) {
		firstPR := firstPRCheck("", "opened")
		assert.Equal(t, false, firstPR)
	})
	t.Run("Test First PR - should be true", func(t *testing.T) {
		firstPR := firstPRCheck("Test", "opened")
		assert.Equal(t, true, firstPR)
	})
	t.Run("Test First PR - should Be false", func(t *testing.T) {
		firstPR := firstPRCheck("Test", "closed")
		assert.Equal(t, false, firstPR)
	})
}

func TestBranchDestroyer(t *testing.T) {
	t.Run("Test Branch Destroyer Deletes Ref", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()

		mux.HandleFunc(
			"/repos/Spazzy757/paul/git/refs/heads/test",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "DELETE")
			},
		)

		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := branchDestroyer(context.Background(), e.PullRequest, mClient, "test")
		assert.Equal(t, nil, err)
	})
}

func TestBranchDestroyerCheck(t *testing.T) {
	protected := []string{"main", "master"}
	cfg := &types.BranchDestroyer{
		Enabled:           true,
		ProtectedBranches: protected,
	}
	t.Run("Test Branch Destroyer - not enabled", func(t *testing.T) {
		disabledCfg := &types.BranchDestroyer{
			Enabled:           false,
			ProtectedBranches: protected,
		}
		destroyBranch := branchDestroyerCheck(
			disabledCfg,
			"completed",
			"main",
			"feature",
		)
		assert.Equal(t, false, destroyBranch)
	})
	t.Run("Test Branch Destroyer - not completed", func(t *testing.T) {
		destroyBranch := branchDestroyerCheck(
			cfg,
			"opened",
			"main",
			"feature",
		)
		assert.Equal(t, false, destroyBranch)
	})
	t.Run("Test Branch Destroyer - default branch", func(t *testing.T) {
		destroyBranch := branchDestroyerCheck(
			cfg,
			"completed",
			"main",
			"main",
		)
		assert.Equal(t, false, destroyBranch)
	})
	t.Run("Test Branch Destroyer - protected branch", func(t *testing.T) {
		destroyBranch := branchDestroyerCheck(
			cfg,
			"completed",
			"main",
			"master",
		)
		assert.Equal(t, false, destroyBranch)
	})
	t.Run("Test Branch Destroyer - valid", func(t *testing.T) {
		destroyBranch := branchDestroyerCheck(
			cfg,
			"completed",
			"main",
			"feature",
		)
		assert.Equal(t, true, destroyBranch)
	})

}

func TestPullRequestHandler(t *testing.T) {
	mClient, mux, serverURL, teardown := test.GetMockClient()
	defer teardown()
	mux.HandleFunc(
		"/repos/Spazzy757/paul/pulls",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
			fmt.Fprint(w, `[{"number":1}]`)
		},
	)
	yamlFile, err := ioutil.ReadFile("../../PAUL.yaml")
	assert.Equal(t, nil, err)
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
	ctx := context.Background()
	t.Run("Test firstPR", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("opened-pr")
		input := &github.PullRequestReviewRequest{
			Body: github.String(
				"Greetings!\nThank you for contributing to my source code,\nIf this is your first time contributing to Paul, please make\nsure to read the [CONTRIBUTING.md](https://github.com/Spazzy757/paul/blob/main/CONTRIBUTING.md)\n",
			),
			Event: github.String("COMMENT"),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/reviews",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.PullRequestReviewRequest)
				json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, r.Method, "POST")
				assert.Equal(t, input, v)
				fmt.Fprint(w, `{"id":1}`)
			},
		)

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := PullRequestHandler(ctx, e, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test BranchDestroyer", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("merged-pr")
		mux.HandleFunc(
			"/repos/Spazzy757/paul/git/refs/heads/feature-added-webserver",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "DELETE")
			},
		)

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := PullRequestHandler(ctx, e, mClient)
		assert.Equal(t, nil, err)
	})

}

func TestGetPullRequestListForUser(t *testing.T) {
	t.Run("Test Get Pull Request returns a list and no err", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()

		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				fmt.Fprint(w, `[{"number":1}, {"number":2}]`)
			},
		)

		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		list, err := getPullRequestListForUser(context.Background(), mClient, e)
		assert.Equal(t, nil, err)
		assert.Equal(t, 2, len(list))
	})
	t.Run("Test Get Pull Request returns an err", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()

		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, ``)
			},
		)

		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		_, err := getPullRequestListForUser(context.Background(), mClient, e)
		assert.NotEqual(t, nil, err)
	})
}

func TestLimitPRCheck(t *testing.T) {
	cfg := &types.LimitPullRequests{
		MaxNumber: 1,
	}
	t.Run("Test PRs exceed limit", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()

		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				fmt.Fprint(w, `[{"number":1}, {"number":2}]`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PATCH", r.Method)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/reviews",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "POST")
				fmt.Fprint(w, `{"id":1}`)
			},
		)

		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := limitPRCheck(context.Background(), mClient, e, cfg)
		assert.Equal(t, nil, err)
	})
	t.Run("Test PRs under limit", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()

		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				fmt.Fprint(w, `[{"number":1}]`)
			},
		)

		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := limitPRCheck(context.Background(), mClient, e, cfg)
		assert.Equal(t, nil, err)
	})
	t.Run("Test PRs limit with err", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()

		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, ``)
			},
		)

		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := limitPRCheck(context.Background(), mClient, e, cfg)
		assert.NotEqual(t, nil, err)
	})
}
