package github

import (
	"context"
	"log"

	"github.com/google/go-github/v32/github"
)

//PullRequestHandler handler for the pull request event
func PullRequestHandler(event *github.PullRequestEvent) {
	client, ctx := getClient(*event.Installation.ID)
	// Get Paul Config
	rc := &repoClient{ctx: ctx, client: client.Repositories}
	cfg, err := getPaulConfig(
		event.Repo.Owner.Login,
		event.Repo.Name,
		event.Repo.GetContentsURL(),
		event.Repo.GetDefaultBranch(),
		rc,
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
