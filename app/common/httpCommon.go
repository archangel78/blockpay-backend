package common

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

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

func VerifyHeaders (expectedHeaders []string, headers http.Header) (bool, error, map[string]string) {
	output := make(map[string]string)
	for _, header := range expectedHeaders {
		value, exists := headers[header]
		if !exists{
			error := errors.New(header+" header does not exist")
			return false, error, nil
		}
		output[header] = value[0]
	}
	return true, nil, output
}

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
