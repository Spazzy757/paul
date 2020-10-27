package router

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/Spazzy757/paul/pkg/test"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGetInstallationId(t *testing.T) {
	t.Run("Test Get Installation ID from Payload", func(t *testing.T) {
		mockPayload := []byte(`
        {
          "installation": {"id": 1}
        }
        `)
		id, err := getInstallationId(mockPayload)
		assert.Equal(t, int64(1), id)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Get Installation ID from Payload Fails", func(t *testing.T) {
		mockPayload := []byte(`
        {
          "installation": {"id": 1}
        `)
		id, err := getInstallationId(mockPayload)
		assert.Equal(t, int64(0), id)
		assert.NotEqual(t, nil, err)
	})
}

func TestHandleError(t *testing.T) {
	w := httptest.NewRecorder()
	t.Run("Test Handle Error with no Error", func(t *testing.T) {
		check := handleError(w, nil)
		assert.Equal(t, false, check)
	})
	t.Run("Test Handle Error with Error", func(t *testing.T) {
		check := handleError(w, fmt.Errorf("test"))
		assert.Equal(t, true, check)
	})
}

func TestGetRouter(t *testing.T) {
	t.Run("Test Returns the Router", func(t *testing.T) {
		r := GetRouter()
		expected := &mux.Router{}
		assert.Equal(t, reflect.TypeOf(expected), reflect.TypeOf(r))
	})
}

func TestGithubWebHookHandler(t *testing.T) {
	os.Setenv("SECRET_KEY", "test")
	t.Run("Test Fails on Get Client", func(t *testing.T) {
		webhookPayload := test.GetMockPayload("label-command")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		req.Header.Set("Content-Type", "application/json")
		signature := generateGitHubSha("test", webhookPayload)
		req.Header.Set("X-Hub-Signature", signature)

		GithubWebHookHandler(w, req)
		response := w.Result()
		assert.Equal(t, 400, response.StatusCode)
	})
	t.Run("Test Fails on Validation", func(t *testing.T) {
		webhookPayload := test.GetMockPayload("label-command")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		req.Header.Set("Content-Type", "application/json")
		signature := generateGitHubSha("not_valid", webhookPayload)
		req.Header.Set("X-Hub-Signature", signature)

		GithubWebHookHandler(w, req)
		response := w.Result()
		assert.Equal(t, 400, response.StatusCode)
	})
	t.Run("Test Fails on Getting InstallationID", func(t *testing.T) {
		webhookPayload := []byte(`
        {
          "installation": {"id": 1}
        `)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(webhookPayload))
		req.Header.Set("X-GitHub-Event", "issue_comment")
		req.Header.Set("Content-Type", "application/json")
		signature := generateGitHubSha("not_valid", webhookPayload)
		req.Header.Set("X-Hub-Signature", signature)

		GithubWebHookHandler(w, req)
		response := w.Result()
		assert.Equal(t, 400, response.StatusCode)
	})

}

func generateGitHubSha(secret string, body []byte) string {
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write(body)
	encodeString := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("sha1=%v", encodeString)
}
