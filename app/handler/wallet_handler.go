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

	"github.com/archangel78/blockpay-backend/app/common"
	"github.com/archangel78/blockpay-backend/app/session"
	"github.com/btcsuite/btcutil/base58"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"
)

type WalletCreateResponse struct {
	PublicKey     string `json:"Publickey"`
	PrivateId    string `json:"PrivateId"`
	SolanaVersion string `json:"Version"`
}

type Transaction struct {
	TransactionId string `json:"transactionId"`
	FromAccount   string `json:"fromAccount"`
	ToAccount     string `json:"toAccount"`
	TimeStamp     string `json:"ts"`
}

type TransactionResponse struct {
	Transactions []Transaction `json:"transactions"`
}

func CreateTransaction(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	headers, err := common.VerifyHeaders([]string{"Transactionkey", "Iv"}, r.Header)
	if err != nil {
		common.RespondError(w, 400, err.Error())
		return
	}
	valid, err := common.VerifyTransactionKey(db, payload.AccountName, headers["Transactionkey"], headers["Iv"])
	if err != nil {
		common.RespondError(w, 400, "Invalid transaction key")
		return
	}
	fmt.Println(valid)
	successful, err := common.SendSol(db, payload, headers["Transactionkey"])
	if !successful {
		fmt.Println(err)
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
