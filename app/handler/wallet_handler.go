package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/archangel78/blockpay-backend/app/session"
	"github.com/portto/solana-go-sdk/client"
	"github.com/portto/solana-go-sdk/rpc"
)

func CreateWallet(db *sql.DB, w http.ResponseWriter, r *http.Request, payload session.Payload) {
	// create a RPC client
	c := client.NewClient(rpc.MainnetRPCEndpoint)

	// get the current running Solana version
	response, err := c.GetVersion(context.TODO())
	if err != nil {
			panic(err)
	}

	fmt.Println("version", response.SolanaCore)
}
