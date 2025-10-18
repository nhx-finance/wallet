package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	hiero "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"
	"github.com/joho/godotenv"
	"github.com/nhx-finance/wallet/internal/api"
	"github.com/nhx-finance/wallet/internal/payments"
	"github.com/nhx-finance/wallet/internal/stores"
	"github.com/nhx-finance/wallet/migrations"
	"github.com/stripe/stripe-go/v83"
)

type Application struct {
	Logger *log.Logger
	DB *sql.DB
	HieroClient *hiero.Client
	TransactionHandler *api.TransactionHandler
	StripeHandler *payments.StripeHandler
}

func loadEnvironmentVariables() error {
	err := godotenv.Load()
	
	if err != nil {
		fmt.Println("No .env file found (using environment variables from system)")
		return err
	} else {
		fmt.Println("Environment variables loaded from .env file")
		return nil
	}
}



func NewApplication() (*Application, error) {
	if err := loadEnvironmentVariables(); err != nil {
		return nil, err
	}
	accountID, err := hiero.AccountIDFromString(os.Getenv("OPERATOR_ACCOUNT_ID"))
	if err != nil {
		panic(err)
	}

	privateKey, err := hiero.PrivateKeyFromStringEd25519(os.Getenv("OPERATOR_KEY"))
	if err != nil {
		panic(err)
	}

	client := hiero.ClientForTestnet()

	client.SetOperator(accountID, privateKey)
	stripeClient := stripe.NewClient(os.Getenv("STRIPE_SECRET"))
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	pgDB, err := stores.Open()
	if err != nil {
		return nil, err
	}
	err = stores.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// stores
	transactionStore := stores.NewPostgresTransactionStore(pgDB)

	// handlers
	stripeHandler := payments.NewStripeHandler(stripeClient)
	transactionHandler := api.NewTransactionHandler(transactionStore, client, logger, stripeHandler)
	
	app := &Application{
		Logger: logger,
		HieroClient: client,
		DB: pgDB,
		TransactionHandler: transactionHandler,
		StripeHandler: stripeHandler,
	}

	return app, nil
}

func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}