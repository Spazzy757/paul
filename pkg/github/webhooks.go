package github

import (
	"context"
	"net/http"

	"github.com/google/go-github/v34/github"
)

//IncomingWebhook handles an incoming webhook request
func IncomingWebhook(
	ctx context.Context,
	r *http.Request,
	payload []byte,
	client *github.Client,
) error {
	event, parseErr := github.ParseWebHook(github.WebHookType(r), payload)
	if parseErr != nil {
		return parseErr
	}
	var err error
	switch e := event.(type) {
	case *github.IssueCommentEvent:
		err = IssueCommentHandler(ctx, e, client)
	case *github.PullRequestEvent:
		err = PullRequestHandler(ctx, e, client)
	default:
		break
	}
	return err
}
