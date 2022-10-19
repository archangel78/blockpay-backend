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
	AccountName  string `json:"accountName"`
	EmailId      string `json:"emailId"`
	PhonenNo     string `json:"phoneNo"`
	CountryCode  string `json:"countryCode"`
	PasswordHash string `json:"passwordHash"`
}

func Login(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	headers := r.Header
	print(r.Header)
	password, passwordExists := headers["Password"]

	if !passwordExists {
		common.RespondError(w, 400, "Password header does not exist")
		return
	}

	var result *sql.Rows
	var err error
	aName, aNameLogin := headers["Accountname"]
	emailId, emailLogin := headers["Emailid"]
	phoneNo, phoneLogin := headers["Phoneno"]

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
	} else if phoneLogin {
		result, err = db.Query("select * from Users where phoneNumber=?", string(phoneNo[0]))

		if err != nil {
			common.RespondError(w, 400, "Some internal error occurred LISEPNO")
			return
		}
	} else {
		common.RespondError(w, 400, "Emailid or Accountname heaader does not exist")
		return
	}

	for result.Next() {
		var accountDetails AccountDetails
		err = result.Scan(&accountDetails.CountryCode, &accountDetails.PhonenNo, &accountDetails.AccountName, &accountDetails.EmailId, &accountDetails.PasswordHash)

		if err != nil {
			fmt.Println(err)
			common.RespondError(w, 400, "Some internal error occurred LIANSC")
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(accountDetails.PasswordHash), []byte(password[0]))
		if err == nil {
			jwtTokens, err := session.GenerateTokenPair(accountDetails.AccountName, accountDetails.EmailId)
			if err != nil {
				common.RespondError(w, 401, "Unauthorized")
				return
			}

			common.RespondJSON(w, 200, map[string]string{"accessToken": jwtTokens.AccessTokenSigned, "refreshToken": jwtTokens.RefreshTokenSigned, "message": "successful"})
			return
		}
		break
	}
	common.RespondError(w, 401, "Unauthorized")
}

func CreateAccount(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	headers := r.Header
	fmt.Println(r.Header, "ca")
	expectedParams := []string{"Emailid", "Accountname", "Password", "Phoneno", "Countrycode"}
	neededParams, err := common.VerifyHeaders(expectedParams, headers)
	fmt.Println(neededParams["Emailid"])
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}

	result, err := db.Query("select * from Users where accountName=? or emailId=? or phoneNumber=?", neededParams["Accountname"], neededParams["Emailid"], neededParams["Phoneno"])

	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred CASEQRY")
		return
	}

	for result.Next() {
		common.RespondError(w, 409, "AccountName Or email id or Phone number already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(neededParams["Password"]), bcrypt.DefaultCost)

	_, err = db.Exec("INSERT INTO Users (accountName, emailId, passwordHash, phoneNumber, countryCode) VALUES (?, ?, ?, ?, ?)", neededParams["Accountname"], neededParams["Emailid"], hashedPassword, neededParams["Phoneno"], neededParams["Countrycode"])

	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred CAISEXEC")
		return
	}

	common.RespondJSON(w, 200, map[string]string{"message": "successful"})
}

func RenewToken(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header["Accesstoken"]
	refreshToken := r.Header["Refreshtoken"]

	if len(accessToken) != 1 || len(refreshToken) != 1 {
		common.RespondJSON(w, 401, map[string]string{"message": "accessToken and refreshToken should be sent"})
		return
	}

	token, err := session.RenewAccessToken(accessToken[0], refreshToken[0])

	if err != nil {
		if strings.Contains(err.Error(), "early") {
			common.RespondJSON(w, 401, map[string]string{"message": err.Error()})
		} else {
			common.RespondJSON(w, 401, map[string]string{"message": "Invalid tokens"})
		}
		return
	}

	common.RespondJSON(w, 401, map[string]string{"accessToken": token.AccessTokenSigned, "message": "successful"})
}

func TestJwtAccessToken(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header["Accesstoken"]
	if len(accessToken) != 1 {
		common.RespondJSON(w, 200, map[string]string{"valid": "false"})
		return
	}
	_, valid, err := session.VerifyAccessToken(accessToken[0])
	if err != nil {
		if strings.Contains(err.Error(), "expired") {
			common.RespondJSON(w, 200, map[string]string{"valid": "expired", "message": "expired"})
			return
		}
		common.RespondJSON(w, 401, map[string]string{"valid": "false"})
		return
	}
	if valid {
		common.RespondJSON(w, 200, map[string]string{"valid": "true"})
		return
	}
	common.RespondJSON(w, 401, map[string]string{"valid": "false"})
}

func CheckAccount(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	reqHeaders, err := common.VerifyHeaders([]string{"username"}, r.Header)
	if err != nil {
		common.RespondError(w, 400, err.Error())
	}
	result, err := db.Query("select * from Users where accountName=?", reqHeaders["Accountname"])

	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred CACCSEQRY")
		return
	}
	for result.Next() {
		var accountDetails AccountDetails
		err = result.Scan(&accountDetails.CountryCode, &accountDetails.PhonenNo, &accountDetails.AccountName, &accountDetails.EmailId)
		common.RespondJSON(w, 200, map[string]string{"message": "successful", "emailid": accountDetails.EmailId, "username": accountDetails.AccountName})
		return
	}
	common.RespondError(w, 400, "Username does not exist")
}
