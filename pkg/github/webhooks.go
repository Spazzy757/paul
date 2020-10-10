package github

import (
	"log"
	"net/http"

	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/google/go-github/v32/github"
)

//IncomingWebhook handles an incoming webhook request
func IncomingWebhook(r *http.Request) error {
	// handle authentication
	secret_key := helpers.GetEnv("SECRET_KEY", "")
	payload, validationErr := github.ValidatePayload(r, []byte(secret_key))
	if validationErr != nil {
		return validationErr
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		return err
	}

	switch e := event.(type) {
	case *github.IssueCommentEvent:
		IssueCommentHandler(e)
	case *github.PullRequestEvent:
		PullRequestHandler(e)
	default:
		log.Printf("unknown event type %s\n", github.WebHookType(r))
		return nil
	}
	return nil
}
