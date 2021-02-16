package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/Spazzy757/paul/pkg/animals"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/Spazzy757/paul/pkg/test"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

func getIssueCommentMockPayload(payloadType string) []byte {
	fileLocation := fmt.Sprintf("../../mocks/%v.json", payloadType)
	file, _ := ioutil.ReadFile(fileLocation)
	return []byte(file)
}

func TestCreateComment(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		webhookPayload := getIssueCommentMockPayload("opened-pr")
		input := &github.IssueComment{Body: github.String("test")}
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		mux.HandleFunc(
			"/repos/Spazzy757/paul/issues/0/comments",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.IssueComment)
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, v, input)
				fmt.Fprint(w, `{"id":1}`)
			},
		)

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := createIssueComment(context.Background(), e, mClient, "test")
		assert.Equal(t, nil, err)
	})
}

func TestHandleCats(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()

		input := &github.IssueComment{
			Body: github.String(
				"My Most Trusted Minion\n\n ![my favorite minion](https://cdn2.thecatapi.com/images/40g.jpg)",
			),
		}

		mux.HandleFunc(
			"/repos/Spazzy757/paul/issues/9/comments",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.IssueComment)
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, v, input)
				fmt.Fprint(w, `{"id":1}`)
			},
		)

		webhookPayload := getIssueCommentMockPayload("cat-command")
		catAPIResponse := `[
            {
                "breeds":[],
                "id":"40g",
                "url":"https://cdn2.thecatapi.com/images/40g.jpg",
                "width":640,
                "height":426
            }
        ]`
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(catAPIResponse))
		})
		httpClient, httpteardown := helpers.MockHTTPClient(h)
		defer httpteardown()

		catClient := animals.NewCatClient()
		catClient.HttpClient = httpClient
		catClient.Url = "https://example.com"

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := catsHandler(context.Background(), e, mClient, catClient)
		assert.Equal(t, nil, err)
	})
}

func TestHandleDogs(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()

		input := &github.IssueComment{
			Body: github.String(
				"Despite how it looks it is well trained\n\n ![loyal soldier](https://cdn2.thedogapi.com/images/40g.jpg)",
			),
		}

		mux.HandleFunc(
			"/repos/Spazzy757/paul/issues/9/comments",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.IssueComment)
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, v, input)
				fmt.Fprint(w, `{"id":1}`)
			},
		)

		webhookPayload := getIssueCommentMockPayload("dog-command")
		dogAPIResponse := `[
            {
                "breeds":[],
                "id":"40g",
                "url":"https://cdn2.thedogapi.com/images/40g.jpg",
                "width":640,
                "height":426
            }
        ]`
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(dogAPIResponse))
		})
		httpClient, httpteardown := helpers.MockHTTPClient(h)
		defer httpteardown()

		dogClient := animals.NewDogClient()
		dogClient.HttpClient = httpClient
		dogClient.Url = "https://example.com"

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := dogsHandler(context.Background(), e, mClient, dogClient)
		assert.Equal(t, nil, err)
	})
}

func TestHandleLabels(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := &types.PaulConfig{
			Maintainers: []string{
				"Spazzy757",
			},
			Labels: true,
		}

		input := []string{"test"}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/issues/9/labels",
			func(w http.ResponseWriter, r *http.Request) {
				var v []string
				_ = json.NewDecoder(r.Body).Decode(&v)
				assert.Equal(t, v, input)
				fmt.Fprint(w, `[{"url":"u"}]`)
			},
		)
		// Just needed to get the right event type
		webhookPayload := getIssueCommentMockPayload("dog-command")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := labelHandler(context.Background(), cfg, e, mClient, input)
		assert.Equal(t, nil, err)
	})
}

func TestHandleRemoveLabels(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := &types.PaulConfig{
			Maintainers: []string{
				"Spazzy757",
			},
		}

		mux.HandleFunc(
			"/repos/Spazzy757/paul/issues/9/labels/test",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "DELETE")
			},
		)

		// Just needed to get the right event type
		webhookPayload := getIssueCommentMockPayload("removelabel-command")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := removeLabelHandler(context.Background(), cfg, e, mClient, []string{"test"})
		assert.Equal(t, nil, err)
	})
}

func TestCheckStringInList(t *testing.T) {
	maintainers := []string{"yes", "no", "maybe"}
	t.Run("Test Maintainer returns true", func(t *testing.T) {
		assert.Equal(t, true, checkStringInList(maintainers, "yes"))
	})
	t.Run("Test Non Maintainer returns false", func(t *testing.T) {
		assert.Equal(t, false, checkStringInList(maintainers, "I don't know"))
	})
}

func TestGetCommand(t *testing.T) {
	t.Run("Test If Command, command returned", func(t *testing.T) {
		comment := "/cat"
		expected := "cat"
		cmd, _ := getCommand(comment)
		assert.Equal(t, expected, cmd)
	})
	t.Run("Test If not Command, nothing is returned", func(t *testing.T) {
		comment := "cat"
		expected := ""
		cmd, _ := getCommand(comment)
		assert.Equal(t, expected, cmd)
	})
	t.Run("Test If Command has args, command and args returned", func(t *testing.T) {
		comment := "/label invalid"
		expectedCommand := "label"
		expectedArgs := []string{"invalid"}
		cmd, args := getCommand(comment)
		assert.Equal(t, expectedCommand, cmd)
		assert.Equal(t, expectedArgs, args)
	})

}

func TestApproveHandler(t *testing.T) {
	t.Run("Test Approve Command Is handled", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		cfg := &types.PaulConfig{
			Maintainers: []string{
				"Spazzy757",
			},
		}

		webhookPayload := getIssueCommentMockPayload("approve-command")
		input := &github.PullRequestReviewRequest{
			Event: github.String("APPROVE"),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/9/reviews",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.PullRequestReviewRequest)
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, r.Method, "POST")
				assert.Equal(t, input, v)
				fmt.Fprint(w, `{"id":1}`)
			},
		)

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := approveHandler(context.Background(), cfg, e, mClient)
		assert.Equal(t, nil, err)
	})
}

func TestMergeHandler(t *testing.T) {
	mClient, mux, _, teardown := test.GetMockClient()
	defer teardown()
	cfg := &types.PaulConfig{
		Maintainers: []string{
			"Spazzy757",
		},
	}
	t.Run("Test Merge Pull Request", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("merge-command")
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/7/merge",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "PUT")
				fmt.Fprint(w, `
			{
			  "sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			  "merged": true,
			  "message": "Pull Request successfully merged"
			}`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/7",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				mockPr := `{
                    "number":7,
                    "mergeable": true,
                    "base": {
                        "repo": {"name":"paul", "owner":{"login":"Spazzy757"}}
                    }
                }`
				fmt.Fprint(w, mockPr)
			},
		)
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		e.Issue.Number = github.Int(7)
		err := mergeHandler(context.Background(), cfg, e, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Merge Pull Request Cant Merge", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("merge-command")
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/9",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				fmt.Fprint(w, `{"number":9}`)
			})
		input := &github.IssueComment{
			Body: github.String("This Pull Request Can not be merge currently"),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/issues/9/comments",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.IssueComment)
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, v, input)
				fmt.Fprint(w, `{"id":1}`)
			},
		)
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := mergeHandler(context.Background(), cfg, e, mClient)
		assert.Equal(t, nil, err)
	})
}

func TestIssueCommentHandler(t *testing.T) {
	mClient, mux, serverURL, teardown := test.GetMockClient()
	defer teardown()
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
	ctx := context.Background()
	t.Run("Test created Command", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("opened-pr")
		input := &github.IssueComment{Body: github.String("test")}
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		mux.HandleFunc(
			"/repos/Spazzy757/paul/issues/0/comments",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.IssueComment)
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, v, input)
				fmt.Fprint(w, `{"id":1}`)
			},
		)

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := IssueCommentHandler(ctx, e, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test created Command", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("approve-command")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		input := &github.PullRequestReviewRequest{
			Event: github.String("APPROVE"),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/9/reviews",
			func(w http.ResponseWriter, r *http.Request) {
				v := new(github.PullRequestReviewRequest)
				_ = json.NewDecoder(r.Body).Decode(v)
				assert.Equal(t, r.Method, "POST")
				assert.Equal(t, input, v)
				fmt.Fprint(w, `{"id":1}`)
			},
		)
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := IssueCommentHandler(ctx, e, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test label Command", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("label-command")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		input := []string{"enhancement"}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/issues/9/labels",
			func(w http.ResponseWriter, r *http.Request) {
				var v []string
				_ = json.NewDecoder(r.Body).Decode(&v)
				assert.Equal(t, v, input)
				fmt.Fprint(w, `[{"url":"u"}]`)
			},
		)
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := IssueCommentHandler(ctx, e, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test remove label Command", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("removelabel-command")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		mux.HandleFunc(
			"/repos/Spazzy757/paul/issues/9/labels/enhancement",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "DELETE")
			},
		)
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := IssueCommentHandler(ctx, e, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Merge Pull Request", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("merge-command")
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/7/merge",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "PUT")
				fmt.Fprint(w, `
			{
			  "sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			  "merged": true,
			  "message": "Pull Request successfully merged"
			}`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/7",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				mockPr := `{
                    "number":7,
                    "mergeable": true,
                    "base": {
                        "repo": {"name":"paul", "owner":{"login":"Spazzy757"}}
                    }
                }`
				fmt.Fprint(w, mockPr)
			},
		)
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		e.Issue.Number = github.Int(7)
		err := IssueCommentHandler(ctx, e, mClient)
		assert.Equal(t, nil, err)
	})

	t.Run("Test unknown Command", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("unknown-command")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := IssueCommentHandler(ctx, e, mClient)
		assert.Equal(t, nil, err)
	})

}
