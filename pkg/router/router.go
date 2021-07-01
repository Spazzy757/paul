package router

import (
	"context"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	paulclient "github.com/Spazzy757/paul/pkg/client"
	paulgithub "github.com/Spazzy757/paul/pkg/github"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/google/go-github/v36/github"
	"github.com/gorilla/mux"
)

//GetRouter .
func GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/webhooks", GithubWebHookHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist")))
	return r
}

//GithubWebHookHandler .
func GithubWebHookHandler(w http.ResponseWriter, r *http.Request) {
	// handle authentication
	secretKey := helpers.GetEnv("SECRET_KEY", "")
	payload, validationErr := github.ValidatePayload(r, []byte(secretKey))
	if handleError(w, validationErr) {
		return
	}
	instllationID, err := getInstallationId(payload)
	if handleError(w, err) {
		return
	}
	gClient, err := paulclient.GetInstallationClient(instllationID)
	ctx := context.Background()
	if handleError(w, err) {
		return
	}
	err = paulgithub.IncomingWebhook(ctx, r, payload, gClient)
	if handleError(w, err) {
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleError(w http.ResponseWriter, err error) bool {
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("webhook error occurred")
		w.WriteHeader(http.StatusBadRequest)
		return true
	}
	return false
}

// Payload is used to get the installation ID from payload
type Payload struct {
	Installation Installation `json:"installation"`
}

// Installation is used to get the installation ID from the payload
type Installation struct {
	ID int64 `json:"id"`
}

func getInstallationId(payload []byte) (int64, error) {
	p := &Payload{}
	err := json.Unmarshal(payload, p)
	if err != nil {
		return 0, err
	}
	return p.Installation.ID, err
}
