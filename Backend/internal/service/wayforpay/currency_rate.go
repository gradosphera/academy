package wayforpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type currencyRateRequest struct {
	TransactionType   string `json:"transactionType"`
	MerchantAccount   string `json:"merchantAccount"`
	MerchantSignature string `json:"merchantSignature"`
	OrderDate         int64  `json:"orderDate"`
	APIVersion        string `json:"apiVersion"`
}

type currencyRateResponse struct {
	ReasonCode int64                      `json:"reasonCode"`
	Reason     string                     `json:"reason"`
	RatesDate  int64                      `json:"ratesDate"`
	Rates      map[string]decimal.Decimal `json:"rates"`
}

func CurrencyRates(
	ctx context.Context,
	merchantAccount, merchantSecretKey string,
) (map[string]decimal.Decimal, error) {

	orderDate := time.Now().Unix()

	signatureValues := []string{
		merchantAccount,
		strconv.FormatInt(orderDate, 10),
	}

	signature := generateSignature(merchantSecretKey, signatureValues...)

	requestBody := currencyRateRequest{
		TransactionType:   "CURRENCY_RATES",
		MerchantAccount:   merchantAccount,
		MerchantSignature: signature,
		OrderDate:         orderDate,
		APIVersion:        "1",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wayForPayAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request to WayForPay: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to WayForPay: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	var currResp currencyRateResponse
	if err := json.Unmarshal(respBody, &currResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if currResp.ReasonCode != 1100 {
		return nil, fmt.Errorf("WayForPay error: %s (code: %d)", currResp.Reason, currResp.ReasonCode)
	}

	return currResp.Rates, nil
}
