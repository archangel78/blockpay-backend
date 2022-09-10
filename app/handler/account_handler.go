package handler

import (
	"net/http"

	"github.com/archangel78/blockpay-backend/app/common"
)

func CreateAccount(db string, w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	expectedParams := []string{"emailId", "accountName", "password"}
	valid, err, output := common.VerifyRequest(expectedParams, urlParams)
	if !valid {
		if err != nil {
			common.RespondError(w, 400, err.Error())
		} else {
			common.RespondError(w, 400, "")
		}
		return
	}

	common.RespondJSON(w, 200, output)
}
