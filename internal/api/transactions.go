package api

import (
	"encoding/json"
	"log"
	"net/http"

	hiero "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"
	"github.com/nhx-finance/wallet/internal/payments"
	"github.com/nhx-finance/wallet/internal/stores"
	"github.com/nhx-finance/wallet/internal/utils"
)

type OnRampRequest struct {
	Email string `json:"email"`
	Asset string `json:"asset"`
	Quantity float64 `json:"quantity"`
	ImageURL string `json:"image_url"`
}

type TransactionHandler struct {
	TransactionStore stores.TransactionStore
	StripeHandler *payments.StripeHandler
	HieroClient *hiero.Client
	Logger *log.Logger
}

func NewTransactionHandler (transactionStore stores.TransactionStore, hieroClient *hiero.Client, logger *log.Logger, stripeHandler *payments.StripeHandler) *TransactionHandler {
	return &TransactionHandler{
		TransactionStore: transactionStore,
		HieroClient: hieroClient,
		Logger: logger,
		StripeHandler: stripeHandler,
	}
}

func (th *TransactionHandler) HandleCardOnramp(w http.ResponseWriter, r *http.Request){
	var req OnRampRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request body"})
		return
	}

	session, err := th.StripeHandler.CreateCheckoutSession(req.Email, req.Asset, req.Quantity, req.ImageURL)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "failed to create checkout session"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"session": session})
}
