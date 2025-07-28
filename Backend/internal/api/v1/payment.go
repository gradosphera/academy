package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/model"
	"academy/internal/service"
	"academy/internal/service/jwt"
	"academy/internal/service/upload"
	"academy/internal/service/wayforpay"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (h *V1Handler) WayForPayUpdate(c fiber.Ctx) error {
	var update wayforpay.InvoiceStatusUpdate
	if err := c.Bind().JSON(&update); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	paymentID, err := uuid.Parse(update.OrderReference)
	if err != nil {
		return apperrors.BadRequest("invalid order reference", err)
	}

	payment, err := h.paymentService.GetByID(c.Context(), paymentID)
	if err != nil {
		return apperrors.Internal("error while getting payment", err)
	}
	if payment.MiniApp == nil {
		return apperrors.Internal("payment not includes mini app")
	}
	if len(payment.MiniApp.PaymentMetadata) == 0 {
		return apperrors.Internal("no settings for payment service provided")
	}

	var paymentMetadata model.PaymentMetadata
	err = json.Unmarshal(payment.MiniApp.PaymentMetadata, &paymentMetadata)
	if err != nil {
		return apperrors.Internal("error decoding payment metadata", err)
	}
	paymentMetadata.WayForPaySecretKey, err = h.securityService.DecryptString(paymentMetadata.WayForPaySecretKey)
	if err != nil {
		return apperrors.Internal("failed to get payments info", err)
	}

	err = wayforpay.VerifyStatus(paymentMetadata.WayForPaySecretKey, update)
	if err != nil {
		return apperrors.BadRequest("error while verifing update", err)
	}

	err = h.paymentService.UpdateWayForPayPayment(c.Context(), payment, update.TransactionStatus)
	if err != nil {
		return apperrors.BadRequest("error while updating wayforpay payment", err)
	}

	return nil
}

func (h *V1Handler) BuyProductLevelWithTON(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if claims.IsOwner || claims.IsMod {
		return apperrors.Unauthorized("only students can purchase product level")
	}

	productLevelID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productLevelID == uuid.Nil {
		return apperrors.BadRequest("invalid product id")
	}

	miniApp, err := h.miniAppService.GetByID(c.Context(), claims.MiniAppID)
	if err != nil || miniApp == nil {
		return apperrors.NotFound("mini app not found", err)
	}

	if !slices.Contains(miniApp.ActivePaymentServices, model.PaymentServiceTON) {
		return apperrors.BadRequest("mini app do not support TON payments")
	}

	productLevel, err := h.productLevelService.GetByID(c.Context(), productLevelID)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	product, err := h.productService.GetByID(c.Context(), productLevel.ProductID, false)
	if err != nil {
		return apperrors.Internal("error while getting the product", err)
	}

	if product.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	productAccess := model.NewProductAccess(claims.UserID, productLevel.ProductID)
	productAccess, err = h.productService.CheckProductAccess(c.Context(), productAccess)
	if err != nil {
		return apperrors.Internal("failed to get product access", err)
	}
	if productAccess.DeletedAt != nil {
		return apperrors.Unauthorized("user deleted from accessing the product")
	}

	if len(miniApp.PaymentMetadata) == 0 {
		return apperrors.BadRequest("payments not setup")
	}

	var paymentMetadata model.PaymentMetadata
	err = json.Unmarshal(miniApp.PaymentMetadata, &paymentMetadata)
	if err != nil {
		return apperrors.Internal("failed to get payments info", err)
	}

	if paymentMetadata.PaymentMetadataTON.TONAddress == "" {
		return apperrors.BadRequest("payments not setup")
	}

	payment, err := h.paymentService.CreateTONPayment(c.Context(),
		product, &paymentMetadata.PaymentMetadataTON, claims.UserID, productLevel)

	if err != nil {
		return apperrors.Internal("error while creating payment", err)
	}

	return c.JSON(fiber.Map{
		"payment": payment,
	})
}

func (h *V1Handler) BuyProductLevelWithWayForPay(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if claims.IsOwner || claims.IsMod {
		return apperrors.Unauthorized("only students can purchase product level")
	}

	productLevelID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productLevelID == uuid.Nil {
		return apperrors.BadRequest("invalid product id")
	}

	miniApp, err := h.miniAppService.GetByID(c.Context(), claims.MiniAppID)
	if err != nil || miniApp == nil {
		return apperrors.NotFound("mini app not found", err)
	}

	if !slices.Contains(miniApp.ActivePaymentServices, model.PaymentServiceWayForPay) {
		return apperrors.BadRequest("mini app do not support TON payments")
	}

	productLevel, err := h.productLevelService.GetByID(c.Context(), productLevelID)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	product, err := h.productService.GetByID(c.Context(), productLevel.ProductID, false)
	if err != nil {
		return apperrors.Internal("error while getting the product", err)
	}

	if product.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	productAccess := model.NewProductAccess(claims.UserID, productLevel.ProductID)
	productAccess, err = h.productService.CheckProductAccess(c.Context(), productAccess)
	if err != nil {
		return apperrors.Internal("failed to get product access", err)
	}
	if productAccess.DeletedAt != nil {
		return apperrors.Unauthorized("user deleted from accessing the product")
	}

	if len(miniApp.PaymentMetadata) == 0 {
		return apperrors.BadRequest("payments not setup")
	}

	var paymentMetadata model.PaymentMetadata
	err = json.Unmarshal(miniApp.PaymentMetadata, &paymentMetadata)
	if err != nil {
		return apperrors.Internal("failed to get payments info", err)
	}

	if paymentMetadata.PaymentMetadataWayForPay.WayForPayLogin == "" {
		return apperrors.BadRequest("payments not setup")
	}

	paymentMetadata.WayForPaySecretKey, err =
		h.securityService.DecryptString(paymentMetadata.WayForPaySecretKey)

	if err != nil {
		return apperrors.Internal("failed to get payments info", err)
	}

	var returnURL string
	if miniApp.URL != "" {
		returnURL = fmt.Sprintf("%s?startapp=product_id=%s", miniApp.URL, product.ID)
	}

	payment, err := h.paymentService.CreateWayForPayPayment(c.Context(),
		product, &paymentMetadata.PaymentMetadataWayForPay, claims.UserID, productLevel, returnURL)

	if err != nil {
		return apperrors.Internal("error while creating payment", err)
	}

	return c.JSON(fiber.Map{
		"payment": payment,
	})
}

func (h *V1Handler) GetPayments(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	var req model.GetPaymentsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	req.Limit = validateLimit(req.Limit)

	userID := claims.UserID

	if h.isPermitted(c.Context(), &claims, model.PermissionSubscriptionManagement) {
		if !claims.IsOwner {
			miniApp, err := h.miniAppService.GetByID(c.Context(), claims.MiniAppID)
			if err != nil || miniApp == nil {
				return apperrors.Internal("mini app not found", err)
			}
			if miniApp.Owner == nil {
				return apperrors.Internal("mini app has no teacher", err)
			}
			userID = miniApp.Owner.ID
		}
	}

	payments, total, err := h.paymentService.Find(c.Context(), &model.FilterPayments{
		MiniAppID: claims.MiniAppID,
		UserID:    []uuid.UUID{userID},
		Status:    req.Status,

		Limit:  req.Limit,
		Offset: req.Offset,
	})

	if err != nil {
		return apperrors.Internal("error while getting payments", err)
	}

	return c.JSON(fiber.Map{
		"payments": payments,
		"total":    total,
	})
}

func (h *V1Handler) GetPayment(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}
	paymentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if paymentID == uuid.Nil {
		return apperrors.BadRequest("invalid request data")
	}

	payment, err := h.paymentService.GetByID(c.Context(), paymentID)
	if err != nil {
		return apperrors.Internal("error while getting payment", err)
	}

	if (payment.UserID != claims.UserID) && !h.isPermitted(c.Context(), &claims, model.PermissionStudentManagement) {
		return apperrors.Unauthorized("payment access not allowed")
	}

	return c.JSON(fiber.Map{
		"payment": payment,
	})
}

func (h *V1Handler) GetStudentsPayments(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims,
		model.PermissionAccountSettings,
		model.PermissionProductsControl,
	) {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.GetStudentsPaymentsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	req.Limit = validateLimit(req.Limit)

	payments, total, err := h.paymentService.Find(c.Context(), &model.FilterPayments{
		MiniAppID: claims.MiniAppID,
		Status:    []model.PaymentStatus{model.PaymentStatusCompleted},

		Limit:  req.Limit,
		Offset: req.Offset,
	})

	if err != nil {
		return apperrors.Internal("error while getting payments", err)
	}

	return c.JSON(fiber.Map{
		"payments": payments,
		"total":    total,
	})
}

func (h *V1Handler) ExportStudentsPayments(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims,
		model.PermissionAccountSettings,
		model.PermissionProductsControl,
	) {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.ExportStudentsPaymentsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	excelData, err := h.paymentService.ExportMiniAppPayments(c.Context(), claims.MiniAppID, &req)
	if errors.Is(err, service.ErrNoData) {
		return apperrors.BadRequest("no data for selected date range")
	}
	if err != nil {
		return apperrors.Internal("error while getting excel file", err)
	}

	filenameBase := fmt.Sprintf("payments_%s-%s.xlsx", req.DateFrom, req.DateTo)

	miniAppPath := upload.MaterialFilePath{
		MiniAppID: claims.MiniAppID,
	}

	filename, _, err := h.uploadService.UploadExcel(miniAppPath.String(), excelData, filenameBase)
	if err != nil {
		return apperrors.Internal("error while saving excel file", err)
	}

	return c.JSON(fiber.Map{
		"file_path": filename,
	})
}
