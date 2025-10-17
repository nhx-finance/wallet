package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	hiero "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"
	"github.com/nhx-finance/wallet/internal/stores"
	"github.com/nhx-finance/wallet/internal/utils"
)

type OnRampRequest struct {
	AmountKSH float64 `json:"amount_ksh"`
	Phone string `json:"phone"`
	HederaAccountID string `json:"hedera_account_id"`
}

type TransactionHandler struct {
	TransactionStore stores.TransactionStore
	HieroClient *hiero.Client
	Logger *log.Logger
}

func NewTransactionHandler (transactionStore stores.TransactionStore, hieroClient *hiero.Client, logger *log.Logger) *TransactionHandler {
	return &TransactionHandler{
		TransactionStore: transactionStore,
		HieroClient: hieroClient,
		Logger: logger,
	}
}

func (th *TransactionHandler) HandleOnramp(w http.ResponseWriter, r *http.Request){
	var req OnRampRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		th.Logger.Println("failed to decode onramp request", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "failed to decode onramp request"})
		return
	}

	exchangeRate, err := utils.GetUSDCKSHExchangeRate()
	if err != nil {
		th.Logger.Println("failed to get exchange rate", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "failed to get exchange rate"})
		return
	}

	amountUSDC := req.AmountKSH / exchangeRate
	if amountUSDC < 0.006 {
		th.Logger.Printf("amount USDC is too small, amountKSH: %f, amountUSDC: %f", req.AmountKSH, amountUSDC)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "amount is too small, minimum amount is 0.006 USDC(~1 KSH)"})
		return
	}
	
	tx := stores.Transaction{
		Phone: req.Phone,
		HederaAccountID: req.HederaAccountID,
		Type: "onramp",
		AmountKSH: req.AmountKSH,
		AmountUSDC: amountUSDC,
		ExchangeRate: exchangeRate,
		Status: "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}