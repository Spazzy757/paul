package github

import (
	"bytes"
	"context"
	"github.com/Spazzy757/paul/pkg/cats"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
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

func TestHandleCats(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload()
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
			w.Write([]byte(catAPIResponse))
		})
		httpClient, teardown := helpers.MockHTTPClient(h)
		defer teardown()

		catClient := cats.NewClient()
		catClient.HttpClient = httpClient
		catClient.Url = "https://example.com"

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		ctx := context.Background()
		mc := &mockIssueClient{
			resp: &github.IssueComment{
				ID: github.Int64(1),
			},
		}
		is := &issueClient{ctx: ctx, client: mc}
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		switch e := event.(type) {
		case *github.IssueCommentEvent:
			if err := handleCats(e, is, catClient); err != nil {
				t.Fatalf("comment on issue: %v", err)
			}
		default:
			t.Fatalf("Event Type Not Issue Comment")
		}
	})
}
