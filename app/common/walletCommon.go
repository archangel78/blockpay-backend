package common

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/archangel78/blockpay-backend/app/session"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/program/system"
	"github.com/portto/solana-go-sdk/rpc"
	"github.com/portto/solana-go-sdk/types"
)

type WalletDetails struct {
	accountName   string
	walletPubKey  string
	walletPrivKey string
	walletPrivId  string
}

type TransactionDetails struct {
	FromAccount   string  `json:"fromAccount"`
	ToAccount     string  `json:"toAccount"`
	Amount        string  `json:"amount"`
	LamportAmount float64 `json:"lamportAmount"`
	Prover        string  `json:"prover"`
	ExpiryTime    string  `json:"expiryTime"`
}

func VerifyTransactionKey(db *sql.DB, fromAccount string, toAccount string, sendAmount string, transactionKey string, ivStr string) (bool, *TransactionDetails, error) {
	result, err := db.Query("select walletPrivId from Wallet where accountName=?", string(fromAccount))
	for result.Next() {
		var walletDetails WalletDetails
		err = result.Scan(&walletDetails.walletPrivId)
		if err != nil {
			fmt.Println(err)
			return false, nil, err
		}
		transactionDetails, err := decrypt([]byte(walletDetails.walletPrivId), transactionKey, ivStr)
		if err != nil {
			fmt.Println(err)
			return false, nil, err
		}

		if fromAccount != transactionDetails.FromAccount {
			return false, nil, errors.New("Invalid From Account")
		}

		if toAccount != transactionDetails.ToAccount {
			return false, nil, errors.New("Invalid To Account")
		}

		if sendAmount != transactionDetails.Amount {
			return false, nil, errors.New("Invalid transaction Amount")
		}

		lamportAmount, err := strconv.ParseFloat(transactionDetails.Amount, 32)
		if err != nil {
			return false, nil, err
		}
		transactionDetails.LamportAmount = lamportAmount

		return true, transactionDetails, nil
	}
	return false, nil, errors.New("Some unknown error")
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

func SendSol(db *sql.DB, payload session.Payload, transactionDetails *TransactionDetails) (bool, error) {
	result, err := db.Query("select walletPubKey, walletPrivKey from Wallet where accountName=?", transactionDetails.ToAccount)
	if err != nil {
		return false, err
	}
	for result.Next() {
		var towalletDetails WalletDetails
		err = result.Scan(&towalletDetails.walletPubKey, &towalletDetails.walletPrivKey)
		if err != nil {
			fmt.Println(err)
			return false, err
		}

		privResult, err := db.Query("select walletPrivKey from Wallet where accountName=?", transactionDetails.FromAccount)
		if err != nil {
			return false, err
		}

		for privResult.Next() {
			var fromWalletDetails WalletDetails
			err = privResult.Scan(&fromWalletDetails.walletPrivKey)

			if err != nil {
				fmt.Println(err)
				return false, err
			}

			importedFromWallet, err := types.AccountFromBase58(fromWalletDetails.walletPrivKey)
			if err != nil {
				fmt.Println(err)
				return false, err
			}

			importedToWallet, err := types.AccountFromBase58(towalletDetails.walletPrivKey)
			if err != nil {
				fmt.Println(err)
				return false, err
			}

			c := client.NewClient(rpc.DevnetRPCEndpoint)
			resp, err := c.GetLatestBlockhash(context.Background())

			tx, err := types.NewTransaction(types.NewTransactionParam{
				Message: types.NewMessage(types.NewMessageParam{
					FeePayer:        importedFromWallet.PublicKey,
					RecentBlockhash: resp.Blockhash,
					Instructions: []types.Instruction{
						system.Transfer(system.TransferParam{
							From:   importedFromWallet.PublicKey,
							To:     importedToWallet.PublicKey,
							Amount: uint64(transactionDetails.LamportAmount),
						}),
					},
				}),
				Signers: []types.Account{importedFromWallet},
			})
			if err != nil {
				fmt.Println(err)
				return false, err
			}
			_, err = c.SendTransaction(context.Background(), tx)
			if err != nil {
				return false, err
			}
			return true, nil
		}
		return false, errors.New("first loop ended")
	}
	return false, errors.New("didnt go into loop")
}
