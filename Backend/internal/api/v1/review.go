package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/model"
	"academy/internal/service/jwt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (h *V1Handler) ReviewLesson(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	lessonID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if lessonID == uuid.Nil {
		return apperrors.BadRequest("invalid lesson id")
	}
	if err := h.checkLesson(c.Context(), claims.MiniAppID, lessonID); err != nil {
		return err
	}

	var req model.ReviewRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	review := model.NewLessonReview(claims.UserID, lessonID, &req)

	err = h.reviewService.Create(c.Context(), review)
	if err != nil {
		return apperrors.Internal("failed to create lesson review", err)
	}

	return nil
}
