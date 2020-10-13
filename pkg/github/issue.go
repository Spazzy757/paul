package github

import (
	"context"
	"fmt"
	"github.com/Spazzy757/paul/pkg/animals"
	"github.com/google/go-github/v32/github"
	"log"
	"strings"
)

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

/*
IssueCommentHandler takes an incoming event of type IssueCommentEvent and
runs logic against it
*/
func IssueCommentHandler(event *github.IssueCommentEvent) {
	// load github client
	client, ctx := getClient(*event.Installation.ID)
	// load Paul Config from repo
	rc := &repoClient{ctx: ctx, client: client.Repositories}
	cfg, err := getPaulConfig(
		event.Repo.Owner.Login,
		event.Repo.Name,
		event.Repo.GetContentsURL(),
		rc,
	)
	if err != nil {
		log.Fatalf("An error occurred fetching config %v", err)
	}

	// Check comments for any commands
	if *event.Action == "created" {
		// Get Comment
		comment := event.GetComment()
		// Get Which Command is run
		// Throw away args as they are not used currently
		cmd, _ := getCommand(*comment.Body)
		// Create Client To pass through to handlers
		isClient := &issueClient{
			ctx:    ctx,
			client: client.Issues,
		}
		// Switch statement to handle different commands
		var err error
		switch {
		// Case of /cat command
		case cmd == "cat" && cfg.PullRequests.CatsEnabled:
			// Get the Cat Client
			animalClient := animals.NewCatClient()
			err = handleCats(event, isClient, animalClient)
		// Case of /dog command
		case cmd == "dog" && cfg.PullRequests.DogsEnabled:
			// Get the Dog Client
			animalClient := animals.NewDogClient()
			err = handleDogs(event, isClient, animalClient)
		default:
			break
		}
		if err != nil {
			log.Fatalf("An error occurred with the command %v: %v", cmd, err)
		}
	}
}

// getCommand strips out the command and any args that are given
func getCommand(comment string) (string, []string) {
	var args []string
	if !strings.HasPrefix(comment, "/") {
		return "", args
	}
	commands := strings.Split(comment[1:], " ")
	return commands[0], commands[1:]
}

// handleCats is the handler for the /cat command
func handleCats(
	is *github.IssueCommentEvent,
	isClient *issueClient,
	catClient *animals.Client,
) error {
	cat, err := catClient.GetLink()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("My Most Trusted Minion\n\n ![my favorite minion](%v)", cat.Url)
	catErr := createIssueComment(is, isClient, message)
	if catErr != nil {
		return catErr
	}
	return nil
}

// handleDogs is the handler for the /dog command
func handleDogs(
	is *github.IssueCommentEvent,
	isClient *issueClient,
	dogClient *animals.Client,
) error {
	dog, err := dogClient.GetLink()
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Despite how it looks it is well trained\n\n ![loyal soldier](%v)", dog.Url)
	dogErr := createIssueComment(is, isClient, message)
	if dogErr != nil {
		return dogErr
	}
	return nil
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
