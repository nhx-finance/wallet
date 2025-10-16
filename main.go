package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/nhx-finance/wallet/internal/app"
	"github.com/nhx-finance/wallet/internal/routes"
)

func main() {
	orcus, err := app.NewApplication()

	if err != nil {
		panic(err)
	}

	defer func(DB *sql.DB) {
		err := DB.Close()
		if err != nil {
			panic(err)
		}
	}(orcus.DB)

	var port int
	flag.IntVar(&port, "port", 8080, "backend sever port")
	flag.Parse()

	orcus.Logger.Println("Application running")

	r := routes.SetUpRoutes(orcus)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           r,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 30,
	}

	orcus.Logger.Printf("Listening on port: %d", port)

	err = server.ListenAndServe()

	if err != nil {
		orcus.Logger.Fatal(err)
	}
}