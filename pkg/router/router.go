package router

import (
	"log"
	"net/http"

	paulclient "github.com/Spazzy757/paul/pkg/client"
	"github.com/Spazzy757/paul/pkg/github"
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
	installationIDHeader := r.Header.Get("X-GitHub-Hook-Installation-Target-ID")
	client, err := paulclient.GetClient(installationIDHeader)
	if err != nil {
		log.Printf("Error: %v", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = github.IncomingWebhook(r, client)
	if err != nil {
		log.Printf("Error: %v", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
