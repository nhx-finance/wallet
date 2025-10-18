package payments

import (
	"log"
	"net/http"

	"github.com/nhx-finance/wallet/internal/utils"
	"github.com/stripe/stripe-go/v83"
)


type StripeHandler struct {
	StripeClient *stripe.Client
}

func NewStripeHandler(stripeClient *stripe.Client) *StripeHandler {
	return &StripeHandler{
		StripeClient: stripeClient,
	}
}

func (sh *StripeHandler) AccountBalance(w http.ResponseWriter, r *http.Request){
	params := &stripe.BalanceRetrieveParams{}
	balance, err := sh.StripeClient.V1Balance.Retrieve(r.Context(), params)
	if err != nil {
		log.Printf("failed to get account balance: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "failed to get account balance"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"balance": balance})
}