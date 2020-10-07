package github

import (
	"context"
	"github.com/Spazzy757/paul/pkg/config"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

const githubApi = "https://api.github.com"

const openPRMessage = "Thanks for Opening a PR In this Repo!\nWe are validating the PR and will be in contact shortly"

type pullRequest interface {
	CreateReview(ctx context.Context, owner string, repo string, number int, review *github.PullRequestReviewRequest) (*github.PullRequestReview, *github.Response, error)
}

type pullRequestClient struct {
	ctx    context.Context
	client pullRequest
}

func IncomingWebhook(data []byte, r *http.Request) {
	ctx := context.Background()
	event, err := github.ParseWebHook(github.WebHookType(r), data)
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		return
	}
	//payloadSecret := r.Header.Get("X-Hub-Signature")
	//payload, err := github.ValidatePayload(r, []byte("my-secret-key"))
	switch e := event.(type) {
	case *github.PushEvent:
		// this is a commit push, do something with it
	case *github.PullRequestEvent:
		gclient := getClient(*e.Installation.ID)
		pr := &pullRequestClient{ctx: ctx, client: gclient.PullRequests}
		// this is a pull request, do something with it
		if e.Action != nil && *e.Action == "opened" {
			_ = comment(e.PullRequest, pr, openPRMessage)
		}
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

func getClient(installationId int64) *github.Client {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Can't load config: %v", err)
	}
	token, tokenErr := helpers.GetAccessToken(cfg, installationId)
	if tokenErr != nil {
		log.Fatalf("Can't load config: %v", err)
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return client
}

func comment(pr *github.PullRequest, client *pullRequestClient, message string) error {
	pullRequestReviewRequest := &github.PullRequestReviewRequest{
		Body:  &message,
		Event: github.String("COMMENT"),
	}

	_, _, err := client.client.CreateReview(
		client.ctx,
		*pr.Base.User.Login,
		pr.Base.Repo.GetName(),
		pr.GetNumber(),
		pullRequestReviewRequest,
	)
	if err != nil {
		return err
	}
	return nil
}
