package handler

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/archangel78/blockpay-backend/app/common"
	"github.com/archangel78/blockpay-backend/app/session"
	"github.com/btcsuite/btcutil/base58"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"
)

type WalletCreateResponse struct {
	PublicKey     string `json:"Publickey"`
	PrivateId     string `json:"PrivateId"`
	SolanaVersion string `json:"Version"`
}

type Transaction struct {
	TransactionId string `json:"transactionId"`
	FromAccount   string `json:"fromAccount"`
	ToAccount     string `json:"toAccount"`
	TimeStamp     string `json:"ts"`
}

type Wallet struct {
	AccountName   string `json:"accountName"`
	WalletPubKey  string `json:"walletPubKey"`
	WalletPrivKey string `json:"walletPrivKey"`
	WalletPrivId  string `json:"walletPrivId"`
}

type TransactionResponse struct {
	Transactions []Transaction `json:"transactions"`
}

func CreateTransaction(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	headers, err := common.VerifyHeaders([]string{"Transactionkey", "Iv", "Fromaccount", "Toaccount", "Amount"}, r.Header)
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}
	valid, transactionDetails, err := common.VerifyTransactionKey(db, payload.AccountName, headers["Toaccount"], headers["Amount"], headers["Transactionkey"], headers["Iv"])
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}

	if !valid {
		common.RespondError(w, 400, "Some internal error occurred CTVTK")
		return
	}

	successful, err := common.SendSol(db, payload, transactionDetails)
	if err != nil {
		fmt.Println(err)
		common.RespondError(w, 500, "Some internal error occurred")
		return
	}
	if !successful {
		common.RespondError(w, 500, "Some internal error occurred")
		return
	}
}

func GetTransactionHistory(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	result, err := db.Query("select * from Transactions where fromAccount=?", payload.AccountName)

	if err != nil {
		common.RespondError(w, 400, "Some internal error occurred GTHSEQ")
		return
	}

	transactionHistory := []Transaction{}

	for result.Next() {
		var newTransaction Transaction
		err = result.Scan(&newTransaction.TransactionId, &newTransaction.FromAccount, &newTransaction.ToAccount, &newTransaction.TransactionId)
		transactionHistory = append(transactionHistory, newTransaction)

		if err != nil {
			common.RespondError(w, 400, "Some internal error occurred GTHTHARN")
			return
		}
	}
	common.RespondJSON(w, 200, TransactionResponse{Transactions: transactionHistory})
	return
}

func CreateWallet(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	result, err := db.Query("select * from Wallet where accountName=?", payload.AccountName)

	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred CWSEQRY")
		return
	}

	for result.Next() {
		common.RespondError(w, 409, "Wallet already exists for this account")
		return
	}

	// create a RPC client
	c := client.NewClient(rpc.MainnetRPCEndpoint)

	// get the current running Solana version
	version, err := c.GetVersion(context.TODO())
	if err != nil {
		fmt.Println(err)
		common.RespondError(w, 500, "Some internal error occurred CNMNET")
		return
	}
	wallet := types.NewAccount()

	hash := md5.New()
	io.WriteString(hash, base58.Encode(wallet.PrivateKey))
	var md5HashSeed uint64 = binary.BigEndian.Uint64(hash.Sum(nil))
	var walletPrivId string = GenerateRandomPrivId(32, md5HashSeed)

	response := WalletCreateResponse{
		PublicKey:     wallet.PublicKey.ToBase58(),
		PrivateId:     walletPrivId,
		SolanaVersion: version.SolanaCore,
	}

	_, err = db.Exec("INSERT INTO Wallet (accountName, walletPubKey, walletPrivKey, walletPrivId) VALUES (?, ?, ?, ?)", payload.AccountName, wallet.PublicKey.ToBase58(), base58.Encode(wallet.PrivateKey), walletPrivId)

	if err != nil {
		fmt.Println(err)
		common.RespondError(w, 500, "Some internal error occurred ")
		return
	}
	common.RespondJSON(w, 200, response)
}

func GenerateRandomPrivId(length int, seed uint64) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(int64(seed))
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
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
