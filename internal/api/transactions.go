package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
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

type STKPushResponse struct {
	MerchantRequestID string `json:"MerchantRequestID"`
	CheckoutRequestID string `json:"CheckoutRequestID"`
	ResponseCode string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage string `json:"CustomerMessage"`
}

type AuthorizationResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn string `json:"expires_in"`
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
	if req.AmountKSH < 1 {
		th.Logger.Printf("amount KSH is too small, amountKSH: %f", req.AmountKSH)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "amount is too small, minimum amount is 1 KSH"})
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
	stkPushResp, err := initiateSTKPush(req)
	if err != nil {
		th.Logger.Println("failed to initiate STK push", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "failed to initiate STK push"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"transaction": tx, "stk_push_response": *stkPushResp})
	
}

func initiateSTKPush(req OnRampRequest) (*STKPushResponse, error) {
	log.Printf("initiating STK push for request: %+v\n", req)
	url := os.Getenv("STK_PUSH_URL")
	if url == "" {
		return nil, errors.New("STK_PUSH_URL is not set")
	}
	businessShortCode := os.Getenv("BUSINESS_SHORT_CODE")
	if businessShortCode == "" {
		return nil, errors.New("BUSINESS_SHORT_CODE is not set")
	}
	consumerKey := os.Getenv("CONSUMER_KEY")
	if consumerKey == "" {
		return nil, errors.New("CONSUMER_KEY is not set")
	}
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	if consumerSecret == "" {
		return nil, errors.New("CONSUMER_SECRET is not set")
	}
	passKey := os.Getenv("PASS_KEY")
	if passKey == "" {
		return nil, errors.New("PASS_KEY is not set")
	}

	businessShortCodeInt, err := strconv.ParseInt(businessShortCode, 10, 64)
	if err != nil {
		return nil, errors.New("BUSINESS_SHORT_CODE must be a valid integer")
	}
	phoneInt, err := strconv.ParseInt(req.Phone, 10, 64)
	if err != nil {
		return nil, errors.New("phone number must be a valid integer")
	}

	method := "POST"
	timestamp := time.Now().Format("YYYYMMDDHHmmss")
	password := base64.StdEncoding.EncodeToString([]byte(businessShortCode + passKey + timestamp))
	payloadData := map[string]any{
		"BusinessShortCode": businessShortCodeInt,
		"Password": password,
		"Timestamp": timestamp,
		"TransactionType": "CustomerPayBillOnline",
		"Amount": int(req.AmountKSH),
		"PartyA": phoneInt,
		"PartyB": businessShortCodeInt,
		"PhoneNumber": phoneInt,
		"CallBackURL": "https://mydomain.com/path",
		"AccountReference": "NHXWALLET",
		"TransactionDesc": "USDC Purchase",
	}

	log.Println("payload data: ", payloadData)
	
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		log.Println("failed to marshal payload", err)
		return nil, err
	}
	payload := bytes.NewReader(payloadBytes)
	client := &http.Client{}
	httpReq, err := http.NewRequest(method, url, payload)
	if err != nil {
		log.Println("failed to create HTTP request", err)
		return nil, err
	}

	accessToken, err := getAccessToken()
	if err != nil {
		log.Println("failed to get access token", err)
		return nil, err
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("Authorization", "Bearer " + accessToken.AccessToken)
	
	res, err := client.Do(httpReq)
	if err != nil {
		log.Println("failed to do HTTP request", err)
		return nil, err
	}
	log.Println("STK push response status code: ", res.StatusCode)
	log.Println("STK push response: ", res.Body)
	defer res.Body.Close()
	
	var stkResp STKPushResponse
	if err := json.NewDecoder(res.Body).Decode(&stkResp); err != nil {
		log.Println("failed to decode STK push response", err)
		return nil, err
	}
	
	return &stkResp, nil
}

func getAccessToken() (*AuthorizationResponse, error) {
	url := os.Getenv("AUTHORIZATION_URL")
	if url == "" {
		return nil, errors.New("AUTHORIZATION_URL is not set")
	}
	consumerKey := os.Getenv("CONSUMER_KEY")
	if consumerKey == "" {
		return nil, errors.New("CONSUMER_KEY is not set")
	}
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	if consumerSecret == "" {
		return nil, errors.New("CONSUMER_SECRET is not set")
	}

	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(consumerKey + ":" + consumerSecret))

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Add("Authorization", "Basic " + encodedCredentials)
	httpReq.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	res, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var authResp AuthorizationResponse
	if err := json.NewDecoder(res.Body).Decode(&authResp); err != nil {
		return nil, err
	}
	
	return &authResp, nil
}