package github

import (
	"net/http"

	paulclient "github.com/Spazzy757/paul/pkg/client"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/google/go-github/v32/github"
)

//IncomingWebhook handles an incoming webhook request
func IncomingWebhook(r *http.Request, client *paulclient.GithubClient) error {
	// handle authentication
	secret_key := helpers.GetEnv("SECRET_KEY", "")
	payload, validationErr := github.ValidatePayload(r, []byte(secret_key))
	if validationErr != nil {
		return validationErr
	}
	event, parseErr := github.ParseWebHook(github.WebHookType(r), payload)
	if parseErr != nil {
		return parseErr
	}
	var err error
	switch e := event.(type) {
	case *github.IssueCommentEvent:
		err = IssueCommentHandler(e, client)
	case *github.PullRequestEvent:
		err = PullRequestHandler(e, client)
	default:
		break
	}
	return err
}
