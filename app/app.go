package app

import (
	"log"
	"net/http"

	"github.com/archangel78/blockpay-backend/app/handler"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
}

func (a App) SetRoutes() {
	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/create_account", handler.CreateAccount)
	log.Fatal(http.ListenAndServe(":8080", a.Router))
}	
