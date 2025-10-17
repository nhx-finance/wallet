package utils

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

const (
	DefaultExchangeRate = 128.83
)

type Envelope map[string]interface{}

func WriteJSON(w http.ResponseWriter, status int, data Envelope) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}
	return nil
}

func ReadParamID(r *http.Request, key string) (string, error) {
	idParam := chi.URLParam(r, key)
	if idParam == "" {
		return "", errors.New("id parameter is required")
	}
	return idParam, nil
}


type CoinPriceResponse struct {
	Code string          `json:"code"`
	Data CoinPriceData   `json:"data"`
}


type CoinPriceData struct {
	FiatExchangeRate FiatExchangeRate `json:"fiatExchangeRate"`
}


type FiatExchangeRate struct {
	Name    string  `json:"name"`
	Symbol  string  `json:"symbol"`
	Sign    string  `json:"sign"`
	UsdRate float64 `json:"usdRate"`
}

func GetUSDCKSHExchangeRate() (float64, error) {
	url := os.Getenv("EXCHANGE_RATE_URL")
	if url == "" {
		log.Println("EXCHANGE_RATE_URL is not set, using default exchange rate")
		return DefaultExchangeRate, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Println("failed to fetch exchange rate, using default exchange rate", err)
		return DefaultExchangeRate, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("failed to fetch exchange rate, using default exchange rate")
		return DefaultExchangeRate, nil
	}

	var priceResp CoinPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		log.Println("failed to decode exchange rate response, using default exchange rate", err)
		return DefaultExchangeRate, err
	}

	return priceResp.Data.FiatExchangeRate.UsdRate, nil
}