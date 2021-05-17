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
	"github.com/google/go-github/v35/github"
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
				_ = json.NewDecoder(r.Body).Decode(v)
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
	mClient, mux, _, teardown := test.GetMockClient()
	webhookPayload := test.GetMockPayload("opened-pr")
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)

	defer teardown()
	t.Run("Test First PR - no message", func(t *testing.T) {
		cfg := types.PaulConfig{
			PullRequests: types.PullRequests{
				OpenMessage: "",
			},
		}
		err := firstPRCheck(context.Background(), cfg, mClient, e)
		assert.Equal(t, nil, err)
	})
	t.Run("Test First PR - should be true", func(t *testing.T) {
		cfg := types.PaulConfig{
			PullRequests: types.PullRequests{
				OpenMessage: "test",
			},
		}
		input := &github.PullRequestReviewRequest{
			Body:  github.String("test"),
			Event: github.String("COMMENT"),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/reviews",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.PullRequestReviewRequest)
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, r.Method, "POST")
				assert.Equal(t, input, v)
				fmt.Fprint(w, `{"id":1}`)
			},
		)
		err := firstPRCheck(context.Background(), cfg, mClient, e)
		assert.Equal(t, nil, err)
	})
	t.Run("Test First PR - should Be false", func(t *testing.T) {
		mergedPayload := test.GetMockPayload("merged-pr")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(mergedPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")
		event, _ := github.ParseWebHook(github.WebHookType(req), mergedPayload)
		e := event.(*github.PullRequestEvent)
		cfg := types.PaulConfig{
			PullRequests: types.PullRequests{
				OpenMessage: "test",
			},
		}
		err := firstPRCheck(context.Background(), cfg, mClient, e)
		assert.Equal(t, nil, err)
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
	protected := []string{"master"}
	mClient, mux, _, teardown := test.GetMockClient()
	defer teardown()
	cfg := types.PaulConfig{
		BranchDestroyer: types.BranchDestroyer{
			Enabled:           true,
			ProtectedBranches: protected,
		},
	}
	disabledCfg := types.PaulConfig{
		BranchDestroyer: types.BranchDestroyer{
			Enabled: false,
		},
	}
	webhookPayload := test.GetMockPayload("merged-pr")

	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")

	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)
	ctx := context.Background()
	t.Run("Test Branch Destroyer - not enabled", func(t *testing.T) {
		err := branchDestroyerCheck(ctx, disabledCfg, mClient, e)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Branch Destroyer - not completed", func(t *testing.T) {
		e.Action = github.String("opened")
		err := branchDestroyerCheck(ctx, cfg, mClient, e)
		e.Action = github.String("completed")
		assert.Equal(t, nil, err)
	})
	t.Run("Test Branch Destroyer - default branch", func(t *testing.T) {
		e.PullRequest.Head.Ref = github.String("main")
		err := branchDestroyerCheck(ctx, cfg, mClient, e)
		e.PullRequest.Head.Ref = github.String("feature-added-webserver")
		assert.Equal(t, nil, err)
	})
	t.Run("Test Branch Destroyer - protected branch", func(t *testing.T) {
		e.PullRequest.Head.Ref = github.String("master")
		err := branchDestroyerCheck(ctx, cfg, mClient, e)
		e.PullRequest.Head.Ref = github.String("feature-added-webserver")
		assert.Equal(t, nil, err)
	})
	t.Run("Test Branch Destroyer - not merged", func(t *testing.T) {
		e.PullRequest.Merged = github.Bool(false)
		err := branchDestroyerCheck(ctx, cfg, mClient, e)
		e.PullRequest.Merged = github.Bool(true)
		assert.Equal(t, nil, err)
	})

	t.Run("Test Branch Destroyer - valid", func(t *testing.T) {
		mux.HandleFunc(
			"/repos/Spazzy757/paul/git/refs/heads/feature-added-webserver",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "DELETE")
			},
		)
		err := branchDestroyerCheck(ctx, cfg, mClient, e)
		assert.Equal(t, nil, err)
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
	yamlFile, err := ioutil.ReadFile("../../.github/PAUL.yaml")
	assert.Equal(t, nil, err)
	mux.HandleFunc(
		"/repos/Spazzy757/paul/contents/",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "GET")
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
		fmt.Fprint(w, string(yamlFile))
	})
	mux.HandleFunc(
		"/repos/Spazzy757/paul/commits/83e12d84247dcc85e05ea18d558be01ce6b0c128/check-runs",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"total_count":1,
                                "check_runs": [{
                                    "id": 1,
                                    "head_sha": "deadbeef",
                                    "status": "completed",
                                    "conclusion": "neutral",
                                    "started_at": "2018-05-04T01:14:52Z",
                                    "completed_at": "2018-05-04T01:14:52Z"}]}`)
		},
	)

	mux.HandleFunc(
		"/repos/Spazzy757/paul/check-runs/1",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{
			            "id": 1,
                        "name":"DeveloperCertificateOfOrigin",
						"status": "completed",
						"conclusion": "failed",
						"started_at": "2018-05-04T01:14:52Z",
						"completed_at": "2018-05-04T01:14:52Z",
                        "output":{
                            "title": "Mighty test report",
							"summary":"There are 0 failures, 2 warnings and 1 notice",
							"text":"You may have misspelled some words."
                        }
                }`)
		},
	)

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
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, r.Method, "POST")
				assert.Equal(t, input, v)
				fmt.Fprint(w, `{"id":1}`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/commits",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[
					  {
						"sha": "2",
						"parents": [
						  {
							"sha": "1"
						  }
						],
                        "commit": {
                            "message": "Signed-off-by: test"
                        }
					  }
					]`)
			},
		)

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := PullRequestHandler(ctx, e, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Empty Description Prompts a review", func(t *testing.T) {

		input := &github.PullRequestReviewRequest{
			Body:  github.String("There seems to be no description on this Pull Request, please provide a description so I can understand this Pull Requests Context"),
			Event: github.String("COMMENT"),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/2/reviews",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.PullRequestReviewRequest)
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, r.Method, "POST")
				assert.Equal(t, input, v)
				fmt.Fprint(w, `{"id":1}`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/2/commits",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[
					  {
						"sha": "2",
						"parents": [
						  {
							"sha": "1"
						  }
						],
                        "commit": {
                            "message": "Signed-off-by: test"
                        }
					  }
					]`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/2",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PATCH", r.Method)
			},
		)

		webhookPayload := test.GetMockPayload("empty-description-pr")

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)

		e.PullRequest.Number = github.Int(2)
		e.Number = github.Int(2)
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
	cfg := types.PaulConfig{
		PullRequests: types.PullRequests{
			LimitPullRequests: types.LimitPullRequests{
				MaxNumber: 1,
			},
		},
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
		err := limitPRCheck(context.Background(), cfg, mClient, e)
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
		err := limitPRCheck(context.Background(), cfg, mClient, e)
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
		err := limitPRCheck(context.Background(), cfg, mClient, e)
		assert.NotEqual(t, nil, err)
	})
}

func TestMergePullRequest(t *testing.T) {
	t.Run("Test merge pull request", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()

		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/merge",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "PUT")
				fmt.Fprint(w, `
			{
			  "sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			  "merged": true,
			  "message": "Pull Request successfully merged"
			}`)
			})

		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := mergePullRequest(context.Background(), mClient, e.PullRequest)
		assert.Equal(t, nil, err)
	})
	t.Run("Test merge pull request fails", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()

		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/merge",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "PUT")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, ``)
			})

		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := mergePullRequest(context.Background(), mClient, e.PullRequest)
		assert.NotEqual(t, nil, err)
	})
}

func TestEmptyDescriptionCheck(t *testing.T) {
	t.Run("Test Empty Description Prompts a review", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := types.PaulConfig{
			EmptyDescriptionCheck: types.EmptyDescriptionCheck{
				Enabled:  true,
				Enforced: false,
			},
		}

		input := &github.PullRequestReviewRequest{
			Body:  github.String(emptyDescriptionMessage),
			Event: github.String("COMMENT"),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/reviews",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.PullRequestReviewRequest)
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, r.Method, "POST")
				assert.Equal(t, input, v)
				fmt.Fprint(w, `{"id":1}`)
			},
		)

		webhookPayload := test.GetMockPayload("empty-description-pr")

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := emptyDescriptionCheck(context.Background(), cfg, mClient, e)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Empty Description Prompts a review Enforced", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := types.PaulConfig{
			EmptyDescriptionCheck: types.EmptyDescriptionCheck{
				Enabled:  true,
				Enforced: true,
			},
		}

		input := &github.PullRequestReviewRequest{
			Body:  github.String(emptyDescriptionMessage),
			Event: github.String("COMMENT"),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/reviews",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.PullRequestReviewRequest)
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, r.Method, "POST")
				assert.Equal(t, input, v)
				fmt.Fprint(w, `{"id":1}`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PATCH", r.Method)
			},
		)

		webhookPayload := test.GetMockPayload("empty-description-pr")

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := emptyDescriptionCheck(context.Background(), cfg, mClient, e)
		assert.Equal(t, nil, err)
	})

	t.Run("Test Non Empty Description Does Nothing", func(t *testing.T) {
		mClient, _, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := types.PaulConfig{
			EmptyDescriptionCheck: types.EmptyDescriptionCheck{
				Enabled:  true,
				Enforced: false,
			},
		}

		webhookPayload := test.GetMockPayload("opened-pr")

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := emptyDescriptionCheck(context.Background(), cfg, mClient, e)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Error Gets Returned", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := types.PaulConfig{
			EmptyDescriptionCheck: types.EmptyDescriptionCheck{
				Enabled:  true,
				Enforced: false,
			},
		}

		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/reviews",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, ``)
			},
		)
		webhookPayload := test.GetMockPayload("empty-description-pr")

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := emptyDescriptionCheck(context.Background(), cfg, mClient, e)
		assert.NotEqual(t, nil, err)
	})
}

func TestDCOCheck(t *testing.T) {
	ctx := context.Background()
	webhookPayload := test.GetMockPayload("opened-pr")

	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")

	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)
	t.Run("Test DCO Check Disabled", func(t *testing.T) {
		mClient, _, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := types.PaulConfig{
			PullRequests: types.PullRequests{
				DCOCheck: false,
			},
		}
		err := dcoCheck(ctx, cfg, mClient, e)
		assert.Equal(t, nil, err)
	})
	t.Run("Test DCO Check Unsigned", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := types.PaulConfig{
			PullRequests: types.PullRequests{
				DCOCheck: true,
			},
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":1,
                                "check_runs": [{
                                    "id": 1,
                                    "head_sha": "deadbeef",
                                    "status": "completed",
                                    "conclusion": "neutral",
                                    "started_at": "2018-05-04T01:14:52Z",
                                    "completed_at": "2018-05-04T01:14:52Z"}]}`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/commits",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[
					  {
						"sha": "2",
						"parents": [
						  {
							"sha": "1"
						  }
						],
                        "commit": {
                            "message": "Test"
                        }
					  }
					]`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/1",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{
			            "id": 1,
                        "name":"DeveloperCertificateOfOrigin",
						"status": "completed",
						"conclusion": "failed",
						"started_at": "2018-05-04T01:14:52Z",
						"completed_at": "2018-05-04T01:14:52Z",
                        "output":{
                            "title": "Mighty test report",
							"summary":"There are 0 failures, 2 warnings and 1 notice",
							"text":"You may have misspelled some words."
                        }
                }`)
			},
		)
		err := dcoCheck(ctx, cfg, mClient, e)
		assert.Equal(t, nil, err)
	})
	t.Run("Test DCO Check Signed", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := types.PaulConfig{
			PullRequests: types.PullRequests{
				DCOCheck: true,
			},
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":1,
                                "check_runs": [{
                                    "id": 1,
                                    "head_sha": "deadbeef",
                                    "status": "completed",
                                    "conclusion": "neutral",
                                    "started_at": "2018-05-04T01:14:52Z",
                                    "completed_at": "2018-05-04T01:14:52Z"}]}`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/commits",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[
					  {
						"sha": "2",
						"parents": [
						  {
							"sha": "1"
						  }
						],
                        "commit": {
                            "message": "Signed-off-by: test"
                        }
					  }
					]`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/1",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{
			            "id": 1,
                        "name":"DeveloperCertificateOfOrigin",
						"status": "completed",
						"conclusion": "failed",
						"started_at": "2018-05-04T01:14:52Z",
						"completed_at": "2018-05-04T01:14:52Z",
                        "output":{
                            "title": "Mighty test report",
							"summary":"There are 0 failures, 2 warnings and 1 notice",
							"text":"You may have misspelled some words."
                        }
                }`)
			},
		)
		err := dcoCheck(ctx, cfg, mClient, e)
		assert.Equal(t, nil, err)
	})
}

func TestVerifyCheck(t *testing.T) {
	ctx := context.Background()
	webhookPayload := test.GetMockPayload("opened-pr")

	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")

	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)
	t.Run("Test Verify Check Disabled", func(t *testing.T) {
		mClient, _, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := types.PaulConfig{
			PullRequests: types.PullRequests{
				VerifiedCommitCheck: false,
			},
		}
		err := verifiedCommitCheck(ctx, cfg, mClient, e)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Verify Check Commits Unverified", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := types.PaulConfig{
			PullRequests: types.PullRequests{
				VerifiedCommitCheck: true,
			},
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":1,
                                "check_runs": [{
                                    "id": 1,
                                    "head_sha": "deadbeef",
                                    "status": "completed",
                                    "conclusion": "neutral",
                                    "started_at": "2018-05-04T01:14:52Z",
                                    "completed_at": "2018-05-04T01:14:52Z"}]}`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/commits",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[
					  {
						"sha": "2",
						"parents": [
						  {
							"sha": "1"
						  }
						],
                        "commit": {
                            "message": "Test"
                        }
					  }
					]`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/1",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{
			            "id": 1,
                        "name":"Verified Commits",
						"status": "completed",
						"conclusion": "failed",
						"started_at": "2018-05-04T01:14:52Z",
						"completed_at": "2018-05-04T01:14:52Z",
                        "output":{
                            "title": "Mighty test report",
							"summary":"There are 0 failures, 2 warnings and 1 notice",
							"text":"You may have misspelled some words."
                        }
                }`)
			},
		)
		err := verifiedCommitCheck(ctx, cfg, mClient, e)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Verify Check Commits Verified", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := types.PaulConfig{
			PullRequests: types.PullRequests{
				VerifiedCommitCheck: true,
			},
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":1,
                                "check_runs": [{
                                    "id": 1,
                                    "head_sha": "deadbeef",
                                    "status": "completed",
                                    "conclusion": "neutral",
                                    "started_at": "2018-05-04T01:14:52Z",
                                    "completed_at": "2018-05-04T01:14:52Z"}]}`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/commits",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[
					  {
						"sha": "2",
						"parents": [
						  {
							"sha": "1"
						  }
						],
                        "commit": {
                            "message": "Signed-off-by: test"
                        }
					  }
					]`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/1",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{
			            "id": 1,
                        "name":"Verified Commits",
						"status": "completed",
						"conclusion": "failed",
						"started_at": "2018-05-04T01:14:52Z",
						"completed_at": "2018-05-04T01:14:52Z",
                        "output":{
                            "title": "Mighty test report",
							"summary":"There are 0 failures, 2 warnings and 1 notice",
							"text":"You may have misspelled some words."
                        }
                }`)
			},
		)
		err := dcoCheck(ctx, cfg, mClient, e)
		assert.Equal(t, nil, err)
	})
}
