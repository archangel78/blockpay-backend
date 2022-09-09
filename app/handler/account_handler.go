package handler

import (
	"fmt"
	"net/http"
)

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	if emailId, exists := urlParams["emailId"]; exists {
		if password, exists := urlParams["password"]; exists{
			fmt.Println(emailId, password)
		}
	}
}
 