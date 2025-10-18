package payments

import (
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