package github

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncomingWebhook(t *testing.T) {
	os.Setenv("SECRET_KEY", "test")
	t.Run("Test Unknown Webhook Payload is Handled correctly", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("installation")

		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "installation")
		req.Header.Set("Content-Type", "application/json")
		signature := generateGitHubSha("test", webhookPayload)
		req.Header.Set("X-Hub-Signature", signature)
		err := IncomingWebhook(req)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Incoming Webhook Fails if payload is incorrect", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`1234`)))
		req.Header.Set("X-GitHub-Event", "installation")
		req.Header.Set("Content-Type", "application/json")
		signature := generateGitHubSha("test", []byte(`1234`))
		req.Header.Set("X-Hub-Signature", signature)
		err := IncomingWebhook(req)
		assert.NotEqual(t, nil, err)
	})
	t.Run("Test Incoming Webhook Validation Fails if Not Correct", func(t *testing.T) {
		webhookPayload := getIssueCommentMockPayload("installation")
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "installation")
		req.Header.Set("Content-Type", "application/json")
		signature := generateGitHubSha("notcorrect", webhookPayload)
		req.Header.Set("X-Hub-Signature", signature)
		err := IncomingWebhook(req)
		assert.NotEqual(t, nil, err)
	})

}

func generateGitHubSha(secret string, body []byte) string {
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write(body)
	encodeString := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("sha1=%v", encodeString)
}
