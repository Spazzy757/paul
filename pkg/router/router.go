package router

import (
	"github.com/Spazzy757/paul/pkg/github"
	"github.com/gorilla/mux"
	"net/http"
)

func GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", GithubWebHookHandler)
	return r
}

func GithubWebHookHandler(w http.ResponseWriter, r *http.Request) {
	err := github.IncomingWebhook(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
