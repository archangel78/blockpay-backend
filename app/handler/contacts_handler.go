package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/archangel78/blockpay-backend/app/common"
	"github.com/archangel78/blockpay-backend/app/session"
)

func GetValidContacts(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	var contacts []string
	err = json.Unmarshal(b, &contacts)
	if len(contacts) > 3500 {
		fmt.Println("too many contacts")
		common.RespondError(w, 500, "Too many contacts to process")
		return
	}

	var mappedContacts = make(map[int]string)
	var mappedIndex = make(map[string]int)
	for index, element := range contacts {
		number := strings.Replace(element, " ", "", -1)
		if len(number) < 8 || len(number) > 15 {
			continue
		}
		if number[:1] == "+" {
			if number[:3] != "+91" {
				continue
			} else {
				number = number[3:]
			}
		}
		numericRegex := regexp.MustCompile(`[0-9]+`)
		if !numericRegex.MatchString(number) {
			continue
		}
		mappedContacts[index] = number
		mappedIndex[number] = index
	}

	result, err := db.Query("select phoneNumber from Users order by accountName")

	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred GVCSE")
		return
	}
	var newIndices = []int{}
	for result.Next() {
		var phoneNumber string
		result.Scan(&phoneNumber)
		if _, exists := mappedIndex[phoneNumber]; exists {
			newIndices = append(newIndices, mappedIndex[phoneNumber])
		}
	}
	for i := 0; i < len(contacts); i++ {
		if !contains(newIndices, i) {
			newIndices = append(newIndices, i)
		}
	}
	jsonIndices, err := json.Marshal(newIndices)
	if err != nil {
		fmt.Println(err)
		common.RespondError(w, 500, "Some internal error occurred")
	}
	common.RespondJSON(w, 200, map[string]string{"message": "successful", "indices": string(jsonIndices)})
}

func VerifyPhoneNumber(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	headers, err := common.VerifyHeaders([]string{"Phoneno"}, r.Header)
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}
	phoneNumber := headers["Phoneno"]

	if len(phoneNumber) < 8 {
		common.RespondError(w, 409, "Phone number doesn't exist")
		return
	}

	if phoneNumber[:1] == "+" {
		if phoneNumber[:3] == "+91" {
			phoneNumber = phoneNumber[3:]
		} else {
			common.RespondError(w, 409, "Phone number doesn't exist")
			return
		}
	}
	phoneNumber = strings.Replace(phoneNumber, " ", "", -1)

	result, err := db.Query("select accountName from Users where phoneNumber=?", phoneNumber)

	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred VPNSE1")
		return
	}
	for result.Next() {
		var accountName string
		err = result.Scan(&accountName)
		if err != nil {
			common.RespondError(w, 500, "Some internal error occurred VPNSCN1")
			return
		}

		result1, err := db.Query("select fullName from OtherDetails where accountName=?", accountName)
		if err != nil {
			common.RespondError(w, 500, "Some internal error occurred VPNSE1")
			return
		}

		for result1.Next() {
			var fullName string
			err = result1.Scan(&fullName)
			if err != nil {
				fmt.Println(err)
				common.RespondError(w, 500, "Some internal error occurred VPNSCN1")
				return
			}
			if payload.AccountName != accountName {
				common.RespondJSON(w, 200, map[string]string{"message": "successful", "fullName": fullName, "accountName": accountName})
			} else {
				common.RespondError(w, 400, "Can't send money to yourself")
			}
			return
		}
		break
	}
	common.RespondError(w, 409, "Phone number doesn't exist")
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
