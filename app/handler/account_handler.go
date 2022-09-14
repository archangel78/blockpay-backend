package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/archangel78/blockpay-backend/app/common"
	"github.com/archangel78/blockpay-backend/app/session"
)

type AccountDetails struct {
	AccountName string	`json:"accountName"`
	EmailId string		`json:"emailId"`
	PasswordHash string	`json:"passwordHash"`
}

func TestJwtAccessToken(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header["Accesstoken"]
	_, valid, err := session.VerifyAccessToken(accessToken[0])

	if err != nil {
		fmt.Println(err)
		if strings.Contains(err.Error(), "expired"){
			common.RespondJSON(w, 200, map[string]string{"valid": "expired"})	
			return
		}
		common.RespondJSON(w, 401, map[string]string{"valid": "false"})
		return
	}
	if valid {
		common.RespondJSON(w, 200, map[string]string{"valid": "true"})
	}
	common.RespondJSON(w, 401, map[string]string{"valid": "false"})
}

func RenewToken(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header["Accesstoken"]
	_, valid, err := session.VerifyAccessToken(accessToken[0])

	if err != nil {
		fmt.Println(err)
		common.RespondJSON(w, 401, map[string]string{"valid": "false"})
	}
	if valid {
		common.RespondJSON(w, 200, map[string]string{"valid": "true"})
	}
	common.RespondJSON(w, 401, map[string]string{"valid": "false"})
}

func Login(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	headers := r.Header
	password, passwordExists := headers["Password"]
	
	if !passwordExists {
		common.RespondError(w, 400, "Password header does not exist")
		return
	}

	var result *sql.Rows
	var err error
	aName, aNameLogin := headers["Accountname"]
	emailId, emailLogin := headers["Emailid"]

	if aNameLogin {
		result, err = db.Query("select * from Users where accountName=?", aName[0])
		
		if err != nil {
			common.RespondError(w, 400, "Some internal error occurred LISEAN")
			return
		}
	} else if emailLogin {
		result, err = db.Query("select * from Users where emailId=?", string(emailId[0]))

		if err != nil {
			common.RespondError(w, 400, "Some internal error occurred LISEEID")
			return
		}
	} else {
		common.RespondError(w, 400, "Emailid or Accountname heaader does not exist")
		return
	}

	for result.Next() {
		var accountDetails AccountDetails
		err = result.Scan(&accountDetails.AccountName, &accountDetails.EmailId, &accountDetails.PasswordHash)

		if err != nil {
			common.RespondError(w, 400, "Some internal error occurred LIANSC")
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(accountDetails.PasswordHash), []byte(password[0]))
		if err == nil {
			jwtTokens, err := session.GenerateTokenPair(accountDetails.AccountName, accountDetails.EmailId)

			if err != nil {
				fmt.Println(err)
				common.RespondError(w, 401, "Unauthorized")
				return
			}

			common.RespondJSON(w, 200, map[string]string{"accessToken": jwtTokens.AccessTokenSigned, "refreshToken": jwtTokens.RefreshTokenSigned})
			return
		}
		break
	}	
	common.RespondError(w, 401, "Unauthorized")
}

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