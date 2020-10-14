package github

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

func getMockPayload() []byte {
	file, _ := ioutil.ReadFile("../../mocks/pr.json")
	return []byte(file)
}

type mockPullRequestService struct {
	resp *github.PullRequestReview
}

func (m *mockPullRequestService) CreateReview(
	ctx context.Context,
	owner, repo string,
	number int,
	review *github.PullRequestReviewRequest,
) (*github.PullRequestReview, *github.Response, error) {
	return m.resp, nil, nil
}

func TestCreateReview(t *testing.T) {
	t.Run("Test Webhook is Handled correctly", func(t *testing.T) {
		webhookPayload := getMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")
		ctx := context.Background()
		mc := &mockPullRequestService{}
		pr := &pullRequestClient{ctx: ctx, pullRequestService: mc}
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.PullRequestEvent)
		err := reviewComment(e.PullRequest, pr, "test")
		assert.Equal(t, nil, err)
	})
}
