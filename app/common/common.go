package common

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

func VerifyRequest(expectedParams []string, urlParams url.Values) (bool, error, map[string]string) {
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
	RespondJSON(w, code, map[string]string{"error": message})
}
