package github

import (
	"context"
	"fmt"

	"github.com/Spazzy757/paul/pkg/config"
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v36/github"
)

const (
	emptyDescriptionMessage = "There seems to be no description in your Pull Request.Please add an understanding of what this change proposes to do and why it is needed"
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
		event.Repo.Owner.GetLogin(),
		event.Repo.GetName(),
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
	err = emptyDescriptionCheck(ctx, cfg, client, event)
	if err != nil {
		return err
	}
	err = limitPRCheck(ctx, cfg, client, event)
	if err != nil {
		return err
	}
	err = verifiedCommitCheck(ctx, cfg, client, event)
	if err != nil {
		return err
	}
	err = dcoCheck(ctx, cfg, client, event)
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

// emptyDescriptionCheck checks if there is a description
func emptyDescriptionCheck(
	ctx context.Context,
	cfg types.PaulConfig,
	client *github.Client,
	event *github.PullRequestEvent,
) error {
	if cfg.EmptyDescriptionCheck.Enabled &&
		event.GetAction() == "opened" &&
		event.PullRequest.GetBody() == "" {
		message := emptyDescriptionMessage
		if cfg.EmptyDescriptionCheck.Message != "" {
			message = cfg.EmptyDescriptionCheck.Message
		}
		err := reviewComment(
			ctx,
			event.GetPullRequest(),
			client,
			message,
		)
		if cfg.EmptyDescriptionCheck.Enforced {
			err = closePullRequest(ctx, client, event)
		}
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

//dcoCheck
func dcoCheck(
	ctx context.Context,
	cfg types.PaulConfig,
	client *github.Client,
	event *github.PullRequestEvent,
) error {
	if cfg.PullRequests.DCOCheck {
		check, err := createSuccessfulDCOCheck(ctx, event, client)
		if err != nil {
			return err
		}
		commits, err := getPullRequestCommits(ctx, event, client)
		if err != nil {
			return err
		}
		anonymousSign := hasAnonymousSign(commits)
		unsignedCommits := hasUnsigned(commits)
		if unsignedCommits || anonymousSign {
			err = updateExistingDCOCheck(ctx, client, event, check, failed)
		} else {
			err = updateExistingDCOCheck(ctx, client, event, check, success)
		}
		return err
	}
	return nil
}

//verifiedCommitCheck
func verifiedCommitCheck(
	ctx context.Context,
	cfg types.PaulConfig,
	client *github.Client,
	event *github.PullRequestEvent,
) error {
	if cfg.PullRequests.VerifiedCommitCheck {
		check, err := createSuccessfulVerifyCheck(ctx, event, client)
		if err != nil {
			return err
		}
		commits, err := getPullRequestCommits(ctx, event, client)
		if err != nil {
			return err
		}
		unverifiedCommits := hasUnverified(commits)
		if unverifiedCommits {
			err = updateExistingVerifyCheck(ctx, client, event, check, failed)
		} else {
			err = updateExistingVerifyCheck(ctx, client, event, check, success)
		}
		return err
	}
	return nil
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

// mergePullRequest will merge a pull request with the rebase feature
func mergePullRequest(
	ctx context.Context,
	client *github.Client,
	pr *github.PullRequest,
) error {
	options := &github.PullRequestOptions{
		MergeMethod: "merge",
	}
	_, _, err := client.PullRequests.Merge(
		ctx,
		pr.Base.Repo.Owner.GetLogin(),
		pr.Base.Repo.GetName(),
		pr.GetNumber(),
		"",
		options,
	)
	return err
}
