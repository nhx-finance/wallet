package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	DefaultExchangeRate = 128.83
)

type Envelope map[string]any

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


func GetUSDCKSHExchangeRate() (float64, error) {
	// TODO: Implement actual exchange rate fetching
	return DefaultExchangeRate, nil
}