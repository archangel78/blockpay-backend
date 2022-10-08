package common

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/archangel78/blockpay-backend/app/session"
)

type WalletDetails struct {
	accountName   string
	walletPubKey  string
	walletPrivKey string
	walletPrivId  string
}

type TransactionDetails struct {
	FromAccount string `json:"fromAccount"`
	ToAccount   string `json:"toAccount"`
	Amount      string `json:"amount"`
	Prover      string `json:"prover"`
	ExpiryTime  string `json:"expiryTime"`
}

func VerifyTransactionKey(db *sql.DB, fromAccount string, transactionKey string, ivStr string) (bool, error) {
	result, err := db.Query("select walletPrivId from Wallet where accountName=?", string(fromAccount))
	for result.Next() {
		var walletDetails WalletDetails
		err = result.Scan(&walletDetails.walletPrivId)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		transactionDetails, err := decrypt([]byte(walletDetails.walletPrivId), transactionKey, ivStr)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		fmt.Println(transactionDetails)
		break
	}
	return true, nil
}

func decrypt(key []byte, cryptoText string, ivStr string) (*TransactionDetails, error) {
	ciphertext, _ := hex.DecodeString(cryptoText)
	iv, _ := hex.DecodeString(ivStr)

	var block cipher.Block
	var err error
	if block, err = aes.NewCipher(key); err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(ciphertext, ciphertext)

	var plaintext = strings.Replace(string(ciphertext), "'", "\"", -1)
	plaintext = strings.TrimSpace(plaintext)
	plaintext = strings.Replace(plaintext, "\b", "", -1)
	var transactionDetails TransactionDetails
	fmt.Println(plaintext)

	err = json.Unmarshal([]byte(strings.Replace(plaintext, "'", "\"", -1)), &transactionDetails)
	
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("Invalid transactionKey")
	}
	return &transactionDetails, nil
}

func SendSol(db *sql.DB, payload session.Payload, toAccountName string) (bool, error) {
	result, err := db.Query("select walletPubKey from Wallet where accountName=?", string(toAccountName))

	for result.Next() {
		var walletDetails WalletDetails
		err = result.Scan(&walletDetails.walletPubKey)

		if err != nil {
			fmt.Println(err)
			return false, err
		}
		fmt.Println(walletDetails)
		break
	}
	return true, nil
}
