package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/archangel78/blockpay-backend/app/common"
	"github.com/archangel78/blockpay-backend/app/handler"
	"github.com/archangel78/blockpay-backend/app/session"
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
	cors := cors.New(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{
            http.MethodPost,
            http.MethodGet,
        },
        AllowedHeaders:   []string{"*"},
        AllowCredentials: false,
    })

	log.Fatal(http.ListenAndServe(":8080", cors.Handler(app.Router)))
}

func (app App) SetRoutes() {
	// Handle Account Endpoints
	app.Post("/create_account", app.handleRequest(handler.CreateAccount))
	app.Post("/login", app.handleRequest(handler.Login))
	app.Post("/pre_signup_verify", app.handleRequest(handler.PreSignUpVerify))
	app.Get("/check_account", app.handleAuthenticatedRequest(handler.CheckAccount))
	app.Post("/get_contacts", app.handleAuthenticatedRequest(handler.GetValidContacts))
	
	// Handle session endpoints
	app.Post("/renew_token", handler.RenewToken)
	app.Get("/test_jwt", handler.TestJwtAccessToken)

	// Handle wallet endpoints
	app.Get("/get_transaction_history", app.handleAuthenticatedRequest(handler.GetTransactionHistory))
	app.Post("/create_transaction", app.handleAuthenticatedRequest(handler.CreateTransaction))
	app.Get("/verify_send_amount", app.handleAuthenticatedRequest(handler.VerifyAmount))
	app.Get("/get_balance", app.handleAuthenticatedRequest(handler.GetBalance))
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

func (app App) handleAuthenticatedRequest (handler func (db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload)) http.HandlerFunc{
	return func (w http.ResponseWriter, r *http.Request) {
		headers, err := common.VerifyHeaders([]string{"Accesstoken"}, r.Header)
		if err != nil {
			common.RespondError(w, 400, err.Error())
			return
		} 	

		payload, valid, err := session.VerifyAccessToken(headers["Accesstoken"])
		if err != nil {
			common.RespondJSON(w, 401, map[string]string{"message": "Invalid Access Token"})
			return
		}
		if !valid {
			common.RespondJSON(w, 401, map[string]string{"message": "Invalid Access Token"})
			return
		}
		handler(app.db, w, r, *payload)
	}
}

func (app App) handleRequest (handler func (db *sql.DB, w http.ResponseWriter, r *http.Request)) http.HandlerFunc{
	return func (w http.ResponseWriter, r *http.Request) {
		handler(app.db, w, r)
	}
}