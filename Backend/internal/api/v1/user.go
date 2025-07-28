package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/model"
	"academy/internal/service/jwt"
	"academy/internal/service/upload"
	"encoding/json"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *V1Handler) GetUser(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	user, err := h.userService.GetByID(c.Context(), claims.UserID)
	if err != nil || user == nil {
		return apperrors.NotFound("user not found", err)
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *V1Handler) EditUser(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	// Owner is not permitted to edit himself with this method. Use edit of the whole mini-app.
	if claims.IsOwner {
		return apperrors.Unauthorized("user is not permitted")
	}

	mpForm, err := c.MultipartForm()
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	users := mpForm.Value["user"]
	if len(users) != 1 {
		return apperrors.BadRequest("user not provided")
	}

	var req model.EditUserRequest
	if err := json.Unmarshal([]byte(users[0]), &req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if userNameLimit < utf8.RuneCountInString(req.FirstName) {
		return apperrors.BadRequest("user first name exceeds the limit")
	}
	if userNameLimit < utf8.RuneCountInString(req.LastName) {
		return apperrors.BadRequest("user last name exceeds the limit")
	}

	user, err := h.userService.GetByID(c.Context(), claims.UserID)
	if err != nil || user == nil {
		return apperrors.NotFound("user not found", err)
	}

	isChanged, err := req.UpdateUser(user)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	var isUpdated bool
	var newFiles []string
	var oldFiles []string
	defer func() {
		h.flushFiles(isUpdated, newFiles, oldFiles)
	}()

	if avatars := mpForm.File["avatar"]; !req.DeleteAvatar && 0 < len(avatars) {
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

		oldFiles = append(oldFiles, user.Avatar)
		newFiles = append(newFiles, avatarURL)

		user.Avatar = avatarURL
		user.AvatarSize = avatarSize
		isChanged = true

		if claims.IsOwner {
			if ownerAvatarSizeLimit < avatarSize {
				return apperrors.BadRequest("avatar size exceeds the limit")
			}
		} else {
			if avatarSizeLimit < avatarSize {
				return apperrors.BadRequest("avatar size exceeds the limit")
			}
		}
	}

	if req.DeleteAvatar {
		oldFiles = append(oldFiles, user.Avatar)
		user.Avatar = ""
		user.AvatarSize = 0
		isChanged = true
	}

	if isChanged {
		err := h.userService.Update(c.Context(), user)
		if err != nil {
			return apperrors.Internal("error while updating user", err)
		}
	}

	isUpdated = true

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func (h *V1Handler) UserHomeworks(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	var req model.FilterUserHomeworkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	req.Limit = validateLimit(req.Limit)

	homework, total, err := h.lessonProgressService.UserHomework(c.Context(), claims.UserID, &req)
	if err != nil {
		return apperrors.Internal("error while getting homework by product", err)
	}

	return c.JSON(fiber.Map{
		"homework": homework,
		"total":    total,
	})
}

func (h *V1Handler) StudentStats(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionStudentManagement, model.PermissionAnalytics) {
		return apperrors.Unauthorized("user is not permitted")
	}

	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if userID == uuid.Nil {
		return apperrors.BadRequest("invalid user id")
	}

	stats, err := h.userService.StudentStats(c.Context(), claims.MiniAppID, userID)
	if err != nil {
		return apperrors.Internal("error while getting student product stats", err)
	}

	return c.JSON(fiber.Map{
		"stats": stats,
	})
}

func (h *V1Handler) BanUser(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionStudentManagement) {
		return apperrors.Unauthorized("user is not permitted")
	}

	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if userID == uuid.Nil {
		return apperrors.BadRequest("invalid user id")
	}

	var req model.BanUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if len(req.ProductID) == 0 {
		return apperrors.BadRequest("invalid request data")
	}

	for _, productID := range req.ProductID {
		err := h.checkProduct(c.Context(), claims.MiniAppID, productID)
		if err != nil {
			return err
		}
	}

	filesToDelete, err := h.userService.Ban(c.Context(), userID, &req)
	if err != nil {
		return apperrors.Internal("error while banning user", err)
	}

	for _, f := range filesToDelete {
		err := h.uploadService.Delete(f)
		if err != nil {
			h.logger.Error("error while deleting file", zap.String("err", err.Error()))
		}
	}

	return nil
}

func (h *V1Handler) UnbanUser(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionStudentManagement) {
		return apperrors.Unauthorized("user is not permitted")
	}

	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	user, err := h.userService.GetByID(c.Context(), userID)
	if err != nil {
		return apperrors.Internal("error while getting user", err)
	}
	if user.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.UnbanUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if len(req.ProductID) == 0 {
		return apperrors.BadRequest("invalid request data")
	}

	err = h.userService.Unban(c.Context(), userID, &req)
	if err != nil {
		return apperrors.Internal("error while banning user", err)
	}

	return nil
}

func (h *V1Handler) ListBannedUser(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionStudentManagement) {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.ListBannedUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	users, total, err := h.userService.ListBanned(c.Context(), claims.MiniAppID, &req)
	if err != nil {
		return apperrors.Internal("failed to list banned users", err)
	}

	return c.JSON(fiber.Map{
		"users": users,
		"total": total,
	})
}

func (h *V1Handler) LevelUpUser(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionStudentManagement) {
		return apperrors.Unauthorized("user is not permitted")
	}

	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if userID == uuid.Nil {
		return apperrors.BadRequest("invalid user id")
	}

	var req model.LevelUpUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if len(req.ProductLevelID) == 0 {
		return apperrors.BadRequest("invalid request data")
	}

	products, err := h.productLevelService.GetProducts(c.Context(), req.ProductLevelID)
	if err != nil {
		return apperrors.Internal("faile to get product", err)
	}

	if len(products) != 1 {
		return apperrors.BadRequest("user can be leveled up only within single product")
	}

	product := products[0]

	if product.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	productLevels := make([]*model.ProductLevel, 0, len(req.ProductLevelID))
	for _, productLevelID := range req.ProductLevelID {
		productLevel, err := h.productLevelService.GetByID(c.Context(), productLevelID)
		if err != nil {
			return apperrors.BadRequest("failed to find product level by id", err)
		}

		productLevels = append(productLevels, productLevel)
	}

	err = h.userService.LevelUp(c.Context(), userID, product, productLevels)
	if err != nil {
		return apperrors.Internal("error while banning user", err)
	}

	return nil
}

func (h *V1Handler) UserLevels(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionStudentManagement, model.PermissionAnalytics) {
		return apperrors.Unauthorized("user is not permitted")
	}

	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if userID == uuid.Nil {
		return apperrors.BadRequest("invalid user id")
	}

	var req model.UserLevelsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	levels, err := h.userService.Levels(c.Context(), claims.MiniAppID, userID, &req)
	if err != nil {
		return apperrors.Internal("error while getting user levels", err)
	}

	return c.JSON(fiber.Map{
		"levels": levels,
	})
}
