package common

import (
	"database/sql"
	"fmt"

	config "github.com/archangel78/blockpay-backend/mysql-config"
	_ "github.com/go-sql-driver/mysql"
)

func OpenDbConnection(dbConfig *config.DbConfig) (*sql.DB, error) {
	conString := dbConfig.Username + ":" + dbConfig.Password + "@tcp(" + dbConfig.Hostname + ":" + dbConfig.Port + ")/" + dbConfig.DatabaseName
	db, err := sql.Open(dbConfig.Protocol, conString)

	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetPreparedStatement(db *sql.DB, preparedString string) (*sql.Stmt, error) {
	stmt, err := db.Prepare(preparedString)
	if err != nil {
		return nil, err
	}
	return stmt, err
}

func WriteTransaction(db *sql.DB, transactionId string, fromAccount string, toAccount string, toWallet string, transactionAmount string, toName string, fromName string) error {
	_, err := db.Exec("INSERT INTO Transactions (transactionId, fromAccount, toAccount, toWallet, transactionAmount, toName, fromName) VALUES (?, ?, ?, ?, ?, ?, ?)", transactionId, fromAccount, toAccount, toWallet, transactionAmount, toName, fromName)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
