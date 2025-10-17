package stores

import (
	"database/sql"
	"time"
)

type Transaction struct {
	ID string `json:"id"`
	Phone string `json:"phone"`
	HederaAccountID string `json:"hedera_account_id"`
	Type string `json:"type"`
	AmountKSH float64 `json:"amount_ksh"`
	AmountUSDC float64 `json:"amount_usdc"`
	ExchangeRate float64 `json:"exchange_rate"`
	Status string `json:"status"`
	MpesaCheckoutID string `json:"mpesa_checkout_id"`
	MpesaReceiptNumber string `json:"mpesa_receipt_number"`
	HederaTxID string `json:"hedera_tx_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostgresTransactionStore struct {
	db *sql.DB
}

func NewPostgresTransactionStore(db *sql.DB) *PostgresTransactionStore {
	return &PostgresTransactionStore{db: db}
}

type TransactionStore interface {
	CreateTransaction(tx Transaction) (*Transaction, error)
}

func (pt *PostgresTransactionStore) CreateTransaction(tx Transaction) (*Transaction, error) {
	query := `
	INSERT INTO transactions (
	phone, hedera_account_id, type, amount_ksh, amount_usdc, exchange_rate, status, mpesa_checkout_id, created_at, updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id, phone, hedera_account_id, type, amount_ksh, amount_usdc, exchange_rate, status, mpesa_checkout_id, created_at, updated_at
	`

	err := pt.db.QueryRow(query, tx.Phone, tx.HederaAccountID, tx.Type, tx.AmountKSH, tx.AmountUSDC, tx.ExchangeRate, tx.Status, tx.MpesaCheckoutID, tx.CreatedAt, tx.UpdatedAt).
	Scan(&tx.ID, &tx.Phone, &tx.HederaAccountID, &tx.Type, &tx.AmountKSH, &tx.AmountUSDC, &tx.ExchangeRate, &tx.Status, &tx.MpesaCheckoutID, &tx.CreatedAt, &tx.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}