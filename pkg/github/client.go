package github

import (
	"context"
	"github.com/Spazzy757/paul/pkg/config"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"log"
)

func getClient(installationId int64) (*github.Client, context.Context) {
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
	return client, ctx
}
