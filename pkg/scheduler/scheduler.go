package scheduler

import (
	"context"
	paulclient "github.com/Spazzy757/paul/pkg/client"
	paulgithub "github.com/Spazzy757/paul/pkg/github"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/google/go-github/v35/github"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

// AddSchedule Runs scheduled tasks for Paul
func AddSchedule(c *cron.Cron) {
	stalePullRequestsSchedule := helpers.GetEnv("STALE_CHECK_SCHEDULE", "0 * * * *")
	_, _ = c.AddFunc(stalePullRequestsSchedule, func() {
		gClient, _ := paulclient.GetClient()
		ctx := context.Background()
		installations, _, err := gClient.Apps.ListInstallations(ctx, &github.ListOptions{})
		if handleErr(err) {
			return
		}
		for _, installation := range installations {
			gInstallationClient, err := paulclient.GetInstallationClient(installation.GetID())
			if handleErr(err) {
				continue
			}
			paulgithub.PullRequestsScheduledJobs(ctx, gInstallationClient)
		}
	})
}

func handleErr(err error) bool {
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("webhook error occurred")
		return true
	}
	return false
}
