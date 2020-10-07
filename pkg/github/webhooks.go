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
	case *github.PushEvent:
		// this is a commit push, do something with it
	case *github.PullRequestEvent:
		RunPullRequestChecks(e)
	case *github.WatchEvent:
		// https://developer.github.com/v3/activity/events/types/#watchevent
		// someone starred our repository
		if e.Action != nil && *e.Action == "starred" {
			log.Printf("%s starred repository %s\n",
				*e.Sender.Login, *e.Repo.FullName)
		}
	default:
		log.Printf("unknown event type %s\n", github.WebHookType(r))
		return
	}
}
