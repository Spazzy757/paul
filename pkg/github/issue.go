package github

import (
	"context"
	"fmt"
	"github.com/Spazzy757/paul/pkg/cats"
	"github.com/google/go-github/v32/github"
	"log"
)

/*
IssueCommentHandler takes an incoming event of type IssueCommentEvent and
runs logic against it
*/
func IssueCommentHandler(event *github.IssueCommentEvent) {
	// load github client
	client, ctx := getClient(*event.Installation.ID)
	// load Paul Config from repo
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

	// Check comments for any commands
	// TODO: Make this handle more commands
	if *event.Action == "created" {
		comment := event.GetComment()
		// Check command is /cat and see if repo has cats enabled
		if *comment.Body == "/cat" && cfg.PullRequests.CatsEnabled {
			handleCats(event, client, ctx)
		}
	}
}

func handleCats(is *github.IssueCommentEvent, gclient *github.Client, ctx context.Context) {
	catClient := cats.NewClient()
	cat, err := catClient.GetCat()
	if err != nil {
		log.Fatalf("Error Occurred Fetching Cats: %v", err)
	}
	isClient := &issueClient{
		ctx:    ctx,
		client: gclient.Issues,
	}
	message := fmt.Sprintf("I present my minion\n ![my favorite minion](%v)", cat.Url)
	catErr := createIssueComment(is, isClient, message)
	if catErr != nil {
		log.Println("Failed Commenting Cats")
	}
}

// interface to make testing logic easier
type issue interface {
	CreateComment(
		ctx context.Context,
		owner string,
		repo string,
		number int,
		comment *github.IssueComment,
	) (*github.IssueComment, *github.Response, error)
}

// struct to make testing logic easier
type issueClient struct {
	ctx    context.Context
	client issue
}

// createIssueComment sends a comment to an issue/pull request
func createIssueComment(
	is *github.IssueCommentEvent,
	client *issueClient,
	message string,
) error {
	comment := &github.IssueComment{Body: &message}
	_, _, err := client.client.CreateComment(
		client.ctx,
		*is.Repo.Owner.Login,
		is.Repo.GetName(),
		is.Issue.GetNumber(),
		comment,
	)
	if err != nil {
		return err
	}
	return nil
}
