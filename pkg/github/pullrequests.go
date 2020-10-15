package github

import (
	"github.com/Spazzy757/paul/pkg/types"
	"github.com/google/go-github/v32/github"
)

//PullRequestHandler handler for the pull request event
func PullRequestHandler(event *github.PullRequestEvent) error {
	client, ctx, clientErr := getClient(*event.Installation.ID)
	if clientErr != nil {
		return clientErr
	}
	// Get Paul Config
	rc := &repoClient{ctx: ctx, repoService: client.Repositories}
	cfg, configErr := getPaulConfig(
		event.Repo.Owner.Login,
		event.Repo.Name,
		event.Repo.GetContentsURL(),
		event.Repo.GetDefaultBranch(),
		rc,
	)
	if configErr != nil {
		return configErr
	}
	var err error
	if firstPRCheck(cfg.PullRequests.OpenMessage, *event.Action) {
		client := &pullRequestClient{
			ctx:                ctx,
			pullRequestService: client.PullRequests,
		}
		err = reviewComment(
			event.GetPullRequest(),
			client,
			cfg.PullRequests.OpenMessage,
		)
	}
	if branchDestroyerCheck(
		&cfg.BranchDestroyer,
		*event.Action,
		event.Repo.GetDefaultBranch(),
		event.PullRequest.Head.GetRef(),
	) {
		client := &gitClient{
			ctx:        ctx,
			gitService: client.Git,
		}
		err = branchDestroyer(
			event.GetPullRequest(),
			client,
			event.PullRequest.Head.GetRef(),
		)
	}
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

//reviewComment sends a review comment to a Pull Request
func reviewComment(
	pr *github.PullRequest,
	client *pullRequestClient,
	message string,
) error {
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
	return err
}

//branchDestroyer will delete a branch
func branchDestroyer(
	pr *github.PullRequest,
	client *gitClient,
	branch string,
) error {
	_, err := client.gitService.DeleteRef(
		client.ctx,
		*pr.Base.User.Login,
		pr.Base.Repo.GetName(),
		branch,
	)
	return err
}
