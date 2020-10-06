package router

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/payload", GithubHandler)
	return r
}

func GithubHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v", r.Body)
	w.WriteHeader(http.StatusOK)
}
