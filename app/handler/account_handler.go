package handler

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/archangel78/blockpay-backend/app/common"
)

func CreateAccount(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	expectedParams := []string{"emailId", "accountName", "password"}
	valid, err, neededParams := common.VerifyRequest(expectedParams, urlParams)
	if !valid {
		if err != nil {
			common.RespondError(w, 400, err.Error())
		} else {
			common.RespondError(w, 400, "")
		}
		return
	}

	result, err := db.Query("select * from Users where accountName=? or emailId=?", neededParams["accountName"], neededParams["emailId"])

	if err != nil {
		fmt.Println("CreateAccount Select query error: ", err)
		common.RespondError(w, 500, "Some internal error occurred CASEQRY")
		return
	}

	for result.Next() {
		common.RespondError(w, 409, "AccountName Or email id already exists")
		return
	}

	_, err = db.Exec("INSERT INTO Users (accountName, emailId, passwordHash) VALUES (?, ?, ?)", neededParams["accountName"], neededParams["emailId"], neededParams["password"])

	if err != nil {
		fmt.Println("CreateAccount insert exec error: ", err)
		common.RespondError(w, 500, "Some internal error occurred CAISEXEC")
		return
	}

	common.RespondJSON(w, 200, map[string]string{"successful": "ok"})
}