package github

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/Spazzy757/paul/pkg/animals"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/Spazzy757/paul/pkg/test"
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
		webhookPayload := getIssueCommentMockPayload("pr")

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		mClient := test.GetMockClient()
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := createIssueComment(e, mClient, "test")
		assert.Equal(t, nil, err)
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
		mClient := test.GetMockClient()

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := catsHandler(e, mClient, catClient)
		assert.Equal(t, nil, err)
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
		mClient := test.GetMockClient()
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := dogsHandler(e, mClient, dogClient)
		assert.Equal(t, nil, err)
	})
}

func TestHandleLabels(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
		// Just needed to get the right event type
		webhookPayload := getIssueCommentMockPayload("dog-command")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")

		mClient := test.GetMockClient()
		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := labelHandler(e, mClient, []string{"test"})
		assert.Equal(t, nil, err)
	})
}

func TestHandleRemoveLabels(t *testing.T) {
	t.Run("Test Issue Comment Webhook is Handled correctly", func(t *testing.T) {
		// Just needed to get the right event type
		webhookPayload := getIssueCommentMockPayload("dog-command")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		mClient := test.GetMockClient()

		event, _ := github.ParseWebHook(github.WebHookType(req), webhookPayload)
		e := event.(*github.IssueCommentEvent)
		err := removeLabelHandler(e, mClient, "test")
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
