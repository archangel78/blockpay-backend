package common

import (
	"os"
	"fmt"
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"errors"
	"net/http"
	"net/url"
)

func SendNotification (db *sql.DB, accountName string, message string) {
	result, err := db.Query("select deviceToken from OtherDetails where accountName=?", accountName)
	if err != nil {
		return
	}
	for result.Next() {
		var deviceToken string
		result.Scan(&deviceToken)

		var pushyApiKey = os.Getenv("PUSHY_API_KEY")
		
		var postBody = `{
			"to": "`+deviceToken+`",
			"data": {
				"message": "`+message+`"
			}
		}`
		
		if err != nil {
			fmt.Println(err)
			return
		}		

		responseBody := bytes.NewBuffer([]byte(postBody))
		resp, err := http.Post("https://api.pushy.me/push?api_key="+pushyApiKey, "application/json", responseBody)

		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		sb := string(body)
		fmt.Println(sb)
		return			 
	}
}

func VerifyUrlParams (expectedParams []string, urlParams url.Values) (bool, error, map[string]string) {
	output := make(map[string]string)
	for _, param := range expectedParams {
		value, exists := urlParams[param]
		if !exists{
			error := errors.New(param+" parameter does not exist")
			return false, error, nil
		}
		output[param] = value[0]
	}
	return true, nil, output
}

func VerifyHeaders (expectedHeaders []string, headers http.Header) (map[string]string, error) {
	output := make(map[string]string)
	for _, header := range expectedHeaders {
		value, exists := headers[header]
		if !exists{
			error := errors.New(header+" header does not exist")
			return nil, error
		}
		output[header] = value[0]
	}
	return output, nil
}

// func VerifyOptionalHeaders (optionalHeaders []string, headers http.Header) (error)
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(response))
}

func RespondError(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, map[string]string{"message": message})
}
