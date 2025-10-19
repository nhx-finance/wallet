package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/nhx-finance/wallet/internal/stores"
	"github.com/nhx-finance/wallet/internal/utils"
)

type WebhookHandler struct {
	WebhookStore stores.WebhookStore
	Logger *log.Logger
}

func NewWebhookHandler(webhookStore stores.WebhookStore, logger *log.Logger) *WebhookHandler {
	return &WebhookHandler{
		WebhookStore: webhookStore,
		Logger: logger,
	}
}

type SuccessfulRequest struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value any `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

type FailedRequest struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

func (wh *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var callback SuccessfulRequest
	err := json.NewDecoder(r.Body).Decode(&callback)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request body"})
		wh.Logger.Printf("failed to decode successful request: %v", err)
		return
	}

	if callback.Body.StkCallback.ResultCode != 0 {
		wh.Logger.Printf("failed to process successful request: %v", callback.Body)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": callback.Body.StkCallback.ResultDesc})
		return
	}
}