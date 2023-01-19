package github

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Spazzy757/paul/pkg/test"
	"github.com/google/go-github/v49/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchPullRequestCommits(t *testing.T) {
	mClient, mux, _, teardown := test.GetMockClient()
	webhookPayload := test.GetMockPayload("opened-pr")
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)

	defer teardown()
	t.Run("Test Returns Commits", func(t *testing.T) {
		want := []*github.RepositoryCommit{
			{
				SHA: github.String("3"),
				Parents: []*github.Commit{
					{
						SHA: github.String("2"),
					},
				},
			},
			{
				SHA: github.String("2"),
				Parents: []*github.Commit{
					{
						SHA: github.String("1"),
					},
				},
			},
		}

		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls/1/commits",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `
					[
					  {
						"sha": "3",
						"parents": [
						  {
							"sha": "2"
						  }
						]
					  },
					  {
						"sha": "2",
						"parents": [
						  {
							"sha": "1"
						  }
						]
					  }
					]`)
			},
		)
		commits, err := getPullRequestCommits(context.Background(), e, mClient)
		assert.Equal(t, nil, err)
		assert.Equal(t, want, commits)
	})
	t.Run("Test Returns err", func(t *testing.T) {
		_, err := getPullRequestCommits(context.Background(), e, mClient)
		assert.Equal(t, nil, err)
	})
}

func TestCreateDCOCheck(t *testing.T) {
	webhookPayload := test.GetMockPayload("opened-pr")
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)

	t.Run("Test Creating a DCO Check", func(t *testing.T) {
		check := createDCOCheck(e)
		assert.Equal(t, e.PullRequest.Head.GetSHA(), check.HeadSHA)
	})
}

func TestCreateVerifyCheck(t *testing.T) {
	webhookPayload := test.GetMockPayload("opened-pr")
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)

	t.Run("Test Creating a DCO Check", func(t *testing.T) {
		assertions := require.New(t)
		check := createVerifyCheck(e)
		assertions.Equal(e.PullRequest.Head.GetSHA(), check.HeadSHA)
		assertions.Equal(started, check.GetStatus())
		assertions.Equal(verified, check.Name)
	})
}

func TestDetermineExsitingDCOCheck(t *testing.T) {
	mClient, mux, _, teardown := test.GetMockClient()
	webhookPayload := test.GetMockPayload("opened-pr")
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)
	startedAt, _ := time.Parse(time.RFC3339, "2018-05-04T01:14:52Z")

	defer teardown()
	t.Run("Test Get DCO Checks", func(t *testing.T) {
		want := &github.ListCheckRunsResults{
			Total: github.Int(1),
			CheckRuns: []*github.CheckRun{{
				ID:          github.Int64(1),
				Status:      github.String("completed"),
				CompletedAt: &github.Timestamp{Time: startedAt},
				StartedAt:   &github.Timestamp{Time: startedAt},
				Conclusion:  github.String("neutral"),
				HeadSHA:     github.String("deadbeef"),
			}},
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
		checks, err := getExistingDCOCheck(context.Background(), e, mClient, dco)
		assert.Equal(t, nil, err)
		assert.Equal(t, want, checks)
	})
	t.Run("Test Get DCO Checks Fails", func(t *testing.T) {
		e.PullRequest.Head.SHA = github.String("master")
		_, err := getExistingDCOCheck(context.Background(), e, mClient, dco)
		assert.NotEqual(t, nil, err)
	})

}

func TestUpdateSuccessfulDCOCheck(t *testing.T) {
	startedAt, _ := time.Parse(time.RFC3339, "2018-05-04T01:14:52Z")
	t.Run("Test DCO Check is unsuccessful", func(t *testing.T) {
		input := &github.CheckRun{
			ID:          github.Int64(1),
			Name:        github.String(dco),
			Status:      github.String("completed"),
			StartedAt:   &github.Timestamp{Time: startedAt},
			CompletedAt: &github.Timestamp{Time: startedAt},
			Conclusion:  github.String("neutral"),
			HeadSHA:     github.String("deadbeef"),
		}
		check := updateSuccessfulDCOCheck(input)
		assert.Equal(t, success, check.GetConclusion())
	})
}

func TestUpdateUnsuccessfulDCOCheck(t *testing.T) {
	startedAt, _ := time.Parse(time.RFC3339, "2018-05-04T01:14:52Z")
	t.Run("Test DCO Check is unsuccessful", func(t *testing.T) {
		input := &github.CheckRun{
			ID:          github.Int64(1),
			Name:        github.String(dco),
			Status:      github.String("completed"),
			StartedAt:   &github.Timestamp{Time: startedAt},
			CompletedAt: &github.Timestamp{Time: startedAt},
			Conclusion:  github.String("neutral"),
			HeadSHA:     github.String("deadbeef"),
		}
		check := updateUnsuccessfulDCOCheck(input)
		assert.Equal(t, failed, check.GetConclusion())
	})
}

func TestUpdateUnsuccessfulVerifyCheck(t *testing.T) {
	startedAt, _ := time.Parse(time.RFC3339, "2018-05-04T01:14:52Z")
	t.Run("Test Verify Check is unsuccessful", func(t *testing.T) {
		assertions := require.New(t)
		input := &github.CheckRun{
			ID:          github.Int64(1),
			Name:        github.String(verified),
			Status:      github.String("completed"),
			StartedAt:   &github.Timestamp{Time: startedAt},
			CompletedAt: &github.Timestamp{Time: startedAt},
			Conclusion:  github.String("neutral"),
			HeadSHA:     github.String("deadbeef"),
		}
		check := updateUnsuccessfulVerifyCheck(input)
		assertions.Equal(failed, check.GetConclusion())
		assertions.Equal(verified, check.Name)
	})
}

func TestUpdateSuccessfulVerifyCheck(t *testing.T) {
	startedAt, _ := time.Parse(time.RFC3339, "2018-05-04T01:14:52Z")
	t.Run("Test Verify Check is unsuccessful", func(t *testing.T) {
		assertions := require.New(t)
		input := &github.CheckRun{
			ID:          github.Int64(1),
			Name:        github.String(verified),
			Status:      github.String("completed"),
			StartedAt:   &github.Timestamp{Time: startedAt},
			CompletedAt: &github.Timestamp{Time: startedAt},
			Conclusion:  github.String("neutral"),
			HeadSHA:     github.String("deadbeef"),
		}
		check := updateSuccessfulVerifyCheck(input)
		assertions.Equal(success, check.GetConclusion())
		assertions.Equal(verified, check.Name)
	})
}

func TestCreateSuccessfulDCOCheck(t *testing.T) {
	webhookPayload := test.GetMockPayload("opened-pr")
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)

	t.Run("Test get existing check works", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
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
		_, err := createSuccessfulDCOCheck(context.Background(), e, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test get existing checks returns 2 fails", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":2,
                                "check_runs": [{
                                    "id": 1,
                                    "head_sha": "deadbeef",
                                    "status": "completed",
                                    "conclusion": "neutral",
                                    "started_at": "2018-05-04T01:14:52Z",
                                    "completed_at": "2018-05-04T01:14:52Z"}]}`)
			},
		)
		_, err := createSuccessfulDCOCheck(context.Background(), e, mClient)
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test get existing checks returns invalid payload", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":2,*`)
			},
		)
		_, err := createSuccessfulDCOCheck(context.Background(), e, mClient)
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test get existing checks creates new check", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":0}`)
			},
		)
		mux.HandleFunc("/repos/Spazzy757/paul/check-runs", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{
				  "id": 1,
				  "name": "DeveloperCertificateOfOrigin",
				  "head_sha":"deadbeef",
				  "status": "in_progress",
				  "conclusion": null,
				  "started_at": "2018-05-04T01:14:52Z",
				  "completed_at": null,
                  "output":{"title": "Mighty test report", "summary":"", "text":""}}`)
		})
		_, err := createSuccessfulDCOCheck(context.Background(), e, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test get existing checks creates new check error", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":0}`)
			},
		)
		mux.HandleFunc("/repos/Spazzy757/paul/check-runs", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{`)
		})
		_, err := createSuccessfulDCOCheck(context.Background(), e, mClient)
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test get existing checks creates new check returns bad code", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":0}`)
			},
		)
		mux.HandleFunc("/repos/Spazzy757/paul/check-runs", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{
				  "id": 1,
				  "name": "DeveloperCertificateOfOrigin",
				  "head_sha":"deadbeef",
				  "status": "in_progress",
				  "conclusion": null,
				  "started_at": "2018-05-04T01:14:52Z",
				  "completed_at": null,
                  "output":{"title": "Mighty test report", "summary":"", "text":""}}`)
		})
		_, err := createSuccessfulDCOCheck(context.Background(), e, mClient)
		assert.NotEqual(t, nil, err)
	})
}

func TestCommitChecks(t *testing.T) {
	t.Run("Test Has Unsigned Commits", func(t *testing.T) {
		commits := []*github.RepositoryCommit{
			&github.RepositoryCommit{
				Commit: &github.Commit{
					Message: github.String("Signed-off-by: User users@users.noreply.github.com"),
				},
			},
		}
		check := hasUnsigned(commits)
		assert.Equal(t, false, check)
	})
	t.Run("Test has Unverified Commits", func(t *testing.T) {
		commits := []*github.RepositoryCommit{
			&github.RepositoryCommit{
				Commit: &github.Commit{
					Verification: &github.SignatureVerification{
						Verified: github.Bool(false),
					},
				},
			},
		}
		check := hasUnverified(commits)
		assert.Equal(t, true, check)
	})
	t.Run("Test All Commits are Verified", func(t *testing.T) {
		commits := []*github.RepositoryCommit{
			&github.RepositoryCommit{
				Commit: &github.Commit{
					Verification: &github.SignatureVerification{
						Verified: github.Bool(true),
					},
				},
			},
		}
		check := hasUnverified(commits)
		assert.Equal(t, false, check)
	})
	t.Run("Test Has Unsigned Commits", func(t *testing.T) {
		commits := []*github.RepositoryCommit{
			&github.RepositoryCommit{
				Commit: &github.Commit{
					Message: github.String("My Awesome Feature"),
				},
			},
		}
		check := hasUnsigned(commits)
		assert.Equal(t, true, check)
	})
	t.Run("Test Has Anonymous Signed Commits", func(t *testing.T) {
		commits := []*github.RepositoryCommit{
			&github.RepositoryCommit{
				Commit: &github.Commit{
					Message: github.String("Signed-off-by: User users@users.noreply.github.com"),
				},
			},
		}
		check := hasAnonymousSign(commits)
		assert.Equal(t, true, check)
	})
	t.Run("Test Has No Anonymous Signed Commits", func(t *testing.T) {
		commits := []*github.RepositoryCommit{
			&github.RepositoryCommit{
				Commit: &github.Commit{
					Message: github.String("Signed-off-by: test"),
				},
			},
		}
		check := hasAnonymousSign(commits)
		assert.Equal(t, false, check)
	})
	t.Run("Test Is Anonymous check", func(t *testing.T) {
		check := isAnonymousSign("Signed-off-by: User users@users.noreply.github.com")
		assert.Equal(t, true, check)
		check = isAnonymousSign("Signed-off-by: test")
		assert.Equal(t, false, check)
	})
	t.Run("Test Is Signed check", func(t *testing.T) {
		check := isSigned("Signed-off-by: User users@users.noreply.github.com")
		assert.Equal(t, true, check)
		check = isSigned("Awesome Feature")
		assert.Equal(t, false, check)
	})
	t.Run("Test Is Verified", func(t *testing.T) {
		check := isVerified(github.SignatureVerification{
			Verified: github.Bool(true),
		})
		assert.Equal(t, true, check)
		check = isVerified(github.SignatureVerification{
			Verified: github.Bool(false),
		})
		assert.Equal(t, false, check)
	})
}
func TestUpdateExistingDCOChecks(t *testing.T) {
	mClient, mux, _, teardown := test.GetMockClient()
	defer teardown()
	ctx := context.Background()
	webhookPayload := test.GetMockPayload("opened-pr")
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)
	t.Run("Test Update Successful check", func(t *testing.T) {
		check := &github.CheckRun{
			ID: github.Int64(12345),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/12345",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{
			            "id": 1,
                        "name":"DeveloperCertificateOfOrigin",
						"status": "completed",
						"conclusion": "sucess",
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
		err := updateExistingDCOCheck(ctx, mClient, e, check, success)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Update Unsuccessful check", func(t *testing.T) {
		check := &github.CheckRun{
			ID: github.Int64(12346),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/12346",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{
			            "id": 1,
                        "name":"DeveloperCertificateOfOrigin",
						"status": "completed",
						"conclusion": "sucess",
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
		err := updateExistingDCOCheck(ctx, mClient, e, check, failed)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Update Existing Check Invalid Response Code", func(t *testing.T) {
		check := &github.CheckRun{
			ID: github.Int64(1),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/1",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
		)
		err := updateExistingDCOCheck(ctx, mClient, e, check, failed)
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test Update Existing Check Invalid Response", func(t *testing.T) {
		check := &github.CheckRun{
			ID: github.Int64(2),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/2",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{`)
			},
		)
		err := updateExistingDCOCheck(ctx, mClient, e, check, failed)
		assert.NotEqual(t, nil, err)
	})
}

func TestCreateSuccessfulVerifyCheck(t *testing.T) {
	webhookPayload := test.GetMockPayload("opened-pr")
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)

	t.Run("Test get existing check works", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
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
		_, err := createSuccessfulVerifyCheck(context.Background(), e, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test get existing checks returns 2 fails", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":2,
                                "check_runs": [{
                                    "id": 1,
                                    "head_sha": "deadbeef",
                                    "status": "completed",
                                    "conclusion": "neutral",
                                    "started_at": "2018-05-04T01:14:52Z",
                                    "completed_at": "2018-05-04T01:14:52Z"}]}`)
			},
		)
		_, err := createSuccessfulVerifyCheck(context.Background(), e, mClient)
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test get existing checks returns invalid payload", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":2,*`)
			},
		)
		_, err := createSuccessfulVerifyCheck(context.Background(), e, mClient)
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test get existing checks creates new check", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":0}`)
			},
		)
		mux.HandleFunc("/repos/Spazzy757/paul/check-runs", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{
				  "id": 1,
				  "name": "DeveloperCertificateOfOrigin",
				  "head_sha":"deadbeef",
				  "status": "in_progress",
				  "conclusion": null,
				  "started_at": "2018-05-04T01:14:52Z",
				  "completed_at": null,
                  "output":{"title": "Mighty test report", "summary":"", "text":""}}`)
		})
		_, err := createSuccessfulVerifyCheck(context.Background(), e, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test get existing checks creates new check error", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":0}`)
			},
		)
		mux.HandleFunc("/repos/Spazzy757/paul/check-runs", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{`)
		})
		_, err := createSuccessfulVerifyCheck(context.Background(), e, mClient)
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test get existing checks creates new check returns bad code", func(t *testing.T) {
		mClient, mux, _, teardown := test.GetMockClient()
		defer teardown()
		mux.HandleFunc(
			"/repos/Spazzy757/paul/commits/"+e.PullRequest.Head.GetSHA()+"/check-runs",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"total_count":0}`)
			},
		)
		mux.HandleFunc("/repos/Spazzy757/paul/check-runs", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{
				  "id": 1,
				  "name": "DeveloperCertificateOfOrigin",
				  "head_sha":"deadbeef",
				  "status": "in_progress",
				  "conclusion": null,
				  "started_at": "2018-05-04T01:14:52Z",
				  "completed_at": null,
                  "output":{"title": "Mighty test report", "summary":"", "text":""}}`)
		})
		_, err := createSuccessfulVerifyCheck(context.Background(), e, mClient)
		assert.NotEqual(t, nil, err)
	})
}

func TestUpdateExistingVerifyChecks(t *testing.T) {
	mClient, mux, _, teardown := test.GetMockClient()
	defer teardown()
	ctx := context.Background()
	webhookPayload := test.GetMockPayload("opened-pr")
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
	e := event.(*github.PullRequestEvent)
	t.Run("Test Update Successful check", func(t *testing.T) {
		check := &github.CheckRun{
			ID: github.Int64(12345),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/12345",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{
			            "id": 1,
                        "name":"DeveloperCertificateOfOrigin",
						"status": "completed",
						"conclusion": "sucess",
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
		err := updateExistingVerifyCheck(ctx, mClient, e, check, success)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Update Unsuccessful check", func(t *testing.T) {
		check := &github.CheckRun{
			ID: github.Int64(12346),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/12346",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{
			            "id": 1,
                        "name":"Commits Are Verified",
						"status": "completed",
						"conclusion": "sucess",
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
		err := updateExistingVerifyCheck(ctx, mClient, e, check, failed)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Update Existing Check Invalid Response Code", func(t *testing.T) {
		check := &github.CheckRun{
			ID: github.Int64(1),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/1",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
		)
		err := updateExistingVerifyCheck(ctx, mClient, e, check, failed)
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test Update Existing Check Invalid Response", func(t *testing.T) {
		check := &github.CheckRun{
			ID: github.Int64(2),
		}
		mux.HandleFunc(
			"/repos/Spazzy757/paul/check-runs/2",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{`)
			},
		)
		err := updateExistingVerifyCheck(ctx, mClient, e, check, failed)
		assert.NotEqual(t, nil, err)
	})
}
