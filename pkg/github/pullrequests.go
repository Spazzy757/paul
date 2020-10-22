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
	if firstPRCheck(cfg.PullRequests.OpenMessage, *event.Action) {
		err = reviewComment(
			ctx,
			event.GetPullRequest(),
			client,
			cfg.PullRequests.OpenMessage,
		)
	}
	if branchDestroyerCheck(
		&cfg.BranchDestroyer,
		event.GetAction(),
		event.Repo.GetDefaultBranch(),
		event.PullRequest.Head.GetRef(),
	) {
		err = branchDestroyer(
			ctx,
			event.GetPullRequest(),
			client,
			event.PullRequest.Head.GetRef(),
		)
	}
	err = limitPRCheck(ctx, client, event, &cfg.PullRequests.LimitPullRequests)
	return err

}

//firstPRCheck checks if a PR has just been opened and
func firstPRCheck(message, action string) bool {
	return message != "" && action == "opened"
}

//branchDestroyerCheck checks if branch can be destroyed
func branchDestroyerCheck(
	cfg *types.BranchDestroyer,
	action, defaultBranch, destroyBranch string,
) bool {
	return cfg.Enabled &&
		action == "completed" &&
		destroyBranch != defaultBranch &&
		!checkStringInList(cfg.ProtectedBranches, destroyBranch)
}

//limitPRCheck
func limitPRCheck(
	ctx context.Context,
	client *github.Client,
	event *github.PullRequestEvent,
	limitCfg *types.LimitPullRequests,
) error {
	prs, err := getPullRequestListForUser(ctx, client, event)
	if err != nil {
		return err
	}
	if len(prs) > limitCfg.MaxNumber && limitCfg.MaxNumber != 0 {
		if err = closePullRequest(ctx, client, event); err != nil {
			return err
		}
		err = reviewComment(
			ctx,
			event.GetPullRequest(),
			client,
			"You seem to have opened more PR's than this Repo Allows",
		)
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
