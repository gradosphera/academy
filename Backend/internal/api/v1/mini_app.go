package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/config"
	"academy/internal/model"
	"academy/internal/service/jwt"
	"academy/internal/service/upload"
	"academy/internal/storage/repository"
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func (h *V1Handler) CreateMiniApp(c fiber.Ctx) error {
	validateInitData := true

	if h.config.App.Environment != config.EnvironmentProduction &&
		h.config.Auth.SkipSecurityKey != "" &&
		c.Get("X-API-Key") == h.config.Auth.SkipSecurityKey {

		validateInitData = false
	}

	var req model.CreateMiniAppRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	var botID int64
	var err error
	if req.BotToken != "" {
		if validateInitData {
			botID, err = h.telegramService.ValidateBotToken(c.Context(), req.BotToken)
			if err != nil {
				return apperrors.BadRequest("invalid bot token", err)
			}
		} else {
			splitted := strings.Split(req.BotToken, ":")
			if len(splitted) != 2 {
				return apperrors.BadRequest("invalid bot token")
			}
			number, err := strconv.Atoi(splitted[0])
			if err != nil {
				return apperrors.BadRequest("invalid bot token")
			}
			botID = int64(number)
		}

		req.BotToken, err = h.securityService.EncryptString(req.BotToken)
		if err != nil {
			return apperrors.Internal("failed to set bot token", err)
		}
	}

	initData, err := h.telegramService.ParseAdminToken(req.InitData, validateInitData)
	if err != nil {
		return apperrors.BadRequest("invalid init data", err)
	}
	miniApp := req.ToMiniApp(botID, initData.User.ID)

	if h.config.App.EnableDemoProduct {
		err := h.uploadService.AddDemoProduct(miniApp)
		if err != nil {
			return apperrors.Internal("failed to add demo product", err)
		}
	}

	owner := model.NewSignInWithTelegramMiniApp(initData, miniApp.ID, model.UserRoleOwner)

	err = h.miniAppService.Create(c.Context(), miniApp, owner)
	if err != nil {
		materialPath := upload.MaterialFilePath{MiniAppID: miniApp.ID}
		deleteErr := h.uploadService.Delete(materialPath.String())
		if deleteErr != nil {
			h.logger.Error("error while deleting temp mini-app directory", zap.String("err", deleteErr.Error()))
		}

		return apperrors.BadRequest("failed to create mini app", err)
	}

	return c.JSON(fiber.Map{
		"mini_app": miniApp,
	})
}

func (h *V1Handler) GetMiniApp(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	miniApp, err := h.miniAppService.GetByID(c.Context(), claims.MiniAppID)
	if err != nil || miniApp == nil {
		return apperrors.NotFound("mini app not found", err)
	}

	accesses, err := h.productService.ProductAccessByUser(c.Context(),
		claims.MiniAppID, claims.UserID)
	if err != nil {
		return apperrors.Internal("failed to get product accesses", err)
	}

	return c.JSON(fiber.Map{
		"mini_app": miniApp,
		"accesses": accesses,
	})
}

func (h *V1Handler) EditMiniAppAccount(c fiber.Ctx) error {
	validateInitData := true

	if h.config.App.Environment != config.EnvironmentProduction &&
		h.config.Auth.SkipSecurityKey != "" &&
		c.Get("X-API-Key") == h.config.Auth.SkipSecurityKey {

		validateInitData = false
	}

	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionAccountSettings) {
		return apperrors.Unauthorized("user is not permitted")
	}

	mpForm, err := c.MultipartForm()
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	accounts := mpForm.Value["account"]
	if len(accounts) != 1 {
		return apperrors.BadRequest("data not provided")
	}

	var req model.EditMiniAppAccountRequest
	if err := json.Unmarshal([]byte(accounts[0]), &req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if userBioLimit < utf8.RuneCountInString(req.TeacherBio) {
		return apperrors.BadRequest("user bio exceeds the limit")
	}
	if userLinksNumberLimit < len(req.TeacherLinks) {
		return apperrors.BadRequest("number of links exceeds the limit")
	}
	if userNameLimit < utf8.RuneCountInString(req.TeacherFirstName) {
		return apperrors.BadRequest("user first name exceeds the limit")
	}
	if userNameLimit < utf8.RuneCountInString(req.TeacherLastName) {
		return apperrors.BadRequest("user last name exceeds the limit")
	}
	if req.PaymentMetadataWayForPay != nil {
		var emptyWayForPayMetadata model.PaymentMetadataWayForPay

		if *req.PaymentMetadataWayForPay != emptyWayForPayMetadata {
			err := h.paymentService.VerifyWayForPay(c.Context(), req.PaymentMetadataWayForPay)
			if err != nil {
				return apperrors.BadRequest("failed to verify WayForPay payment metadata")
			}
		}
	}
	if req.PaymentMetadataTON != nil {
		if req.PaymentMetadataTON.TONAddress != "" && len(req.PaymentMetadataTON.TONAddress) != 48 {
			return apperrors.BadRequest("failed to verify TON payment metadata")
		}
	}

	miniApp, err := h.miniAppService.GetByID(c.Context(), claims.MiniAppID)
	if err != nil || miniApp == nil {
		return apperrors.NotFound("mini app not found", err)
	}

	if miniApp.Owner == nil {
		return apperrors.BadRequest("mini app has no teacher", err)
	}

	var botID int64
	if req.BotToken != "" {
		if validateInitData {
			botID, err = h.telegramService.ValidateBotToken(c.Context(), req.BotToken)
			if err != nil {
				return apperrors.BadRequest("invalid bot token", err)
			}
		} else {
			splitted := strings.Split(req.BotToken, ":")
			if len(splitted) != 2 {
				return apperrors.BadRequest("invalid bot token")
			}
			number, err := strconv.Atoi(splitted[0])
			if err != nil {
				return apperrors.BadRequest("invalid bot token")
			}
			botID = int64(number)
		}
	}

	isChangedMiniApp, err := req.UpdateMiniApp(miniApp, botID)
	if err != nil {
		return apperrors.BadRequest("error while applying the request", err)
	}
	if req.BotToken != "" {
		miniApp.BotToken, err = h.securityService.EncryptString(req.BotToken)
		if err != nil {
			return apperrors.Internal("failed to update bot token", err)
		}
		isChangedMiniApp = true
	}

	var paymentMetadata model.PaymentMetadata
	if len(miniApp.PaymentMetadata) != 0 {
		err = json.Unmarshal(miniApp.PaymentMetadata, &paymentMetadata)
		if err != nil {
			return apperrors.Internal("failed to parse payment metadata", err)
		}
	}

	if req.PaymentMetadataTON != nil {
		paymentMetadata.PaymentMetadataTON = *req.PaymentMetadataTON

		isChangedMiniApp = true
	}

	if req.PaymentMetadataWayForPay != nil {

		if req.PaymentMetadataWayForPay.WayForPaySecretKey != "" {

			req.PaymentMetadataWayForPay.WayForPaySecretKey, err =
				h.securityService.EncryptString(req.PaymentMetadataWayForPay.WayForPaySecretKey)

			if err != nil {
				return apperrors.Internal("failed to update payment metadata", err)
			}
		}

		paymentMetadata.PaymentMetadataWayForPay = *req.PaymentMetadataWayForPay

		isChangedMiniApp = true
	}

	rawPaymentMetadata, err := json.Marshal(paymentMetadata)
	if err != nil {
		return apperrors.Internal("failed to update payment metadata", err)
	}
	miniApp.PaymentMetadata = rawPaymentMetadata

	for _, activePaymentService := range miniApp.ActivePaymentServices {
		switch activePaymentService {
		case model.PaymentServiceTON:
			if paymentMetadata.PaymentMetadataTON.TONAddress == "" {
				return apperrors.Internal("TON payment service can't be activated without metadata")
			}
		case model.PaymentServiceWayForPay:
			if paymentMetadata.PaymentMetadataWayForPay.WayForPayLogin == "" {
				return apperrors.Internal("WayForPay payment service can't be activated without metadata")
			}
		}
	}

	isChangedTeacher, err := req.UpdateTeacher(miniApp.Owner)
	if err != nil {
		return apperrors.BadRequest("error while applying the request", err)
	}

	var isUpdated bool
	var newFiles []string
	var oldFiles []string
	defer func() {
		h.flushFiles(isUpdated, newFiles, oldFiles)
	}()

	if avatars := mpForm.File["avatar"]; !req.TeacherDeleteAvatar && 0 < len(avatars) {
		fileExt := strings.ToLower(filepath.Ext(avatars[0].Filename))

		if !isPictureAllowed(avatars[0].Header.Get("Content-Type"), fileExt) {
			return apperrors.BadRequest("image type is not allowed")
		}

		avatarFile, err := avatars[0].Open()
		if err != nil {
			return apperrors.BadRequest("error while opening avatar", err)
		}

		miniAppPath := upload.MaterialFilePath{MiniAppID: claims.MiniAppID}

		avatarURL, avatarSize, err := h.uploadService.Upload(miniAppPath.String(), avatarFile, fileExt)
		if err != nil {
			return apperrors.BadRequest("error while uploading avatar", err)
		}

		oldFiles = append(oldFiles, miniApp.TeacherAvatar)
		newFiles = append(newFiles, avatarURL)

		miniApp.TeacherAvatar = avatarURL
		miniApp.TeacherAvatarSize = avatarSize
		isChangedMiniApp = true

		if ownerAvatarSizeLimit < avatarSize {
			return apperrors.BadRequest("avatar size exceeds the limit")
		}
	}

	if req.TeacherDeleteAvatar {
		oldFiles = append(oldFiles, miniApp.TeacherAvatar)
		miniApp.TeacherAvatar = ""
		miniApp.TeacherAvatarSize = 0
		isChangedMiniApp = true
	}

	if isChangedMiniApp || isChangedTeacher {
		err = h.miniAppService.UpdateWithUsers(c.Context(), miniApp, miniApp.Owner)
		if errors.Is(err, repository.ErrDuplicatedKey) {
			return apperrors.BadRequest("some fields are already in use by other mini-apps", err)
		}
		if err != nil {
			return apperrors.Internal("error while updating mini-app", err)
		}
	}

	isUpdated = true

	return c.JSON(fiber.Map{
		"mini_app": miniApp,
	})
}

func (h *V1Handler) PaymentMetadata(c fiber.Ctx) error {
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

	miniApp, err := h.miniAppService.GetByID(c.Context(), claims.MiniAppID)
	if err != nil || miniApp == nil {
		return apperrors.NotFound("mini app not found", err)
	}

	var paymentMetadata model.PaymentMetadata
	if miniApp.PaymentMetadata != nil {
		err = json.Unmarshal(miniApp.PaymentMetadata, &paymentMetadata)
		if err != nil {
			return apperrors.Internal("failed to decode payment metadata", err)
		}
		paymentMetadata.WayForPaySecretKey = ""
	}

	return c.JSON(fiber.Map{
		"payment_metadata": paymentMetadata,
	})
}

func (h *V1Handler) EditMiniAppBranding(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionBranding) {
		return apperrors.Unauthorized("user is not permitted")
	}

	mpForm, err := c.MultipartForm()
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	data := mpForm.Value["branding"]
	if len(data) != 1 {
		return apperrors.BadRequest("data not provided")
	}

	var req model.EditMiniAppBrandingRequest
	if err := json.Unmarshal([]byte(data[0]), &req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	miniApp, err := h.miniAppService.GetByID(c.Context(), claims.MiniAppID)
	if err != nil || miniApp == nil {
		return apperrors.NotFound("mini app not found", err)
	}

	plan, err := h.miniAppService.GetPlanByPlanID(c.Context(), miniApp.PlanID)
	if err != nil {
		return apperrors.Internal("failed to get mini app plan", err)
	}

	if !plan.Personalization {
		return apperrors.Unauthorized("personalization not allowed for this plan")
	}

	isChanged, err := req.UpdateMiniApp(miniApp)
	if err != nil {
		return apperrors.BadRequest("error while applying the request", err)
	}

	var isUpdated bool
	var newFiles []string
	var oldFiles []string
	defer func() {
		h.flushFiles(isUpdated, newFiles, oldFiles)
	}()

	if logos := mpForm.File["logo"]; !req.DeleteLogo && 0 < len(logos) {
		fileExt := strings.ToLower(filepath.Ext(logos[0].Filename))

		if !isPictureAllowed(logos[0].Header.Get("Content-Type"), fileExt) {
			return apperrors.BadRequest("image type is not allowed")
		}

		logoFile, err := logos[0].Open()
		if err != nil {
			return apperrors.BadRequest("error while opening logo", err)
		}

		miniAppPath := upload.MaterialFilePath{MiniAppID: claims.MiniAppID}
		logoURL, logoSize, err := h.uploadService.Upload(miniAppPath.String(), logoFile, fileExt)
		if err != nil {
			return apperrors.BadRequest("error while uploading logo", err)
		}

		oldFiles = append(oldFiles, miniApp.Logo)
		newFiles = append(newFiles, logoURL)

		miniApp.Logo = logoURL
		miniApp.LogoSize = logoSize
		isChanged = true

		if miniAppLogoSizeLimit < logoSize {
			return apperrors.BadRequest("logo size exceeds the limit")
		}
	}

	if req.DeleteLogo {
		oldFiles = append(oldFiles, miniApp.Logo)
		miniApp.Logo = ""
		miniApp.LogoSize = 0
		isChanged = true
	}

	if isChanged {
		err = h.miniAppService.Update(c.Context(), miniApp)
		if err != nil {
			return apperrors.Internal("error while updating mini-app", err)
		}
	}

	isUpdated = true

	return c.JSON(fiber.Map{
		"mini_app": miniApp,
	})
}

func (h *V1Handler) EditMiniAppAnalytics(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionAnalytics) {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.EditMiniAppAnalyticsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	miniApp, err := h.miniAppService.GetByID(c.Context(), claims.MiniAppID)
	if err != nil || miniApp == nil {
		return apperrors.NotFound("mini app not found", err)
	}

	isChanged, err := req.UpdateMiniApp(miniApp)
	if err != nil {
		return apperrors.BadRequest("error while applying the request", err)
	}

	if isChanged {
		err = h.miniAppService.Update(c.Context(), miniApp)
		if err != nil {
			return apperrors.Internal("error while updating mini-app", err)
		}
	}

	return c.JSON(fiber.Map{
		"mini_app": miniApp,
	})
}

func (h *V1Handler) DeleteMiniApp(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !claims.IsOwner {
		return apperrors.Unauthorized("user is not permitted")
	}

	err := h.miniAppService.Delete(c.Context(), claims.MiniAppID)
	if err != nil {
		return apperrors.Internal("error while deleting mini-app", err)
	}

	pathToDelete := upload.MaterialFilePath{
		MiniAppID: claims.MiniAppID,
	}
	err = h.uploadService.Delete(pathToDelete.String())
	if err != nil {
		h.logger.Error("error while deleting mini-app", zap.String("err", err.Error()))
	}

	return nil
}

func (h *V1Handler) ArchiveMiniApp(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !claims.IsOwner {
		return apperrors.Unauthorized("user is not permitted")
	}

	err := h.miniAppService.SoftDelete(c.Context(), claims.MiniAppID)
	if err != nil {
		return apperrors.Internal("error while archiving mini-app", err)
	}

	return nil
}

func (h *V1Handler) UnarchiveMiniApp(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !claims.IsOwner {
		return apperrors.Unauthorized("user is not permitted")
	}

	err := h.miniAppService.Restore(c.Context(), claims.MiniAppID)
	if err != nil {
		return apperrors.Internal("error while unarchiving mini-app", err)
	}

	return nil
}

func (h *V1Handler) EditSlides(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionBranding) {
		return apperrors.Unauthorized("user is not permitted")
	}

	mpForm, err := c.MultipartForm()
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	data := mpForm.Value["slides"]
	if len(data) != 1 {
		return apperrors.BadRequest("data not provided")
	}

	var req model.EditSlidesRequest
	if err := json.Unmarshal([]byte(data[0]), &req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	var isUpdated bool
	var newFiles []string
	var oldFiles []string
	defer func() {
		h.flushFiles(isUpdated, newFiles, oldFiles)
	}()

	var newSlide *model.Material
	if newSlides := mpForm.File["new_slide"]; 0 < len(newSlides) {
		fileExt := strings.ToLower(filepath.Ext(newSlides[0].Filename))

		if !isPictureAllowed(newSlides[0].Header.Get("Content-Type"), fileExt) {
			return apperrors.BadRequest("image type is not allowed")
		}

		slideFile, err := newSlides[0].Open()
		if err != nil {
			return apperrors.BadRequest("error while opening logo", err)
		}

		miniAppPath := upload.MaterialFilePath{MiniAppID: claims.MiniAppID}
		slideURL, slideSize, err := h.uploadService.Upload(miniAppPath.String(), slideFile, fileExt)
		if err != nil {
			return apperrors.BadRequest("error while uploading logo", err)
		}

		newFiles = append(newFiles, slideURL)

		if slidesSizeLimit < slideSize {
			return apperrors.BadRequest("slide size exceeds the limit")
		}

		newSlide = req.ToMaterial(claims.MiniAppID, slideURL, slideSize)

		if newSlide.Title == "" {
			return apperrors.BadRequest("slide title not provided")
		}
		if newSlide.Description == "" {
			return apperrors.BadRequest("slide description not provided")
		}
		if slideTitleLimit < utf8.RuneCountInString(newSlide.Title) {
			return apperrors.BadRequest("slide title exceeds the limit")
		}
		if slideDescriptionLimit < utf8.RuneCountInString(newSlide.Description) {
			return apperrors.BadRequest("slide description exceeds the limit")
		}
	}

	numberOfSlides := len(req.SlidesOrder)
	if newSlide != nil {
		numberOfSlides++
	}
	if slidesNumberLimit < numberOfSlides {
		return apperrors.BadRequest("number of slides exceeds limit")
	}

	miniApp, err := h.miniAppService.ReorderSlides(c.Context(), claims.MiniAppID, &req, newSlide)
	if err != nil {
		return apperrors.Internal("failed to reorder slides", err)
	}
	if miniApp == nil {
		return apperrors.NotFound("mini app not found", err)
	}

	isUpdated = true

	return c.JSON(fiber.Map{
		"mini_app": miniApp,
	})
}

func (h *V1Handler) Analytics(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionAnalytics) {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.AnalyticsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	analytics, err := h.miniAppService.Analytics(c.Context(), claims.MiniAppID, &req)
	if err != nil {
		return apperrors.Internal("error while getting analytics", err)
	}

	return c.JSON(fiber.Map{
		"analytics": analytics,
	})
}

func (h *V1Handler) Info(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims,
		model.PermissionStudentManagement,
		model.PermissionProductsControl,
		model.PermissionSubscriptionManagement,
		model.PermissionAnalytics,
		model.PermissionBranding,
		model.PermissionAccountSettings,
		model.PermissionStudentInteraction,
	) {
		return apperrors.Unauthorized("user is not permitted")
	}

	info, err := h.miniAppService.GetInfo(c.Context(), claims.MiniAppID)
	if err != nil {
		return apperrors.Internal("error while getting mini-app info", err)
	}

	return c.JSON(fiber.Map{
		"info": info,
	})
}
