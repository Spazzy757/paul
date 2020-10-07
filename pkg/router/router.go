package router

import (
	"github.com/Spazzy757/paul/pkg/github"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", GithubWebHookHandler)
	return r
}

func GithubWebHookHandler(w http.ResponseWriter, r *http.Request) {
	jsonByte, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	github.IncomingWebhook(jsonByte, r)
	w.WriteHeader(http.StatusOK)
}
