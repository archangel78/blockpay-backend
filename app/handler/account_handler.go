package handler

import (
	"fmt"
	"net/http"
	"database/sql"
	"golang.org/x/crypto/bcrypt"

	"github.com/archangel78/blockpay-backend/app/common"
)

func CreateAccount(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	headers := r.Header
	expectedParams := []string{"Emailid", "Accountname", "Password"}
	valid, err, neededParams := common.VerifyHeaders(expectedParams, headers)
	if !valid {
		if err != nil {
			common.RespondError(w, 400, err.Error())
		} else {
			common.RespondError(w, 400, "")
		}
		return
	}

	result, err := db.Query("select * from Users where accountName=? or emailId=?", neededParams["Accountname"], neededParams["Emailid"])

	if err != nil {
		fmt.Println("CreateAccount Select query error: ", err)
		common.RespondError(w, 500, "Some internal error occurred CASEQRY")
		return
	}

	for result.Next() {
		common.RespondError(w, 409, "AccountName Or email id already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(neededParams["Password"]), bcrypt.DefaultCost)

	_, err = db.Exec("INSERT INTO Users (accountName, emailId, passwordHash) VALUES (?, ?, ?)", neededParams["Accountname"], neededParams["Emailid"], hashedPassword)

	if err != nil {
		fmt.Println("CreateAccount insert exec error: ", err)
		common.RespondError(w, 500, "Some internal error occurred CAISEXEC")
		return
	}

	common.RespondJSON(w, 200, map[string]string{"successful": "ok"})
}