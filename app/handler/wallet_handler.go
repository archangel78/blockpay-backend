package handler

import (
	"context"
	"database/sql"
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
	PrivateKey    string `json:"Privatekey"`
	SolanaVersion string `json:"Version"`
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
		common.RespondError(w, 500, "Some internal error occurred CNMNET")
		return
	}

	wallet := types.NewAccount()
	response := WalletCreateResponse{
		PublicKey: wallet.PublicKey.ToBase58(),
		PrivateKey: base58.Encode(wallet.PrivateKey),
		SolanaVersion: version.SolanaCore,
	}

	_, err = db.Exec("INSERT INTO Wallet (accountName, wallet) VALUES (?, ?)", payload.AccountName, wallet.PublicKey.ToBase58())

	if err != nil {
		common.RespondError(w, 500, "Some internal error occurred ")
		return
	}
	common.RespondJSON(w, 200, response)
}
