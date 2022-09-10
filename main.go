package main

import (
	"fmt"

	"github.com/archangel78/blockpay-backend/app"
	config "github.com/archangel78/blockpay-backend/mysql-config"
)

func main() {
	dbConfig, err := config.GetConfig("mysql-config/mysql_config.json")

	if err != nil {
		fmt.Println(err)
		return
	}
	
	app := app.App{}
	app.Initialize(dbConfig)
}
