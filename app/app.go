package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/archangel78/blockpay-backend/app/common"
	"github.com/archangel78/blockpay-backend/app/handler"
	config "github.com/archangel78/blockpay-backend/mysql-config"
)

type App struct {
	Router *mux.Router
	db *sql.DB
}

func (app App) Initialize(dbConfig *config.DbConfig) {
	app.Router = mux.NewRouter()
	db, err := common.OpenDbConnection(dbConfig)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer db.Close()
	app.db = db

	app.SetRoutes()
	log.Fatal(http.ListenAndServe(":8080", app.Router))
}

func (app App) SetRoutes() {
	app.Post("/create_account", app.handleRequest(handler.CreateAccount))
	app.Post("/login", app.handleRequest(handler.Login))
	app.Get("/test_jwt", handler.TestJwtAccessToken)
	app.Get("/renew_token", handler.RenewToken)
}	

func (app App) Get(path string, handler http.HandlerFunc) {
	app.Router.HandleFunc(path, handler).Methods("GET")
}

func (app App) Post(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	app.Router.HandleFunc(path, handler).Methods("POST")
}

func (app App) Put(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	app.Router.HandleFunc(path, handler).Methods("PUT")
}

func (app App) Delete(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	app.Router.HandleFunc(path, handler).Methods("DELETE")
}

func (a App) handleRequest (handler func (db *sql.DB, w http.ResponseWriter, r *http.Request)) http.HandlerFunc{
	return func (w http.ResponseWriter, r *http.Request) {
		handler(a.db, w, r)
	}
}