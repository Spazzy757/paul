package github

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/Spazzy757/paul/pkg/test"
	"github.com/stretchr/testify/assert"
)

const (
	// baseURLPath is a non-empty Client.BaseURL path to use during tests,
	// to ensure relative URLs are used for all endpoints. See issue #752.
	baseURLPath = "/api-v3"
)

func TestIncomingWebhook(t *testing.T) {
	os.Setenv("SECRET_KEY", "test")
	t.Run("Test Unknown Webhook Payload is Handled correctly", func(t *testing.T) {
		mClient, _, _, teardown := test.GetMockClient()
		defer teardown()

		webhookPayload := getIssueCommentMockPayload("installation")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "installation")
		req.Header.Set("Content-Type", "application/json")
		signature := generateGitHubSha("test", webhookPayload)
		req.Header.Set("X-Hub-Signature", signature)
		err := IncomingWebhook(context.Background(), req, webhookPayload, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Incoming Webhook Fails if payload is incorrect", func(t *testing.T) {
		mClient, _, _, teardown := test.GetMockClient()
		defer teardown()
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`1234`)))
		req.Header.Set("X-GitHub-Event", "installation")
		req.Header.Set("Content-Type", "application/json")
		signature := generateGitHubSha("test", []byte(`1234`))
		req.Header.Set("X-Hub-Signature", signature)
		err := IncomingWebhook(context.Background(), req, []byte(`1234`), mClient)
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test Incoming Webhook Validation if Not supported", func(t *testing.T) {
		mClient, _, _, teardown := test.GetMockClient()
		defer teardown()
		webhookPayload := getIssueCommentMockPayload("installation")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "installation")
		req.Header.Set("Content-Type", "application/json")
		signature := generateGitHubSha("notcorrect", webhookPayload)
		req.Header.Set("X-Hub-Signature", signature)
		err := IncomingWebhook(context.Background(), req, webhookPayload, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Incoming Webhook checks PR", func(t *testing.T) {
		mClient, mux, serverURL, teardown := test.GetMockClient()
		defer teardown()
		yamlFile, err := ioutil.ReadFile("../../.github/PAUL.yaml")
		assert.Equal(t, nil, err)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/pulls",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "GET")
				fmt.Fprint(w, `[{"number":1}]`)
			},
		)
		mux.HandleFunc(
			"/repos/Spazzy757/paul/contents/.github",
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
			"/repos/Spazzy757/paul/git/refs/heads/feature-added-webserver",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, "DELETE")
			},
		)

		webhookPayload := getIssueCommentMockPayload("merged-pr")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "pull_request")
		req.Header.Set("Content-Type", "application/json")
		signature := generateGitHubSha("test", webhookPayload)
		req.Header.Set("X-Hub-Signature", signature)
		err = IncomingWebhook(context.Background(), req, webhookPayload, mClient)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Incoming Webhook checks Issue Comment", func(t *testing.T) {
		mClient, mux, serverURL, teardown := test.GetMockClient()
		defer teardown()
		yamlFile, err := ioutil.ReadFile("../../.github/PAUL.yaml")
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

		labelInput := []string{"enhancement"}

		mux.HandleFunc(
			"/repos/Spazzy757/paul/issues/9/labels",
			func(w http.ResponseWriter, r *http.Request) {
				var v []string
				_ = json.NewDecoder(r.Body).Decode(&v)
				assert.Equal(t, v, labelInput)
				fmt.Fprint(w, `[{"url":"u"}]`)
			},
		)

		assert.Equal(t, nil, err)

		webhookPayload := getIssueCommentMockPayload("label-command")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		req.Header.Set("Content-Type", "application/json")
		signature := generateGitHubSha("test", webhookPayload)
		req.Header.Set("X-Hub-Signature", signature)

		err = IncomingWebhook(context.Background(), req, webhookPayload, mClient)
		assert.Equal(t, nil, err)
	})

}

func generateGitHubSha(secret string, body []byte) string {
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write(body)
	encodeString := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("sha1=%v", encodeString)
}
