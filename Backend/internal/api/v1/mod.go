package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/config"
	"academy/internal/model"
	"academy/internal/service/jwt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (h *V1Handler) CreateModInvite(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !claims.IsOwner {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.CreateModInviteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	invite, err := h.miniAppService.CreateModInvite(c.Context(), claims.MiniAppID, &req)
	if err != nil {
		return apperrors.Internal("failed to create mod invite", err)
	}

	return c.JSON(fiber.Map{
		"invite": invite.ID,
	})
}

func (h *V1Handler) ModInvites(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !claims.IsOwner {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.FilterModInvitesRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	invites, total, err := h.miniAppService.ModInvites(c.Context(), claims.MiniAppID, &req)
	if err != nil {
		return apperrors.Internal("failed to get mod invites", err)
	}

	return c.JSON(fiber.Map{
		"invites": invites,
		"total":   total,
	})
}

func (h *V1Handler) DeleteModInvite(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !claims.IsOwner {
		return apperrors.Unauthorized("user is not permitted")
	}

	inviteID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	err = h.miniAppService.DeleteModInvite(c.Context(), claims.MiniAppID, inviteID)
	if err != nil {
		return apperrors.Internal("failed to delete mod invite", err)
	}

	return nil
}

func (h *V1Handler) EditModInvite(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !claims.IsOwner {
		return apperrors.Unauthorized("user is not permitted")
	}

	inviteID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	invite, err := h.miniAppService.GetModInviteByID(c.Context(), inviteID)
	if err != nil {
		return apperrors.BadRequest("failed to get invite", err)
	}

	if invite.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.EditModInviteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	err = h.miniAppService.EditModInvite(c.Context(), inviteID, &req)
	if err != nil {
		return apperrors.Internal("failed to edit mod invite", err)
	}

	return nil
}

func (h *V1Handler) ListMiniApps(c fiber.Ctx) error {
	validateInitData := true

	if h.config.App.Environment != config.EnvironmentProduction &&
		h.config.Auth.SkipSecurityKey != "" &&
		c.Get("X-API-Key") == h.config.Auth.SkipSecurityKey {

		validateInitData = false
	}

	var req model.ListMiniAppsRequest
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

	miniApps, err := h.miniAppService.GetByModTelegramID(c.Context(), initData.User.ID)
	if err != nil {
		return apperrors.Internal("error getting mini apps", err)
	}

	if len(miniApps) == 0 {
		return apperrors.Unauthorized("user is not moderator in any of the mini-apps")
	}

	return c.JSON(fiber.Map{
		"mini_apps": miniApps,
	})
}

func (h *V1Handler) ModPermissions(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !claims.IsOwner {
		return apperrors.Unauthorized("user is not permitted")
	}

	permissions, err := h.miniAppService.ListPermissions(c.Context())
	if err != nil {
		return apperrors.Internal("failed to get list of mod permissions", err)
	}

	return c.JSON(fiber.Map{
		"permissions": permissions,
	})
}
