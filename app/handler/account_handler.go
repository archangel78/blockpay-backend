package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/mail"
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

			result2, err := db.Query("select walletPrivId, walletPubKey from Wallet where accountName=?", accountDetails.AccountName)
			if err != nil {
				fmt.Println(err)
				common.RespondError(w, 500, "Some internal error occurred")
				return
			}
			for result2.Next() {
				var walletPrivId string
				var walletPubKey string
				err = result2.Scan(&walletPrivId, &walletPubKey)

				if err != nil {
					fmt.Println(err)
					common.RespondError(w, 500, "Some internal error occurred")
					return
				}
				common.RespondJSON(w, 200, map[string]string{"accessToken": jwtTokens.AccessTokenSigned, "refreshToken": jwtTokens.RefreshTokenSigned, "message": "successful", "walletPrivId": walletPrivId, "walletPubKey": walletPubKey, "accountName": accountDetails.AccountName})
				return
			}
			if err != nil {
				fmt.Println(err)
				common.RespondError(w, 500, "Some internal error occurred")
				return
			}
		}
		break
	}
	common.RespondError(w, 401, "Unauthorized")
}

func CreateAccount(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	headers := r.Header
	expectedParams := []string{"Emailid", "Accountname", "Password", "Phoneno", "Countrycode", "Name"}
	neededParams, err := common.VerifyHeaders(expectedParams, headers)
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
		common.RespondError(w, 500, "Some internal error occurred CAISUEXEC")
		return
	}

	_, err = db.Exec("INSERT INTO OtherDetails (accountName, fullName) VALUES (?, ?)", neededParams["Accountname"], neededParams["Name"])
	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred CAISODEXEC")
		return
	}

	walletRes, err := common.CreateWallet(db, neededParams["Accountname"])
	if err != nil {
		common.RespondError(w, 500, "Some internal eroor occurred CAWCRT")
		return
	}

	jwtTokens, err := session.GenerateTokenPair(neededParams["Accountname"], neededParams["Emailid"])
	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred CAJWTGEN")
		return
	}

	common.RespondJSON(w, 200, map[string]string{"accessToken": jwtTokens.AccessTokenSigned, "refreshToken": jwtTokens.RefreshTokenSigned, "message": "successful", "walletPrivId": walletRes.PrivateId, "walletAddress": walletRes.PublicKey, "accountName": neededParams["Accountname"]})
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

func PreSignUpVerify(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	headers := r.Header
	expectedParams := []string{"Emailid", "Accountname", "Password"}
	neededParams, err := common.VerifyHeaders(expectedParams, headers)
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}

	result, err := db.Query("select * from Users where accountName=?", neededParams["Accountname"])

	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred PSVRFSEQRY")
		return
	}

	for result.Next() {
		common.RespondError(w, 409, "Username already exists")
		return
	}

	result, err = db.Query("select * from Users where emailId=?", neededParams["Emailid"])

	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred PSVRFSEQRY")
		return
	}

	for result.Next() {
		common.RespondError(w, 409, "Email Id is already in use")
		return
	}

	_, err = mail.ParseAddress(neededParams["Emailid"])
	if err != nil {
		common.RespondError(w, 400, "Invalid Email Address")
		return
	}

	if len(neededParams["Password"]) < 6 {
		common.RespondError(w, 400, "Password should be atleast of length 6")
		return
	}

	if len(neededParams["Password"]) > 15 {
		common.RespondError(w, 400, "Password should be atmost of length 15")
		return
	}

	common.RespondJSON(w, 200, map[string]string{"message": "successful"})
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
	reqHeaders, err := common.VerifyHeaders([]string{"Username"}, r.Header)
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}
	result, err := db.Query("select * from Users where accountName=?", reqHeaders["Username"])

	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred CACCSEQRY")
		return
	}

	for result.Next() {
		var accountDetails AccountDetails
		err = result.Scan(&accountDetails.CountryCode, &accountDetails.PhonenNo, &accountDetails.AccountName, &accountDetails.EmailId, &accountDetails.PasswordHash)
		if err != nil {
			print(err.Error())
			common.RespondError(w, 500, "Some internal error occurred CACCSCN")
			return
		}
		if accountDetails.AccountName != payload.AccountName {
			common.RespondJSON(w, 200, map[string]string{"message": "successful", "emailid": accountDetails.EmailId, "username": accountDetails.AccountName})
			return
		} else {
			common.RespondError(w, 400, "Can't send money to yourself")
			return
		}
	}
	common.RespondError(w, 400, "Username does not exist")
}
