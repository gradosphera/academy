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
)

type createInvoiceRequest struct {
	TransactionType         string `json:"transactionType"`
	MerchantAccount         string `json:"merchantAccount"`
	MerchantTransactionType string `json:"merchantTransactionType,omitempty"`
	MerchantAuthType        string `json:"merchantAuthType,omitempty"`
	MerchantDomainName      string `json:"merchantDomainName"`
	MerchantSignature       string `json:"merchantSignature"`

	APIVersion   string `json:"apiVersion"`
	Language     string `json:"language,omitempty"`
	NotifyMethod string `json:"notifyMethod,omitempty"`
	ServiceURL   string `json:"serviceUrl,omitempty"`
	ReturnURL    string `json:"returnUrl,omitempty"`

	OrderReference string `json:"orderReference"`
	OrderDate      int64  `json:"orderDate"`
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`

	ProductName  []string `json:"productName"`
	ProductPrice []string `json:"productPrice"`
	ProductCount []string `json:"productCount"`

	// OrderTimeout        string `json:"orderTimeout,omitempty"`
	// HoldTimeout         string `json:"holdTimeout,omitempty"`

	// AlternativeAmount   string `json:"alternativeAmount,omitempty"`
	// AlternativeCurrency string `json:"alternativeCurrency,omitempty"`

	// PaymentSystems      string `json:"paymentSystems,omitempty"`

	// ClientFirstName     string `json:"clientFirstName,omitempty"`
	// ClientLastName      string `json:"clientLastName,omitempty"`
	// ClientEmail         string `json:"clientEmail,omitempty"`
	// ClientPhone         string `json:"clientPhone,omitempty"`
}

type createInvoiceResponse struct {
	ReasonCode     int64  `json:"reasonCode"`
	Reason         string `json:"reason"`
	InvoiceURL     string `json:"invoiceUrl"`
	OrderReference string `json:"orderReference"`
	QRCode         string `json:"qrCode"`
}

func CreateInvoice(
	ctx context.Context,
	merchantAccount, merchantSecretKey, merchantDomainName,
	serviceURL,
	orderID, productName, price, currency, redirectURL string,
) (string, error) {

	orderDate := time.Now().Unix()

	productNames := []string{productName}
	productCounts := []string{"1"}
	productPrices := []string{price}

	signatureValues := []string{
		merchantAccount,
		merchantDomainName,
		orderID,
		strconv.FormatInt(orderDate, 10),
		price,
		currency,
		productName,
		"1",
		price,
	}

	signature := generateSignature(merchantSecretKey, signatureValues...)

	requestBody := createInvoiceRequest{
		TransactionType:         "CREATE_INVOICE",
		MerchantAccount:         merchantAccount,
		MerchantTransactionType: "",
		MerchantAuthType:        "SimpleSignature",
		MerchantDomainName:      merchantDomainName,
		MerchantSignature:       signature,

		APIVersion:   "1",
		Language:     "EN",
		NotifyMethod: "all",
		ServiceURL:   serviceURL,
		ReturnURL:    redirectURL,

		OrderReference: orderID,
		OrderDate:      orderDate,
		Amount:         price,
		Currency:       currency,

		ProductName:  productNames,
		ProductPrice: productPrices,
		ProductCount: productCounts,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wayForPayAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request to WayForPay: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to WayForPay: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("io.ReadAll: %w", err)
	}

	var invoiceResponse createInvoiceResponse
	if err := json.Unmarshal(respBody, &invoiceResponse); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if invoiceResponse.ReasonCode != 1100 {
		return "", fmt.Errorf("WayForPay error: %s (code: %d)", invoiceResponse.Reason, invoiceResponse.ReasonCode)
	}

	return invoiceResponse.InvoiceURL, nil
}
