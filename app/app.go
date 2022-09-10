package app

import (
	"log"
	"net/http"

	"github.com/archangel78/blockpay-backend/app/handler"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	db string
}

func (a App) SetRoutes() {
	a.Router = mux.NewRouter()
	a.Get("/create_account", a.handleRequest(handler.CreateAccount))

	log.Fatal(http.ListenAndServe(":8080", a.Router))
}	

func (a App) Get(path string, handler http.HandlerFunc) {
	a.Router.HandleFunc(path, handler).Methods("GET")
}

func (a App) Post(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, handler).Methods("POST")
}

func (a App) Put(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, handler).Methods("PUT")
}

func (a App) Delete(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, handler).Methods("DELETE")
}

func (a App) handleRequest (handler func (db string, w http.ResponseWriter, r *http.Request)) http.HandlerFunc{
	return func (w http.ResponseWriter, r *http.Request) {
		handler(a.db, w, r)
	}
}