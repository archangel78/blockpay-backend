package main

import (

	"github.com/archangel78/blockpay-backend/app"
)

func main() {
	app := app.App{}
	app.SetRoutes()
}
