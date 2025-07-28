package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/config"
	"academy/internal/model"
	"academy/internal/service/jwt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (h *V1Handler) AdminSignIn(c fiber.Ctx) error {
	validateInitData := true

	if h.config.App.Environment != config.EnvironmentProduction &&
		h.config.Auth.SkipSecurityKey != "" &&
		c.Get("X-API-Key") == h.config.Auth.SkipSecurityKey {

		validateInitData = false
	}

	var req model.SignInAdminRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	initData, err := h.telegramService.ParseAdminToken(req.InitData, validateInitData)
	if err != nil {
		return apperrors.BadRequest("failed to parse token", err)
	}

	if initData.User.ID == 0 {
		return apperrors.BadRequest("no user id provided in initialization data", err)
	}

	miniApp, err := h.miniAppService.GetByOwnerTelegramID(c.Context(), initData.User.ID)
	if err != nil || miniApp == nil {
		return apperrors.BadRequest("invalid user id", err)
	}

	user, err := h.userService.SignInWithTelegramMiniApp(c.Context(), miniApp.ID, initData, model.UserRoleOwner)
	if err != nil {
		return apperrors.BadRequest("failed to sign in", err)
	}

	claims := &jwt.TokenClaims{
		MiniAppID: miniApp.ID,
		UserID:    user.ID,
		IsOwner:   true,
	}

	tokenPair, err := h.jwtService.GenerateTokenPair(c.Context(), claims)
	if err != nil {
		return apperrors.Internal("failed to generate response", err)
	}

	jwtInfo := &model.SignInJWTResp{
		RefreshToken: tokenPair.RefreshToken,
		AccessToken:  tokenPair.AccessToken,
	}

	return c.JSON(fiber.Map{
		"user":     user,
		"jwt_info": jwtInfo,
	})
}

func (h *V1Handler) ModSignIn(c fiber.Ctx) error {
	validateInitData := true

	if h.config.App.Environment != config.EnvironmentProduction &&
		h.config.Auth.SkipSecurityKey != "" &&
		c.Get("X-API-Key") == h.config.Auth.SkipSecurityKey {

		validateInitData = false
	}

	var req model.SignInModRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	initData, err := h.telegramService.ParseAdminToken(req.InitData, validateInitData)
	if err != nil {
		return apperrors.BadRequest("failed to parse token", err)
	}

	if initData.User.ID == 0 {
		return apperrors.BadRequest("no user id provided in initialization data", err)
	}

	var user *model.User
	switch {
	case req.MiniAppID != uuid.Nil:
		user, err = h.userService.SignInWithTelegramMiniApp(c.Context(), req.MiniAppID, initData, model.UserRoleModerator)
		if err != nil {
			return apperrors.BadRequest("failed to sign in", err)
		}
	case req.InviteID != uuid.Nil:
		invite, err := h.miniAppService.GetModInviteByID(c.Context(), req.InviteID)
		if err != nil {
			return apperrors.BadRequest("failed to find invite", err)
		}
		if invite.UserID != uuid.Nil {
			return apperrors.BadRequest("invite already claimed")
		}
		req.MiniAppID = invite.MiniAppID

		user, err = h.userService.SignInWithTelegramMiniApp(c.Context(), invite.MiniAppID, initData, model.UserRoleModerator)
		if err != nil {
			return apperrors.BadRequest("failed to sign in", err)
		}

		invite.UserID = user.ID
		invite.UpdatedAt = time.Now()
		err = h.miniAppService.UpdateModInvite(c.Context(), invite)
		if err != nil {
			return apperrors.BadRequest("failed to claim invite", err)
		}
		req.MiniAppID = invite.MiniAppID
	default:
		return apperrors.Internal("unexpected error")
	}

	permissions, err := h.miniAppService.GetPermissions(c.Context(), user.ID)
	if err != nil {
		return apperrors.Internal("failed to get permissions", err)
	}

	claims := &jwt.TokenClaims{
		MiniAppID: req.MiniAppID,
		UserID:    user.ID,
		IsMod:     true,
	}

	tokenPair, err := h.jwtService.GenerateTokenPair(c.Context(), claims)
	if err != nil {
		return apperrors.Internal("failed to generate response", err)
	}

	jwtInfo := &model.SignInJWTResp{
		RefreshToken: tokenPair.RefreshToken,
		AccessToken:  tokenPair.AccessToken,
	}

	return c.JSON(fiber.Map{
		"user":        user,
		"jwt_info":    jwtInfo,
		"permissions": permissions,
	})
}

func (h *V1Handler) SignIn(c fiber.Ctx) error {
	validateInitData := true

	if h.config.App.Environment != config.EnvironmentProduction &&
		h.config.Auth.SkipSecurityKey != "" &&
		c.Get("X-API-Key") == h.config.Auth.SkipSecurityKey {

		validateInitData = false
	}

	var req model.SignInRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	miniApp, err := h.miniAppService.GetByName(c.Context(), req.MiniAppName)
	if err != nil || miniApp == nil {
		return apperrors.BadRequest("invalid mini-app name", err)
	}

	if miniApp.BotID == 0 {
		return apperrors.Unauthorized("no bot available for mini-app")
	}

	if !miniApp.IsActive {
		return apperrors.BadRequest("mini-app not found")
	}

	initData, err := h.telegramService.ParseToken(miniApp.BotID, req.InitData, validateInitData)
	if err != nil {
		return apperrors.BadRequest("failed to parse token", err)
	}

	if initData.User.ID == 0 {
		return apperrors.BadRequest("no user id provided in initialization data", err)
	}

	user, err := h.userService.SignInWithTelegramMiniApp(c.Context(), miniApp.ID, initData, model.UserRoleStudent)
	if err != nil {
		return apperrors.BadRequest("failed to sign in", err)
	}

	claims := &jwt.TokenClaims{
		MiniAppID: miniApp.ID,
		UserID:    user.ID,
	}

	tokenPair, err := h.jwtService.GenerateTokenPair(c.Context(), claims)
	if err != nil {
		return apperrors.Internal("failed to generate response", err)
	}

	jwtInfo := &model.SignInJWTResp{
		RefreshToken: tokenPair.RefreshToken,
		AccessToken:  tokenPair.AccessToken,
	}

	return c.JSON(fiber.Map{
		"user":     user,
		"jwt_info": jwtInfo,
	})
}

func (h *V1Handler) SignInWithInvite(c fiber.Ctx) error {
	validateInitData := true

	if h.config.App.Environment != config.EnvironmentProduction &&
		h.config.Auth.SkipSecurityKey != "" &&
		c.Get("X-API-Key") == h.config.Auth.SkipSecurityKey {

		validateInitData = false
	}

	var req model.SignInWithInviteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	miniApp, err := h.miniAppService.GetByName(c.Context(), req.MiniAppName)
	if err != nil || miniApp == nil {
		return apperrors.BadRequest("invalid mini-app name", err)
	}

	if miniApp.BotID == 0 {
		return apperrors.Unauthorized("no bot available for mini-app")
	}

	initData, err := h.telegramService.ParseToken(miniApp.BotID, req.InitData, validateInitData)
	if err != nil {
		return apperrors.BadRequest("failed to parse token", err)
	}

	if initData.User.ID == 0 {
		return apperrors.BadRequest("no user id provided in initialization data", err)
	}

	user, err := h.userService.SignInWithTelegramMiniApp(c.Context(), miniApp.ID, initData, model.UserRoleStudent)
	if err != nil {
		return apperrors.BadRequest("failed to sign in", err)
	}

	claims := &jwt.TokenClaims{
		MiniAppID: miniApp.ID,
		UserID:    user.ID,
	}

	tokenPair, err := h.jwtService.GenerateTokenPair(c.Context(), claims)
	if err != nil {
		return apperrors.Internal("failed to generate response", err)
	}

	jwtInfo := &model.SignInJWTResp{
		RefreshToken: tokenPair.RefreshToken,
		AccessToken:  tokenPair.AccessToken,
	}

	productLevel, err := h.productLevelService.ClaimInvite(c.Context(), req.InviteID, user.ID)
	if err != nil {
		return apperrors.BadRequest("failed to claim invite", err)
	}

	return c.JSON(fiber.Map{
		"user":             user,
		"jwt_info":         jwtInfo,
		"product_id":       productLevel.ProductID,
		"product_level_id": productLevel.ID,
	})
}
