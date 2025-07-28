package service

import (
	"academy/internal/config"
	repo "academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/service/wayforpay"
	"academy/internal/storage/repository"
	"academy/internal/types"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type PaymentService struct {
	paymentRepository  *repository.PaymentRepository
	transactionManager *repo.TransactionManager
	webhookURL         string
	disableRefund      bool

	adminWayForPayLogin     string
	adminWayForPaySecretKey string
}

func NewPaymentService(
	cfg *config.Config,
	paymentRepository *repository.PaymentRepository,
	transactionManager *repo.TransactionManager,
) *PaymentService {

	return &PaymentService{
		paymentRepository:  paymentRepository,
		transactionManager: transactionManager,
		webhookURL:         cfg.HTTP.WayForPayWebhook,
		disableRefund:      cfg.Auth.WayForPayDisableRefund,

		adminWayForPayLogin:     cfg.TON.WayForPayLogin,
		adminWayForPaySecretKey: cfg.TON.WayForPaySecretKey,
	}
}

func (s *PaymentService) CreateTONPayment(
	ctx context.Context,
	product *model.Product,
	paymentMetadata *model.PaymentMetadataTON,
	userID uuid.UUID,
	productLevel *model.ProductLevel,
) (*model.Payment, error) {

	rates, err := wayforpay.CurrencyRates(ctx, s.adminWayForPayLogin, s.adminWayForPaySecretKey)
	if err != nil {
		return nil, fmt.Errorf("error while checking currency rates: %w", err)
	}

	payment := model.NewPaymentForProductLevel(userID, product, productLevel)
	payment.URL = paymentMetadata.TONAddress

	var amountUSD decimal.Decimal
	if payment.Currency == "UAH" {
		usdRate, ok := rates["USD"]
		if !ok {
			return nil, fmt.Errorf("no USD rate available")
		}
		amountUSD = payment.Amount.Div(usdRate).RoundDown(2)
	} else {
		currRate, ok := rates[payment.Currency]
		if !ok {
			return nil, fmt.Errorf("no rate available for %q", payment.Currency)
		}
		usdRate, ok := rates["USD"]
		if !ok {
			return nil, fmt.Errorf("no USD rate available")
		}
		amountUSD = payment.Amount.Mul(currRate).Div(usdRate).RoundDown(2)
	}
	payment.AmountUSD = amountUSD

	err = s.paymentRepository.Create(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("error while saving payment: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) CreateWayForPayPayment(
	ctx context.Context,
	product *model.Product,
	paymentMetadata *model.PaymentMetadataWayForPay,
	userID uuid.UUID,
	productLevel *model.ProductLevel,
	returnURL string,
) (*model.Payment, error) {

	rates, err := wayforpay.CurrencyRates(
		ctx, paymentMetadata.WayForPayLogin, paymentMetadata.WayForPaySecretKey)

	if err != nil {
		return nil, fmt.Errorf("error while checking currency rates: %w", err)
	}

	payment := model.NewPaymentForProductLevel(userID, product, productLevel)

	invoiceURL, err := wayforpay.CreateInvoice(
		ctx,
		paymentMetadata.WayForPayLogin,
		paymentMetadata.WayForPaySecretKey,
		paymentMetadata.WayForPayDomainName,
		s.webhookURL,
		payment.ID.String(),
		productLevel.Name,
		productLevel.Price.RoundDown(2).StringFixed(2),
		productLevel.Currency,
		returnURL,
	)

	if err != nil {
		return nil, fmt.Errorf("error while creating invoice: %w", err)
	}
	payment.URL = invoiceURL

	var amountUSD decimal.Decimal
	if payment.Currency == "UAH" {
		usdRate, ok := rates["USD"]
		if !ok {
			return nil, fmt.Errorf("no USD rate available")
		}
		amountUSD = payment.Amount.Div(usdRate).RoundDown(2)
	} else {
		currRate, ok := rates[payment.Currency]
		if !ok {
			return nil, fmt.Errorf("no rate available for %q", payment.Currency)
		}
		usdRate, ok := rates["USD"]
		if !ok {
			return nil, fmt.Errorf("no USD rate available")
		}
		amountUSD = payment.Amount.Mul(currRate).Div(usdRate).RoundDown(2)
	}
	payment.AmountUSD = amountUSD

	err = s.paymentRepository.Create(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("error while saving payment: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) UpdateWayForPayPayment(
	ctx context.Context,
	payment *model.Payment,
	newTransactionStatus wayforpay.TransactionStatus,
) error {

	var newStatus model.PaymentStatus
	var isRefund bool
	switch newTransactionStatus {

	case wayforpay.TransactionStatusInProcessing:
		newStatus = model.PaymentStatusPending

	case wayforpay.TransactionStatusWaitingAuthComplete:
		newStatus = model.PaymentStatusPending

	case wayforpay.TransactionStatusApproved:
		newStatus = model.PaymentStatusCompleted

	case wayforpay.TransactionStatusPending:
		newStatus = model.PaymentStatusPending

	case wayforpay.TransactionStatusExpired:
		newStatus = model.PaymentStatusFailed

	case wayforpay.TransactionStatusRefunded, wayforpay.TransactionStatusVoided:
		newStatus = model.PaymentStatusRefunded
		isRefund = true

	case wayforpay.TransactionStatusDeclined:
		newStatus = model.PaymentStatusPending

	case wayforpay.TransactionStatusRefundInProcessing:
		newStatus = model.PaymentStatusPendingRefund
		isRefund = true

	default:
		return fmt.Errorf("unexected transaction status: %v", newTransactionStatus)
	}

	// Skip already refunded.
	if payment.Status == model.PaymentStatusRefunded {
		return nil
	}

	// Skip if refunds disabled.
	if s.disableRefund && isRefund {
		return nil
	}

	// Skip if status already non-pending.
	if !isRefund && payment.Status != model.PaymentStatusPending {
		return nil
	}

	// Skip if no changes.
	if payment.Status == newStatus {
		return nil
	}

	now := time.Now().UTC()

	if newStatus == model.PaymentStatusCompleted && payment.AccessStart.Time.Before(now) {
		payment.AccessStart = types.NewTime(now)
	}

	payment.Status = newStatus
	payment.UpdatedAt = now

	err := s.paymentRepository.Update(ctx, payment)
	if err != nil {
		return fmt.Errorf("error while updating payment: %w", err)
	}

	return nil
}

func (s *PaymentService) Find(
	ctx context.Context,
	filter *model.FilterPayments,
) ([]*model.Payment, int, error) {

	payments, total, err := s.paymentRepository.Find(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("error while getting payments: %w", err)
	}

	return payments, total, nil
}

func (s *PaymentService) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	payment, err := s.paymentRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment by id: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) VerifyWayForPay(
	ctx context.Context, metadata *model.PaymentMetadataWayForPay,
) error {

	_, err := wayforpay.CurrencyRates(
		ctx, metadata.WayForPayLogin, metadata.WayForPaySecretKey)

	if err != nil {
		return fmt.Errorf("error while checking currency rates: %w", err)
	}

	return nil
}

func (s *PaymentService) ExportMiniAppPayments(
	ctx context.Context,
	miniAppID uuid.UUID,
	req *model.ExportStudentsPaymentsRequest,
) ([]byte, error) {

	dateFrom, err := time.Parse(time.DateOnly, req.DateFrom)
	if err != nil {
		return nil, fmt.Errorf("error parsing DateFrom: %w", err)
	}
	dateTo, err := time.Parse(time.DateOnly, req.DateTo)
	if err != nil {
		return nil, fmt.Errorf("error parsing DateTo: %w", err)
	}

	studentsPayments, err := s.paymentRepository.StudentsPayments(ctx, miniAppID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("failed to get mini-app students payments: %w", err)
	}

	if len(studentsPayments) == 0 {
		return nil, ErrNoData
	}

	f := excelize.NewFile()

	headers := []string{
		"Telegram ID", "Telegram Username",
		"Product", "Product Tier", "Amount (USD)", "Purchase Time",
	}

	cell, _ := excelize.CoordinatesToCellName(1, 1)
	err = f.SetSheetRow("Sheet1", cell, &headers)
	if err != nil {
		return nil, fmt.Errorf("failed to set headers: %w", err)
	}
	for i, p := range studentsPayments {
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		paymentRow := []any{
			p.TelegramID,
			p.TelegramUsername,

			p.ProductName,
			p.ProductLevelName,
			p.AmountUSD,
			p.PaidAt,
		}
		err = f.SetSheetRow("Sheet1", cell, &paymentRow)
		if err != nil {
			return nil, fmt.Errorf("failed to set payment %v: %w", p.PaymentID.String(), err)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("could not write excel to buffer: %w", err)
	}

	return buf.Bytes(), nil
}
