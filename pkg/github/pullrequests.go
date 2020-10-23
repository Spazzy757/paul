package github

import (
	"context"
	"fmt"

	"github.com/Spazzy757/paul/pkg/config"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
)

//PullRequestHandler handler for the pull request event
func PullRequestHandler(
	ctx context.Context,
	event *github.PullRequestEvent,
	client *github.Client,
) error {
	// Get Paul Config
	cfg, configErr := config.GetPaulConfig(
		ctx,
		event.Repo.Owner.Login,
		event.Repo.Name,
		event.Repo.GetContentsURL(),
		event.Repo.GetDefaultBranch(),
		client,
	)
	if configErr != nil {
		return configErr
	}
	var err error
	err = branchDestroyerCheck(ctx, cfg, client, event)
	if err != nil {
		return err
	}
	err = firstPRCheck(ctx, cfg, client, event)
	if err != nil {
		return err
	}
	err = limitPRCheck(ctx, cfg, client, event)
	return err
}

//firstPRCheck checks if a PR has just been opened and
func firstPRCheck(
	ctx context.Context,
	cfg types.PaulConfig,
	client *github.Client,
	event *github.PullRequestEvent,
) error {
	if cfg.PullRequests.OpenMessage != "" &&
		event.GetAction() == "opened" &&
		!checkStringInList(cfg.Maintainers, *event.Sender.Login) {
		err := reviewComment(
			ctx,
			event.GetPullRequest(),
			client,
			cfg.PullRequests.OpenMessage,
		)
		return err
	}
	return nil
}

//branchDestroyerCheck checks if branch can be destroyed
func branchDestroyerCheck(
	ctx context.Context,
	cfg types.PaulConfig,
	client *github.Client,
	event *github.PullRequestEvent,
) error {
	if cfg.BranchDestroyer.Enabled &&
		event.GetAction() == "closed" &&
		event.PullRequest.Head.GetRef() != event.Repo.GetDefaultBranch() &&
		event.PullRequest.GetMerged() &&
		!checkStringInList(
			cfg.BranchDestroyer.ProtectedBranches,
			event.PullRequest.Head.GetRef()) {
		err := branchDestroyer(
			ctx,
			event.GetPullRequest(),
			client,
			event.PullRequest.Head.GetRef(),
		)
		return err
	}
	return nil
}

//limitPRCheck
func limitPRCheck(
	ctx context.Context,
	cfg types.PaulConfig,
	client *github.Client,
	event *github.PullRequestEvent,
) error {
	prs, err := getPullRequestListForUser(ctx, client, event)
	if err != nil {
		return err
	}
	maxNumber := cfg.PullRequests.LimitPullRequests.MaxNumber
	if len(prs) > maxNumber && maxNumber != 0 {
		if err = reviewComment(
			ctx,
			event.GetPullRequest(),
			client,
			"You seem to have opened more PR's than this Repo Allows",
		); err != nil {
			return err
		}
		err = closePullRequest(ctx, client, event)
	}
	return err
}

//reviewComment sends a review comment to a Pull Request
func reviewComment(
	ctx context.Context,
	pr *github.PullRequest,
	client *github.Client,
	message string,
) error {
	pullRequestReviewRequest := &github.PullRequestReviewRequest{
		Body:  &message,
		Event: github.String("COMMENT"),
	}
	_, _, err := client.PullRequests.CreateReview(
		ctx,
		*pr.Base.User.Login,
		pr.Base.Repo.GetName(),
		pr.GetNumber(),
		pullRequestReviewRequest,
	)
	return err
}

//branchDestroyer will delete a branch
func branchDestroyer(
	ctx context.Context,
	pr *github.PullRequest,
	client *github.Client,
	branch string,
) error {
	_, err := client.Git.DeleteRef(
		ctx,
		pr.Base.User.GetLogin(),
		pr.Base.Repo.GetName(),
		fmt.Sprintf("refs/heads/%v", branch),
	)
	return err
}

// getPullRequestListForUser gets all the Pull Requests for the user
func getPullRequestListForUser(
	ctx context.Context,
	client *github.Client,
	event *github.PullRequestEvent,
) ([]*github.PullRequest, error) {
	listOptions := &github.PullRequestListOptions{
		Head: event.Sender.GetLogin(),
		Base: event.PullRequest.Base.GetRef(),
	}
	prList, _, err := client.PullRequests.List(
		ctx,
		event.PullRequest.Base.User.GetLogin(),
		event.PullRequest.Base.Repo.GetName(),
		listOptions,
	)
	return prList, err
}

// getPullRequestListForUser gets all the Pull Requests for the user
func closePullRequest(
	ctx context.Context,
	client *github.Client,
	event *github.PullRequestEvent,
) error {
	pr := event.PullRequest
	updatedPr := &github.PullRequest{
		State: github.String("closed"),
	}
	_, _, err := client.PullRequests.Edit(
		ctx,
		pr.Base.User.GetLogin(),
		pr.Base.Repo.GetName(),
		pr.GetNumber(),
		updatedPr,
	)
	return err
}
