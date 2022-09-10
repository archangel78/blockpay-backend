package common

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	config "github.com/archangel78/blockpay-backend/mysql-config"
)

func OpenDbConnection(dbConfig *config.DbConfig) (*sql.DB, error) {
	conString := dbConfig.Username+":"+dbConfig.Password+"@tcp("+dbConfig.Hostname+":"+dbConfig.Port+")/"+dbConfig.DatabaseName
	db, err := sql.Open(dbConfig.Protocol, conString)

	if err != nil {
		return nil, err
	}
	return db, nil
}