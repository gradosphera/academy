package wayforpay

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	TransactionStatusInProcessing        TransactionStatus = "InProcessing"
	TransactionStatusWaitingAuthComplete TransactionStatus = "WaitingAuthComplete"
	TransactionStatusApproved            TransactionStatus = "Approved"
	TransactionStatusPending             TransactionStatus = "Pending"
	TransactionStatusExpired             TransactionStatus = "Expired"
	TransactionStatusRefunded            TransactionStatus = "Refunded"
	TransactionStatusVoided              TransactionStatus = "Voided"
	TransactionStatusDeclined            TransactionStatus = "Declined"
	TransactionStatusRefundInProcessing  TransactionStatus = "RefundInProcessing"
)

type InvoiceStatusUpdate struct {
	MerchantAccount   string `json:"merchantAccount"`
	OrderReference    string `json:"orderReference"`
	MerchantSignature string `json:"merchantSignature"`

	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
	AuthCode string          `json:"authCode"`

	Email string `json:"email"`
	Phone string `json:"phone"`

	CreatedDate       int64  `json:"createdDate"`
	ProcessingDate    int64  `json:"processingDate"`
	CardPan           string `json:"cardPan"`
	CardType          string `json:"cardType"`
	IssuerBankCountry string `json:"issuerBankCountry"`
	IssuerBankName    string `json:"issuerBankName"`
	RecToken          string `json:"recToken"`

	TransactionStatus TransactionStatus `json:"transactionStatus"`
	Reason            string            `json:"reason"`
	ReasonCode        int64             `json:"reasonCode"`
	Fee               decimal.Decimal   `json:"fee"`
	PaymentSystem     string            `json:"paymentSystem"`
}

func VerifyStatus(merchantSecretKey string, update InvoiceStatusUpdate) error {
	signatureValues := []string{
		update.MerchantAccount,
		update.OrderReference,
		update.Amount.String(),
		update.Currency,
		update.AuthCode,
		update.CardPan,
		string(update.TransactionStatus),
		strconv.Itoa(int(update.ReasonCode)),
	}

	expectedSignature := generateSignature(merchantSecretKey, signatureValues...)

	if update.MerchantSignature != expectedSignature {
		return fmt.Errorf("provided update signature is invalid")
	}

	return nil
}
