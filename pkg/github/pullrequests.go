package github

import (
	"context"
	"github.com/google/go-github/v32/github"
	"log"
)

func PullRequestHandler(event *github.PullRequestEvent) {
	client, ctx := getClient(*event.Installation.ID)
	cfg, err := getPaulConfig(
		event.Repo.Owner.Login,
		event.Repo.Name,
		client,
		event.Repo.GetContentsURL(),
		ctx,
	)
	if err != nil {
		log.Fatalf("An error occurred fetching config %v", err)
	}
	if cfg.PullRequests.OpenMessage != "" && *event.Action == "opened" {
		pr := &pullRequestClient{ctx: ctx, client: client.PullRequests}
		_ = comment(event.GetPullRequest(), pr, cfg.PullRequests.OpenMessage)
	}
	// Check comments for any commands
	if *event.Action == "created" {
		changes := event.GetChanges()
		log.Printf("%+v", changes.Body)
	}
}

type pullRequest interface {
	CreateReview(
		ctx context.Context,
		owner string,
		repo string,
		number int,
		review *github.PullRequestReviewRequest,
	) (*github.PullRequestReview, *github.Response, error)
}

type pullRequestClient struct {
	ctx    context.Context
	client pullRequest
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
