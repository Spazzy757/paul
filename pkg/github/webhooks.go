package github

import (
	"github.com/google/go-github/v32/github"
	"log"
	"net/http"
)

func IncomingWebhook(data []byte, r *http.Request) {
	event, err := github.ParseWebHook(github.WebHookType(r), data)
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		return
	}
	// TODO: Create way to handle X-Hub-Signature
	//payloadSecret := r.Header.Get("X-Hub-Signature")
	//payload, err := github.ValidatePayload(r, []byte("my-secret-key"))
	switch e := event.(type) {
	case *github.IssueCommentEvent:
		IssueCommentHandler(e)
	case *github.PullRequestEvent:
		PullRequestHandler(e)
	default:
		log.Printf("unknown event type %s\n", github.WebHookType(r))
		return
	}
}
