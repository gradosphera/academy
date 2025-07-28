package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/model"
	"academy/internal/service"
	"academy/internal/service/jwt"
	"academy/internal/service/upload"
	"academy/internal/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *V1Handler) CreateProduct(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	mpForm, err := c.MultipartForm()
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	products := mpForm.Value["product"]
	if len(products) != 1 {
		return apperrors.BadRequest("products not provided")
	}

	var req model.CreateProductRequest
	if err := json.Unmarshal([]byte(products[0]), &req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productTitleLimit < utf8.RuneCountInString(req.Title) {
		return apperrors.BadRequest("product title exceeds the limit")
	}
	if productContentTypeLimit < utf8.RuneCountInString(req.ContentType) {
		return apperrors.BadRequest("product type exceeds the limit")
	}
	if productDescriptionLimit < utf8.RuneCountInString(req.Description) {
		return apperrors.BadRequest("product description exceeds the limit")
	}

	product, err := req.ToProduct(claims.MiniAppID)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	var isUpdated bool
	var newFiles []string
	var oldFiles []string
	defer func() {
		h.flushFiles(isUpdated, newFiles, oldFiles)
	}()

	if covers := mpForm.File["cover"]; 0 < len(covers) {
		fileExt := strings.ToLower(filepath.Ext(covers[0].Filename))

		if !isPictureAllowed(covers[0].Header.Get("Content-Type"), fileExt) {
			return apperrors.BadRequest("image type is not allowed")
		}

		coverFile, err := covers[0].Open()
		if err != nil {
			return apperrors.BadRequest("error while opening cover", err)
		}

		productPath := upload.MaterialFilePath{
			MiniAppID: claims.MiniAppID,
			ProductID: product.ID,
		}
		coverURL, coverSize, err := h.uploadService.Upload(productPath.String(), coverFile, fileExt)
		if err != nil {
			return apperrors.BadRequest("error while uploading cover", err)
		}

		newFiles = append(newFiles, coverURL)
		product.Cover = coverURL
		product.CoverSize = coverSize

		if productCoverSizeLimit < coverSize {
			return apperrors.BadRequest("product cover size exceeds the limit")
		}
	}

	err = h.productService.Create(c.Context(), product)
	if err != nil {
		return apperrors.Internal("failed to create product", err)
	}

	isUpdated = true

	return c.JSON(fiber.Map{
		"product": product,
	})
}

func (h *V1Handler) GetProduct(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if productID == uuid.Nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	productAccess := model.NewProductAccess(claims.UserID, productID)
	productAccess, err = h.productService.CheckProductAccess(c.Context(), productAccess)
	if err != nil {
		return apperrors.Internal("failed to get product access", err)
	}
	if productAccess.DeletedAt != nil {
		return apperrors.Unauthorized("user deleted from accessing the product")
	}

	product, err := h.productService.GetByID(c.Context(), productID, true)
	if err != nil {
		return apperrors.Internal("failed to get product", err)
	}

	if product.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	isAdmin := h.isPermitted(c.Context(), &claims, model.PermissionProductsControl)
	isStudent := !claims.IsOwner && !claims.IsMod

	if !isAdmin && !product.IsActive {
		return apperrors.BadRequest("product not active", err)
	}

	// Do not use product.AccessTime because it used as a default duration
	// for product level by front-end.
	if !isAdmin && product.ReleaseDate.Valid {
		err := isAccessible(product.ReleaseDate, types.Interval{})
		if err != nil {
			return apperrors.Unauthorized("product not accessible", err)
		}
	}

	var reviews []*model.Review
	var unlockedLessons []model.UnlockedLesson
	var progress []*model.LessonProgress

	if isStudent {
		unlockedLessons, err = h.lessonService.UnlockedLessons(
			c.Context(), productID, claims.UserID)
		if err != nil {
			return apperrors.Internal("failed to get unlocked lessons", err)
		}

		progress, err = h.lessonProgressService.GetByProductID(
			c.Context(),
			claims.UserID,
			productID,
		)

		if err != nil {
			return apperrors.Internal("failed to get progress", err)
		}

		reviews, err = h.reviewService.GetByUser(c.Context(), claims.UserID, productID)
		if err != nil {
			return apperrors.Internal("failed to get product reviews", err)
		}
	}

	return c.JSON(fiber.Map{
		"product":          product,
		"reviews":          reviews,
		"unlocked_lessons": unlockedLessons,
		"progress":         progress,
		"access":           productAccess,
	})
}

func (h *V1Handler) EditProduct(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	product, err := h.productService.GetByID(c.Context(), productID, true)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if product.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	mpForm, err := c.MultipartForm()
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	products := mpForm.Value["product"]
	if len(products) != 1 {
		return apperrors.BadRequest("products not provided")
	}

	var req model.EditProductRequest
	if err := json.Unmarshal([]byte(products[0]), &req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productTitleLimit < utf8.RuneCountInString(req.Title) {
		return apperrors.BadRequest("product title exceeds the limit")
	}
	if productContentTypeLimit < utf8.RuneCountInString(req.ContentType) {
		return apperrors.BadRequest("product type exceeds the limit")
	}
	if productDescriptionLimit < utf8.RuneCountInString(req.Description) {
		return apperrors.BadRequest("product description exceeds the limit")
	}

	var applyNewLessonAccess *model.LessonAccess
	if req.LessonAccess != product.LessonAccess {
		applyNewLessonAccess = &req.LessonAccess
	}

	var applyNewReleaseDate *types.Time
	if !req.ReleaseDate.IsEqual(product.ReleaseDate) {
		applyNewReleaseDate = &req.ReleaseDate
	}

	var applyNewAccessTime *types.Interval
	if !req.AccessTime.IsEqual(product.AccessTime) {
		applyNewAccessTime = &req.AccessTime
	}

	isChanged, err := req.UpdateProduct(product)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	var isUpdated bool
	var newFiles []string
	var oldFiles []string
	defer func() {
		h.flushFiles(isUpdated, newFiles, oldFiles)
	}()

	if covers := mpForm.File["cover"]; !req.DeleteCover && 0 < len(covers) {
		fileExt := strings.ToLower(filepath.Ext(covers[0].Filename))

		if !isPictureAllowed(covers[0].Header.Get("Content-Type"), fileExt) {
			return apperrors.BadRequest("image type is not allowed")
		}

		coverFile, err := covers[0].Open()
		if err != nil {
			return apperrors.BadRequest("error while opening cover", err)
		}

		productPath := upload.MaterialFilePath{
			MiniAppID: claims.MiniAppID,
			ProductID: product.ID,
		}
		coverURL, coverSize, err := h.uploadService.Upload(productPath.String(), coverFile, fileExt)
		if err != nil {
			return apperrors.BadRequest("error while uploading cover", err)
		}

		oldFiles = append(oldFiles, product.Cover)
		newFiles = append(newFiles, coverURL)

		product.Cover = coverURL
		product.CoverSize = coverSize
		isChanged = true

		if productCoverSizeLimit < coverSize {
			return apperrors.BadRequest("product cover size exceeds the limit")
		}
	}

	if req.DeleteCover {
		oldFiles = append(oldFiles, product.Cover)
		product.Cover = ""
		product.CoverSize = 0
		isChanged = true
	}

	if isChanged {
		err = h.productService.Update(
			c.Context(),
			product,
			applyNewLessonAccess,
			applyNewReleaseDate,
			applyNewAccessTime,
		)
		if err != nil {
			return apperrors.Internal("failed to update product", err)
		}
	}

	isUpdated = true

	return c.JSON(fiber.Map{
		"product": product,
	})
}

func (h *V1Handler) DeleteProduct(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productID == uuid.Nil {
		return apperrors.BadRequest("invalid product id")
	}

	product, err := h.productService.GetByID(c.Context(), productID, false)
	if err != nil {
		return apperrors.Internal("error while getting the product", err)
	}

	if product.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	err = h.productService.Delete(c.Context(), productID)
	if err != nil {
		return apperrors.Internal("error while deleting the product", err)
	}

	pathToDelete := upload.MaterialFilePath{
		MiniAppID: claims.MiniAppID,
		ProductID: productID,
	}
	err = h.uploadService.Delete(pathToDelete.String())
	if err != nil {
		h.logger.Error("error while deleting product", zap.String("err", err.Error()))
	}

	return nil
}

func (h *V1Handler) ProductInvites(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionStudentManagement) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if productID == uuid.Nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if err := h.checkProduct(c.Context(), claims.MiniAppID, productID); err != nil {
		return err
	}

	var req model.FilterProductLevelInvitesRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	invites, total, err := h.productLevelService.ProductLevelInvites(c.Context(), productID, &req)
	if err != nil {
		return apperrors.Internal("error while finding product invites", err)
	}

	return c.JSON(fiber.Map{
		"invites": invites,
		"total":   total,
	})
}

func (h *V1Handler) ProductHomeworks(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionStudentInteraction) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productID == uuid.Nil {
		return apperrors.BadRequest("invalid product id")
	}

	if err := h.checkProduct(c.Context(), claims.MiniAppID, productID); err != nil {
		return err
	}

	var req model.FilterProductHomeworkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	req.Limit = validateLimit(req.Limit)

	homework, total, err := h.lessonProgressService.ProductHomework(c.Context(), productID, &req)
	if err != nil {
		return apperrors.Internal("error while getting homework by product", err)
	}

	return c.JSON(fiber.Map{
		"homework": homework,
		"total":    total,
	})
}

func (h *V1Handler) ReorderProductLessons(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if err := h.checkProduct(c.Context(), claims.MiniAppID, productID); err != nil {
		return err
	}

	var req model.ReorderProductLessonsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	for _, module := range req.Modules {
		if moduleNameLimit < utf8.RuneCountInString(module.ModuleName) {
			return apperrors.BadRequest("module name exceeds the limit")
		}
	}

	product, err := h.productService.ReorderLessons(c.Context(), productID, &req)
	if err != nil {
		return apperrors.Internal("error while reordering lessons", err)
	}

	for _, lessonID := range req.LessonsToDelete {
		pathToDelete := upload.MaterialFilePath{
			MiniAppID: claims.MiniAppID,
			ProductID: productID,
			LessonID:  lessonID,
		}
		if err := h.uploadService.Delete(pathToDelete.String()); err != nil {
			h.logger.Error("error while deleting lesson",
				zap.String("lesson_id", lessonID.String()),
				zap.Error(err),
			)
		}
	}

	return c.JSON(fiber.Map{
		"product": product,
	})
}

func (h *V1Handler) ReorderProductLevels(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if err := h.checkProduct(c.Context(), claims.MiniAppID, productID); err != nil {
		return err
	}

	var req model.ReorderProductLevelsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	product, err := h.productService.ReorderLevels(c.Context(), productID, &req)
	if err != nil {
		return apperrors.Internal("error while reordering lessons", err)
	}

	return c.JSON(fiber.Map{
		"product": product,
	})
}

func (h *V1Handler) ProductFeedback(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionAnalytics) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productID == uuid.Nil {
		return apperrors.BadRequest("invalid product id")
	}

	if err := h.checkProduct(c.Context(), claims.MiniAppID, productID); err != nil {
		return err
	}

	feedback, err := h.productService.Feedback(c.Context(), productID)
	if err != nil {
		return apperrors.Internal("error while getting product feedback", err)
	}

	return c.JSON(fiber.Map{
		"feedback": feedback,
	})
}

func (h *V1Handler) ProductStudents(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionAnalytics) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productID == uuid.Nil {
		return apperrors.BadRequest("invalid product id")
	}

	if err := h.checkProduct(c.Context(), claims.MiniAppID, productID); err != nil {
		return err
	}

	offset := fiber.Query[uint](c, "offset")

	limit := fiber.Query[uint](c, "limit")
	limit = validateLimit(limit)

	usernameSearch := fiber.Query[string](c, "username_search")
	if userTelegramUsernameLimit < len(usernameSearch) {
		return apperrors.BadRequest("username is too long")
	}

	students, err := h.productService.Students(c.Context(), productID, usernameSearch, limit, offset)
	if err != nil {
		return apperrors.Internal("error while getting product students", err)
	}

	return c.JSON(fiber.Map{
		"students": students,
	})
}

func (h *V1Handler) ExportProductStudents(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionAnalytics) {
		return apperrors.Unauthorized("user is not permitted")
	}

	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if productID == uuid.Nil {
		return apperrors.BadRequest("invalid product id")
	}

	product, err := h.productService.GetByID(c.Context(), productID, true)
	if err != nil {
		return apperrors.Internal("error while getting the product", err)
	}

	if product.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.ExportProductStudentsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	excelData, err := h.productService.StudentsExportToExcel(c.Context(), product, &req)
	if errors.Is(err, service.ErrNoData) {
		return apperrors.BadRequest("no data for selected date range")
	}
	if err != nil {
		return apperrors.Internal("error while getting excel file", err)
	}

	filenameBase := fmt.Sprintf("students_progress_report_%s-%s.xlsx", req.DateFrom, req.DateTo)

	productPath := upload.MaterialFilePath{
		MiniAppID: claims.MiniAppID,
		ProductID: productID,
	}
	filename, _, err := h.uploadService.UploadExcel(productPath.String(), excelData, filenameBase)
	if err != nil {
		return apperrors.Internal("error while saving excel file", err)
	}

	// c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	// c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	// c.Set("Content-Length", fmt.Sprintf("%d", len(excelFile)))

	return c.JSON(fiber.Map{
		"file_path": filename,
	})
}

func (h *V1Handler) checkProduct(ctx context.Context, miniAppID, productID uuid.UUID) error {
	product, err := h.productService.GetByID(ctx, productID, false)
	if err != nil {
		return apperrors.Internal("error while getting the product", err)
	}

	if product.MiniAppID != miniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	return nil
}

func (h *V1Handler) checkLesson(ctx context.Context, miniAppID, lessonID uuid.UUID) error {
	lesson, err := h.lessonService.GetByID(ctx, lessonID, uuid.Nil)
	if err != nil {
		return apperrors.Internal("error while getting the lesson", err)
	}

	return h.checkProduct(ctx, miniAppID, lesson.ProductID)
}
