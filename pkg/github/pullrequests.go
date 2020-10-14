package github

import (
	"log"

	"github.com/google/go-github/v32/github"
)

//PullRequestHandler handler for the pull request event
func PullRequestHandler(event *github.PullRequestEvent) {
	client, ctx := getClient(*event.Installation.ID)
	// Get Paul Config
	rc := &repoClient{ctx: ctx, repoService: client.Repositories}
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
		pr := &pullRequestClient{ctx: ctx, pullRequestService: client.PullRequests}
		_ = reviewComment(event.GetPullRequest(), pr, cfg.PullRequests.OpenMessage)
	}
}

func reviewComment(pr *github.PullRequest, client *pullRequestClient, message string) error {
	pullRequestReviewRequest := &github.PullRequestReviewRequest{
		Body:  &message,
		Event: github.String("COMMENT"),
	}

	_, _, err := client.pullRequestService.CreateReview(
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
