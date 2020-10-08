package github

import (
	"bytes"
	"context"
	"github.com/google/go-github/v32/github"
	"io/ioutil"
	"net/http"
	"testing"
)

func getIssueCommentMockPayload() []byte {
	file, _ := ioutil.ReadFile("../../mocks/comment.json")
	return []byte(file)
}

type mockIssueClient struct {
	resp *github.IssueComment
}

func (m *mockIssueClient) CreateComment(
	ctx context.Context,
	owner, repo string,
	number int,
	review *github.IssueComment,
) (*github.IssueComment, *github.Response, error) {
	return m.resp, nil, nil
}

func TestCreateComment(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload()

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		ctx := context.Background()
		mc := &mockIssueClient{
			resp: &github.IssueComment{
				ID: github.Int64(1),
			},
		}
		pr := &issueClient{ctx: ctx, client: mc}
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		switch e := event.(type) {
		case *github.IssueCommentEvent:
			if err := createIssueComment(e, pr, "test"); err != nil {
				t.Fatalf("comment on issue: %v", err)
			}
		default:
			t.Fatalf("Event Type Not Issue Comment")
		}
	})
}
