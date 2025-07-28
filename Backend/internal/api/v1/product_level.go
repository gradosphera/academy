package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/model"
	"academy/internal/service/jwt"
	"academy/internal/service/upload"
	"academy/internal/service/wayforpay"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *V1Handler) CreateProductLevel(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.CreateProductLevelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productLevelNameLimit < utf8.RuneCountInString(req.Name) {
		return apperrors.BadRequest("product level name exceeds the limit")
	}

	err := wayforpay.VerifyAmount(req.Price, req.Currency)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	err = wayforpay.VerifyAmount(req.FullPrice, req.Currency)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	product, err := h.productService.GetByID(c.Context(), req.ProductID, true)
	if err != nil {
		return apperrors.Internal("error while getting the product", err)
	}

	if product.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	productLevel, err := req.ToProductLevel(int64(len(product.Levels)), product.AccessTime)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	err = h.productLevelService.Create(c.Context(), productLevel, req.LessonIDs)
	if err != nil {
		return apperrors.Internal("failed to create product level", err)
	}

	return c.JSON(fiber.Map{
		"product_level": productLevel,
	})
}

func (h *V1Handler) EditProductLevel(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productLevelID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productLevelID == uuid.Nil {
		return apperrors.BadRequest("invalid product level id")
	}

	var req model.EditProductLevelRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productLevelNameLimit < utf8.RuneCountInString(req.Name) {
		return apperrors.BadRequest("product level name exceeds the limit")
	}

	productLevel, err := h.productLevelService.GetByID(c.Context(), productLevelID)
	if err != nil {
		return apperrors.BadRequest("failed to find product level by id", err)
	}

	if err := h.checkProduct(c.Context(), claims.MiniAppID, productLevel.ProductID); err != nil {
		return err
	}

	isChanged, err := req.UpdateProductLevel(productLevel)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if len(req.AddLessonIDs) != 0 || len(req.RemoveLessonIDs) != 0 {
		isChanged = true
	}

	if isChanged {
		err = h.productLevelService.Update(c.Context(), productLevel, req.AddLessonIDs, req.RemoveLessonIDs)
		if err != nil {
			return apperrors.Internal("error while updating product level", err)
		}
	}

	return c.JSON(fiber.Map{
		"product_level": productLevel,
	})
}

func (h *V1Handler) DeleteProductLevel(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productLevelID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productLevelID == uuid.Nil {
		return apperrors.BadRequest("invalid product level id")
	}

	productLevel, err := h.productLevelService.GetByID(c.Context(), productLevelID)
	if err != nil {
		return apperrors.BadRequest("failed to find product level by id", err)
	}

	if err := h.checkProduct(c.Context(), claims.MiniAppID, productLevel.ProductID); err != nil {
		return err
	}

	err = h.productLevelService.Delete(c.Context(), productLevelID)
	if err != nil {
		return apperrors.Internal("error while deleting product level", err)
	}

	pathToDelete := upload.MaterialFilePath{
		MiniAppID:      claims.MiniAppID,
		ProductID:      productLevel.ProductID,
		ProductLevelID: productLevel.ID,
	}
	err = h.uploadService.Delete(pathToDelete.String())
	if err != nil {
		h.logger.Error("error while deleting product level", zap.String("err", err.Error()))
	}

	return nil
}

func (h *V1Handler) CreateProductLevelInvite(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionStudentManagement) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productLevelID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	productLevel, err := h.productLevelService.GetByID(c.Context(), productLevelID)
	if err != nil {
		return apperrors.BadRequest("failed to find product level by id", err)
	}

	if err := h.checkProduct(c.Context(), claims.MiniAppID, productLevel.ProductID); err != nil {
		return err
	}

	invite := model.NewProductLevelInvite(productLevelID)

	err = h.productLevelService.CreateProductLevelInvite(c.Context(), invite)
	if err != nil {
		return apperrors.Internal("error while creating product level invite", err)
	}

	return c.JSON(fiber.Map{
		"invite": invite,
	})
}
