package router

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	paulclient "github.com/Spazzy757/paul/pkg/client"
	paulgithub "github.com/Spazzy757/paul/pkg/github"
	"github.com/Spazzy757/paul/pkg/helpers"
	"github.com/google/go-github/v32/github"
	"github.com/gorilla/mux"
)

//GetRouter .
func GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", GithubWebHookHandler)
	return r
}

//GithubWebHookHandler .
func GithubWebHookHandler(w http.ResponseWriter, r *http.Request) {
	// handle authentication
	secret_key := helpers.GetEnv("SECRET_KEY", "")
	payload, validationErr := github.ValidatePayload(r, []byte(secret_key))
	if handleError(w, validationErr) {
		return
	}
	instllationID, err := getInstallationId(payload)
	if handleError(w, err) {
		return
	}
	gClient, err := paulclient.GetClient(instllationID)
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
		log.Printf("Error: %v", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return true
	}
	return false
}

type Payload struct {
	Installation Installation `json:"installation"`
}

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
