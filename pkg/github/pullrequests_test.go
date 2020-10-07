package github

import (
	"bytes"
	"context"
	"github.com/google/go-github/v32/github"
	"io/ioutil"
	"net/http"
	"testing"
)

func getMockPayload() []byte {
	file, _ := ioutil.ReadFile("../../mocks/pr.json")
	return []byte(file)
}

type mockClient struct {
	resp *github.PullRequestReview
}

func (m *mockClient) CreateReview(ctx context.Context, owner string, repo string, number int, review *github.PullRequestReviewRequest) (*github.PullRequestReview, *github.Response, error) {
	return m.resp, nil, nil
}

func TestCreateReview(t *testing.T) {
	t.Run("Test Webhook is Handled correctly", func(t *testing.T) {
		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")
		ctx := context.Background()
		mc := &mockClient{
			resp: &github.PullRequestReview{
				ID: github.Int64(1),
			},
		}
		pr := &pullRequestClient{ctx: ctx, client: mc}
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		switch e := event.(type) {
		case *github.PullRequestEvent:
			if err := comment(e.PullRequest, pr, "test"); err != nil {
				t.Fatalf("createReview: %v", err)
			}
		default:
			t.Fatalf("Event Type Not Pull Request")
		}
	})
}
