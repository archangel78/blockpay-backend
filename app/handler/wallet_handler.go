package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/archangel78/blockpay-backend/app/common"
	"github.com/archangel78/blockpay-backend/app/session"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
)

type WalletCreateResponse struct {
	PublicKey     string `json:"Publickey"`
	PrivateId     string `json:"PrivateId"`
	SolanaVersion string `json:"Version"`
}

type Transaction struct {
	TransactionId     string `json:"transactionId"`
	FromAccount       string `json:"fromAccount"`
	ToAccount         string `json:"toAccount"`
	ToWallet          string `json:"toWallet"`
	TransactionAmount string `json:"transactionAmount"`
	ToName            string `json:"name"`
	FromName          string `json:"fromName"`
	TimeStamp         string `json:"ts"`
}

type Wallet struct {
	AccountName   string `json:"accountName"`
	WalletPubKey  string `json:"walletPubKey"`
	WalletPrivKey string `json:"walletPrivKey"`
	WalletPrivId  string `json:"walletPrivId"`
}

type TransactionResponse struct {
	Transactions []Transaction `json:"transactions"`
	Message      string        `json:"message"`
}

func CreateTransaction(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	headers, err := common.VerifyHeaders([]string{"Transactionkey", "Iv", "Fromaccount", "Toaccount", "Amount"}, r.Header)
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}
	_, transactionDetails, err := common.VerifyTransactionKey(db, payload.AccountName, headers["Toaccount"], headers["Amount"], headers["Transactionkey"], headers["Iv"])
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}

	// Uncomment to skip transaction verification
	// lamportAmount, err := strconv.ParseFloat(headers["Amount"], 64)
	// if err != nil {
	// 	fmt.Println(err)
	// 	common.RespondError(w, 500, "Some internal error occurred")
	// 	return
	// }
	// transactionDetails := &common.TransactionDetails{FromAccount: headers["Fromaccount"], ToAccount: headers["Toaccount"], LamportAmount: lamportAmount}

	_, signature, err := common.SendSol(db, payload, transactionDetails)
	if err != nil {
		fmt.Println(err)
		common.RespondError(w, 500, "Some internal error occurred")
		return
	}

	result, err := db.Query("select fullName from OtherDetails where accountName=?", headers["Toaccount"])
	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred CTSEFNAM")
		return
	}

	var toName string
	var fromName string
	for result.Next() {
		result.Scan(&toName)
		break
	}

	result, err = db.Query("select fullName from OtherDetails where accountName=?", headers["Fromaccount"])
	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred CTSEFNAM")
		return
	}
	for result.Next() {
		result.Scan(&fromName)
		break
	}

	err = common.WriteTransaction(db, signature[:25], headers["Fromaccount"], headers["Toaccount"], "NA", headers["Amount"], toName, fromName)

	if err != nil {
		fmt.Println(err)
		common.RespondError(w, 500, "Some internal error occurred")
		return
	}
	common.RespondJSON(w, 200, map[string]string{"message": "successful", "transactionId": signature[:25], "name": toName})
	common.SendNotification(db, headers["Toaccount"], "Received "+headers["Amount"]+" SOL:Transaction received from "+toName+":Payment Received:"+toName+":"+headers["Amount"]+" SOL:"+"Account Id:"+headers["Toaccount"]+":"+signature[:25])
}

func CreateOfflineTransaction(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	headers, err := common.VerifyHeaders([]string{"Transactionkey", "Iv", "Fromaccount"}, r.Header)
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}
	fmt.Println(headers);
	_, transactionDetails, err := common.VerifyOfflineTransactionKey(db,  headers["Fromaccount"], payload.AccountName, headers["Transactionkey"], headers["Iv"])
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}

	// Uncomment to skip transaction verification
	// lamportAmount, err := strconv.ParseFloat(headers["Amount"], 64)
	// if err != nil {
	// 	fmt.Println(err)
	// 	common.RespondError(w, 500, "Some internal error occurred")
	// 	return
	// }
	// transactionDetails := &common.TransactionDetails{FromAccount: headers["Fromaccount"], ToAccount: headers["Toaccount"], LamportAmount: lamportAmount}

	_, signature, err := common.SendSol(db, payload, transactionDetails)
	if err != nil {
		fmt.Println(err)
		common.RespondError(w, 500, "Some internal error occurred")
		return
	}

	result, err := db.Query("select fullName from OtherDetails where accountName=?", headers["Toaccount"])
	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred CTSEFNAM")
		return
	}

	var toName string
	var fromName string
	for result.Next() {
		result.Scan(&toName)
		break
	}

	result, err = db.Query("select fullName from OtherDetails where accountName=?", headers["Fromaccount"])
	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred CTSEFNAM")
		return
	}
	for result.Next() {
		result.Scan(&fromName)
		break
	}

	err = common.WriteTransaction(db, signature[:25], headers["Fromaccount"], headers["Toaccount"], "NA", headers["Amount"], toName, fromName)

	if err != nil {
		fmt.Println(err)
		common.RespondError(w, 500, "Some internal error occurred")
		return
	}
	common.RespondJSON(w, 200, map[string]string{"message": "successful", "transactionId": signature[:25], "name": toName})
	common.SendNotification(db, headers["Toaccount"], "Received "+headers["Amount"]+" SOL:Transaction received from "+toName+":Payment Received:"+toName+":"+headers["Amount"]+" SOL:"+"Account Id:"+headers["Toaccount"]+":"+signature[:25])
}


func GetTransactionHistory(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	result, err := db.Query("select transactionId, fromAccount, toAccount, toWallet, transactionAmount, ts, toName, fromName from Transactions where fromAccount=? or toAccount=?", payload.AccountName, payload.AccountName)

	if err != nil {
		fmt.Println(err)
		common.RespondError(w, 400, "Some internal error occurred GTHSEQ")
		return
	}

	transactionHistory := []Transaction{}

	for result.Next() {
		var newTransaction Transaction
		err = result.Scan(&newTransaction.TransactionId, &newTransaction.FromAccount, &newTransaction.ToAccount, &newTransaction.ToWallet, &newTransaction.TransactionAmount, &newTransaction.TimeStamp, &newTransaction.ToName, &newTransaction.FromName)
		transactionHistory = append(transactionHistory, newTransaction)

		if err != nil {
			fmt.Println(err)
			common.RespondError(w, 400, "Some internal error occurred GTHTHARN")
			return
		}
	}
	reverseTransactionHistory := []Transaction{}
	for i := len(transactionHistory) - 1; i >= 0; i-- {
		reverseTransactionHistory = append(reverseTransactionHistory, transactionHistory[i])
	}
	common.RespondJSON(w, 200, TransactionResponse{Transactions: reverseTransactionHistory, Message: "successful"})
	return
}

func GetBalance(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	result, err := db.Query("select walletPubKey from Wallet where accountName=?", payload.AccountName)

	if err != nil {
		common.RespondError(w, 400, "Some internal error occurred GBSEQRY")
		return
	}

	for result.Next() {
		var wallet Wallet
		err = result.Scan(&wallet.WalletPubKey)
		if err != nil {
			fmt.Print(err)
			common.RespondError(w, 500, "Some internal error occurred GBSCN")
			return
		}

		c := client.NewClient(rpc.DevnetRPCEndpoint)
		balance, err := c.GetBalance(
			context.Background(),
			wallet.WalletPubKey,
		)

		_, err = c.RequestAirdrop(
			context.TODO(), 
			wallet.WalletPubKey, 
			100e9,)
		fmt.Println(err)

		if err != nil {
			fmt.Print(err)
			common.RespondError(w, 500, "Some internal error occurred GBGBERR")
			return
		}
		lampBalance := float32(balance)
		solBalance := fmt.Sprintf("%v", lampBalance*0.000000001)
		common.RespondJSON(w, 200, map[string]string{"message": "successful", "balance": solBalance})
		return
	}
	common.RespondError(w, 400, "Invalid Account")
}

func VerifyAmount(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	verifyAmount, err := common.VerifyHeaders([]string{"Amount"}, r.Header)
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}

	result, err := db.Query("select walletPubKey from Wallet where accountName=?", payload.AccountName)

	if err != nil {
		common.RespondError(w, 400, "Some internal error occurred VASEQRY")
		return
	}

	for result.Next() {
		var wallet Wallet
		err = result.Scan(&wallet.WalletPubKey)

		if err != nil {
			fmt.Print(err)
			common.RespondError(w, 500, "Some internal error occurred VASCN")
			return
		}
		// TODO: change to mainnet after production
		c := client.NewClient(rpc.DevnetRPCEndpoint)
		balance, err := c.GetBalance(
			context.Background(),
			wallet.WalletPubKey,
		)
		if err != nil {
			fmt.Print(err)
			common.RespondError(w, 500, "Some internal error occurred VAGBLNC")
			return
		}
		amountFloat, err := strconv.ParseFloat(verifyAmount["Amount"], 32)
		if err != nil {
			fmt.Print(err)
			common.RespondError(w, 400, "Invalid amount sent")
			return
		}

		lamportAmount := amountFloat / 0.000000001

		if int(lamportAmount) <= int(balance) && int(lamportAmount) > 0 {
			common.RespondJSON(w, 200, map[string]string{"message": "successful"})
			return
		} else {
			solBalance := float32(balance) * 0.000000001
			common.RespondJSON(w, 200, map[string]string{"message": "Amount greater than your balance: " + fmt.Sprint(solBalance)})
			return
		}
	}
	common.RespondError(w, 500, "invalid")
}
