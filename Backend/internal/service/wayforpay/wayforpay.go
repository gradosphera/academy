package wayforpay

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const wayForPayAPI = "https://api.wayforpay.com/api"

var allowedCurrencies = map[string]bool{
	"AED": true, "AMD": true, "AUD": true, "AZN": true, "BDT": true,
	"BGN": true, "BRL": true, "BTC": true, "BYN": true, "CAD": true,
	"CHF": true, "CNY": true, "CZK": true, "DKK": true, "DOP": true,
	"DZD": true, "EGP": true, "EUR": true, "GBP": true, "GEL": true,
	"HKD": true, "HUF": true, "IDR": true, "ILS": true, "INR": true,
	"IQD": true, "IRR": true, "JPY": true, "KGS": true, "KRW": true,
	"KZT": true, "LBP": true, "LYD": true, "MAD": true, "MDL": true,
	"MXN": true, "MYR": true, "NOK": true, "NZD": true, "PKR": true,
	"PLN": true, "RON": true, "RSD": true, "RUB": true, "SAR": true,
	"SEK": true, "SGD": true, "THB": true, "TJS": true, "TMT": true,
	"TND": true, "TRY": true, "TWD": true, "UAH": true, "BLG": true,
	"UZS": true, "VND": true, "ZAR": true,
}

var client = &http.Client{Timeout: 30 * time.Second}

func generateSignature(merchantSecretKey string, values ...string) string {
	stringToSign := strings.Join(values, ";")

	hash := hmac.New(md5.New, []byte(merchantSecretKey))
	hash.Write([]byte(stringToSign))

	return hex.EncodeToString(hash.Sum(nil))
}

func VerifyAmount(amount decimal.Decimal, currency string) error {
	if !amount.Equal(amount.RoundDown(2)) {
		return fmt.Errorf("amount include too many decimals")
	}
	if !amount.IsPositive() {
		return fmt.Errorf("amount is not a positive number")
	}

	if ok := allowedCurrencies[currency]; !ok {
		return fmt.Errorf("unsupported currency")
	}

	return nil
}
