package github

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Spazzy757/paul/pkg/animals"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func getIssueCommentMockPayload(payloadType string) []byte {
	fileLocation := fmt.Sprintf("../../mocks/%v.json", payloadType)
	file, _ := ioutil.ReadFile(fileLocation)
	return []byte(file)
}

type mockIssueService struct {
	resp   *github.IssueComment
	labels []*github.Label
}

func (m *mockIssueService) CreateComment(
	ctx context.Context,
	owner, repo string,
	number int,
	review *github.IssueComment,
) (*github.IssueComment, *github.Response, error) {
	return m.resp, nil, nil
}

func (m *mockIssueService) AddLabelsToIssue(
	ctx context.Context,
	owner, repo string,
	number int,
	labels []string,
) ([]*github.Label, *github.Response, error) {
	return m.labels, nil, nil
}

func TestCreateComment(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("pr")

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		ctx := context.Background()
		mc := &mockIssueService{
			resp: &github.IssueComment{
				ID: github.Int64(1),
			},
		}
		pr := &issueClient{ctx: ctx, issueService: mc}
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
			w.Write([]byte(catAPIResponse))
		})
		httpClient, teardown := helpers.MockHTTPClient(h)
		defer teardown()

		catClient := animals.NewCatClient()
		catClient.HttpClient = httpClient
		catClient.Url = "https://example.com"

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		ctx := context.Background()
		mc := &mockIssueService{
			resp: &github.IssueComment{
				ID: github.Int64(1),
			},
		}
		is := &issueClient{ctx: ctx, issueService: mc}
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		switch e := event.(type) {
		case *github.IssueCommentEvent:
			if err := catsHandler(e, is, catClient); err != nil {
				t.Fatalf("comment on issue: %v", err)
			}
		default:
			t.Fatalf("Event Type Not Issue Comment")
		}
	})
}

func TestHandleDogs(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
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
			w.Write([]byte(dogAPIResponse))
		})
		httpClient, teardown := helpers.MockHTTPClient(h)
		defer teardown()

		dogClient := animals.NewDogClient()
		dogClient.HttpClient = httpClient
		dogClient.Url = "https://example.com"

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		ctx := context.Background()
		mc := &mockIssueService{
			resp: &github.IssueComment{
				ID: github.Int64(1),
			},
		}
		is := &issueClient{ctx: ctx, issueService: mc}
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		switch e := event.(type) {
		case *github.IssueCommentEvent:
			if err := dogsHandler(e, is, dogClient); err != nil {
				t.Fatalf("comment on issue: %v", err)
			}
		default:
			t.Fatalf("Event Type Not Issue Comment")
		}
	})
}

func TestHandleLabels(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
		// Just needed to get the right event type
		webhookPayload := getIssueCommentMockPayload("dog-command")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")

		ctx := context.Background()
		mc := &mockIssueService{
			resp: &github.IssueComment{},
		}
		is := &issueClient{ctx: ctx, issueService: mc}
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		switch e := event.(type) {
		case *github.IssueCommentEvent:
			err := labelHandler(e, is, []string{"test"})
			assert.Equal(t, nil, err)
		default:
			t.Fatalf("Event Type Not Issue Comment")
		}
	})
}

func TestMaintainerCheck(t *testing.T) {
	maintainers := []string{"yes", "no", "maybe"}
	t.Run("Test Maintainer returns true", func(t *testing.T) {
		assert.Equal(t, true, maintainerCheck(maintainers, "yes"))
	})
	t.Run("Test Non Maintainer returns false", func(t *testing.T) {
		assert.Equal(t, false, maintainerCheck(maintainers, "I don't know"))
	})
}
