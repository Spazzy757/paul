package router

import (
	"net/http"

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
	err := github.IncomingWebhook(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
