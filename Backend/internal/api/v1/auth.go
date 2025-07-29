package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/model"
	"academy/internal/service/upload"
	"fmt"
	"net/textproto"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *V1Handler) JWTAuthMiddleware(c fiber.Ctx) error {
	token := c.Get("Authorization")

	claims, err := h.jwtService.ParseAccessToken(token)
	if err != nil {
		return apperrors.Unauthorized("failed to parse token", err)
	}

	if claims == nil {
		return apperrors.Internal("failed to parse token")
	}

	c.Locals("claims", *claims)

	return c.Next()
}

func (h *V1Handler) MaterialAuthMiddleware(c fiber.Ctx) error {
	token := c.Query("jwt")
	if token == "" {
		return apperrors.Unauthorized("claims not found")
	}

	claims, err := h.jwtService.ParseAccessToken(token)
	if err != nil {
		return apperrors.Unauthorized("failed to parse token", err)
	}

	if claims == nil {
		return apperrors.Internal("failed to parse token")
	}

	filename, ok := strings.CutPrefix(c.Path(), uploadPath)
	if !ok {
		return apperrors.BadRequest("unexpected path")
	}

	materialPath, err := upload.ParseMaterialFilePath(filename)
	if err != nil {
		return apperrors.BadRequest("unexpected path", err)
	}

	// Restricted users from wrong mini-app.
	if claims.MiniAppID != materialPath.MiniAppID {
		return apperrors.BadRequest("mini-app access is restricted")
	}

	if claims.IsOwner || claims.IsMod {
		return c.Next()
	}

	// Restricted users from seen other users homework submition files.
	if materialPath.UserID != uuid.Nil {
		if claims.UserID != materialPath.UserID {
			return apperrors.BadRequest("user access is restricted")
		}
		return c.Next()
	}

	// Allow access to product/mini-app resourses.
	if materialPath.LessonID == uuid.Nil && materialPath.ProductLevelID == uuid.Nil {
		return c.Next()
	}

	material, err := h.materialService.GetByFilename(c.Context(), filename)
	if err != nil {
		return apperrors.Internal("error getting material", err)
	}

	if material == nil {
		return c.Next()
	}

	if material.ProductLevelID != uuid.Nil {
		isUnlocked, err := h.productLevelService.IsProductLevelUnlocked(
			c.Context(), material.ProductLevelID, claims.UserID)

		if err != nil {
			return apperrors.Internal("error getting product level", err)
		}

		if !isUnlocked {
			return apperrors.Unauthorized("product level access restricted", err)
		}

		return c.Next()
	}

	if material.LessonID != uuid.Nil {
		if material.Category == model.MaterialCategoryLessonCover {
			return c.Next()
		}

		_, err = h.validateLessonAccess(c.Context(), claims, material.LessonID)
		if err != nil {
			return apperrors.Unauthorized("material access restricted", err)
		}
	}

	return c.Next()
}

func (h *V1Handler) RefreshTokens(c fiber.Ctx) error {
	var tokenReq struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind().JSON(&tokenReq); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	claims, err := h.jwtService.ParseRefreshToken(tokenReq.RefreshToken)
	if err != nil {
		return apperrors.Unauthorized("failed to parse token", err)
	}

	if claims == nil {
		return apperrors.Internal("failed to parse token")
	}

	storedRefresh, err := h.jwtService.GetRefreshByUserID(c.Context(), claims.UserID)
	if err != nil {
		return apperrors.Unauthorized("invalid token claims", err)
	}

	if tokenReq.RefreshToken != storedRefresh {
		return apperrors.Unauthorized("refresh token does not match with stored one")
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
		"jwt_info": jwtInfo,
	})
}

// HandleMaterialHeaders fixes bugs with playing video content. Currently not needed.
func (h *V1Handler) HandleMaterialHeaders(c fiber.Ctx) error {
	c.Set("Content-Disposition", "inline")

	// Continue only for video formats that support byte batching.
	if _, ok := allowedVideoExt[strings.ToLower(filepath.Ext(c.Path()))]; !ok {
		return c.Next()
	}

	c.Set("Accept-Ranges", "bytes")

	rangeHeader := c.Get("Range")

	maxSize := int64(h.config.HTTP.RangeLimit)

	defaultRangeHeader := fmt.Sprintf("bytes=0-%d", maxSize)
	// TODO: Consider setting defaultRangeHeader for all errors.

	if rangeHeader == "" {
		h.logger.Error("range-limit: empty header")

		c.Request().Header.Set("Range", defaultRangeHeader)
		return c.Next()
	}

	const b = "bytes="
	if !strings.HasPrefix(rangeHeader, b) {
		h.logger.Error("range-limit: no bytes prefix", zap.String("rangeHeader", rangeHeader))
		return c.Next()
	}

	type httpRange struct {
		start, end string
	}

	var ranges []httpRange
	for _, ra := range strings.Split(rangeHeader[len(b):], ",") {
		ra = textproto.TrimString(ra)
		if ra == "" {
			continue
		}
		start, end, ok := strings.Cut(ra, "-")
		if !ok {
			h.logger.Error("range-limit: invalid range", zap.String("rangeHeader", rangeHeader))
			return c.Next()
		}
		start, end = textproto.TrimString(start), textproto.TrimString(end)
		var r httpRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file,
			// and we are dealing with <suffix-length>
			// which has to be a non-negative integer as per
			// RFC 7233 Section 2.1 "Byte-Ranges".
			if end == "" || end[0] == '-' {
				h.logger.Error("range-limit: invalid range without start", zap.String("rangeHeader", rangeHeader))
				return c.Next()
			}
			e, err := strconv.ParseInt(end, 10, 64)
			if e < 0 || err != nil {
				h.logger.Error("range-limit: invalid end without start", zap.Error(err), zap.String("rangeHeader", rangeHeader))
				return c.Next()
			}
			if maxSize < e {
				e = maxSize
			}
			r.end = strconv.Itoa(int(e))
		} else {
			r.start = start

			s, err := strconv.ParseInt(start, 10, 64)
			if err != nil || s < 0 {
				h.logger.Error("range-limit: invalid start", zap.Error(err), zap.String("rangeHeader", rangeHeader))
				return c.Next()
			}

			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.end = strconv.Itoa(int(s + maxSize))
			} else {
				e, err := strconv.ParseInt(end, 10, 64)
				if err != nil || s > e {
					h.logger.Error("range-limit: invalid end", zap.Error(err), zap.String("rangeHeader", rangeHeader))
					return c.Next()
				}
				if maxSize <= e-s {
					e = s + maxSize - 1
				}
				r.end = strconv.Itoa(int(e))
			}
		}
		ranges = append(ranges, r)
	}

	if len(ranges) != 0 {
		rawRanges := make([]string, 0, len(ranges))
		for _, r := range ranges {
			rawRanges = append(rawRanges, r.start+"-"+r.end)
		}

		c.Request().Header.Set("Range", "bytes="+strings.Join(rawRanges, ", "))
	}

	return c.Next()
}
