package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/model"
	"academy/internal/service/jwt"
	"academy/internal/service/upload"
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *V1Handler) AddChunk(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if materialID == uuid.Nil {
		return apperrors.BadRequest("invalid material id")
	}

	material, err := h.materialService.GetByID(c.Context(), materialID)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if material.MiniAppID != uuid.Nil {
		return apperrors.BadRequest("this type of material cant be chunked")
	}

	if len(material.Metadata) != 0 {
		return apperrors.BadRequest("this material include metadata and do not accept new files")
	}

	if err := h.checkLesson(c.Context(), claims.MiniAppID, material.LessonID); err != nil {
		return err
	}

	chunkIndex, err := strconv.Atoi(c.Params("index"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	mpForm, err := c.MultipartForm()
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	chunks := mpForm.Value["chunk"]
	if len(chunks) != 1 {
		return apperrors.BadRequest("chunks not provided")
	}

	var req model.CreateChunkRequest
	if err := json.Unmarshal([]byte(chunks[0]), &req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	files := mpForm.File["file"]
	if len(files) != 1 {
		return apperrors.BadRequest("one and only one file should be provided within a chunk")
	}

	f, err := files[0].Open()
	if err != nil {
		return apperrors.BadRequest("error while opening file", err)
	}
	fileSize, err := h.uploadService.AddChunk(materialID, int64(chunkIndex), f, req.Hashsum)
	if err != nil && errors.Is(err, upload.ErrHashNotMatch) {
		return apperrors.BadRequest(upload.ErrHashNotMatch.Error(), err)
	}
	if err != nil {
		return apperrors.BadRequest("error while uploading file", err)
	}

	chunk := model.NewChunk(claims.MiniAppID, materialID, int64(chunkIndex), fileSize)

	err = h.chunkService.Create(c.Context(), chunk)
	if err != nil {
		// Clear all chunks if failed to save new chunk (mostly due to size limit).

		err2 := h.uploadService.ClearChunks(materialID)
		if err2 != nil {
			h.logger.Error("error while deleting chunks files", zap.String("err", err2.Error()))
		}

		err2 = h.chunkService.Delete(c.Context(), materialID)
		if err2 != nil {
			h.logger.Error("error while deleting chunks", zap.String("err", err2.Error()))
		}

		return apperrors.Internal("failed to create chunk", err)
	}

	return c.JSON(fiber.Map{
		"size": fileSize,
	})
}

func (h *V1Handler) SubmitChunks(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if materialID == uuid.Nil {
		return apperrors.BadRequest("invalid material id")
	}

	material, err := h.materialService.GetByID(c.Context(), materialID)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if material.MiniAppID != uuid.Nil {
		return apperrors.BadRequest("this type of material cant be chunked")
	}

	if len(material.Metadata) != 0 {
		return apperrors.BadRequest("this material include metadata and do not accept new files")
	}

	var req model.SubmitChunksRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if err := req.Validate(); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	lesson, err := h.lessonService.GetByID(c.Context(), material.LessonID, uuid.Nil)
	if err != nil {
		return apperrors.BadRequest("error while getting lesson", err)
	}

	if err := h.checkProduct(c.Context(), claims.MiniAppID, lesson.ProductID); err != nil {
		return err
	}

	lessonPath := upload.MaterialFilePath{
		MiniAppID: claims.MiniAppID,
		ProductID: lesson.ProductID,
		LessonID:  lesson.ID,
	}

	var isUpdated bool
	var newFiles []string
	var oldFiles []string
	defer func() {
		h.flushFiles(isUpdated, newFiles, oldFiles)
	}()

	chunks, err := h.chunkService.GetByID(c.Context(), materialID)
	if err != nil {
		return apperrors.Internal("failed to get chunks", err)
	}
	if len(chunks) == 0 {
		return apperrors.BadRequest("no chunks was provided", err)
	}

	fileExt := strings.ToLower(filepath.Ext(req.OriginalFilename))

	filename, fileSize, err := h.uploadService.UploadChunks(
		c.Context(), lessonPath.String(), fileExt, chunks)
	if err != nil {
		return apperrors.Internal("failed to submit chunks", err)
	}

	oldFiles = append(oldFiles, material.Filename)
	newFiles = append(newFiles, filename)

	err = h.chunkService.Delete(c.Context(), materialID)
	if err != nil {
		return apperrors.Internal("failed to delete chunks", err)
	}

	material.OriginalFilename = req.OriginalFilename
	material.Filename = filename
	material.Size = fileSize
	material.UpdatedAt = time.Now().UTC()

	// Mark with pending status to compress later by cron-job.
	if _, ok := allowedVideoExt[fileExt]; ok {
		material.Status = model.MaterialStatusPendingCompressing
	}

	if req.Status != "" {
		material.Status = req.Status
	}

	if err := checkMaterialFile(
		material.Category, material.ContentType,
		fileSize, fileExt, "",
	); err != nil {
		return err
	}

	err = h.materialService.Update(c.Context(), material)
	if err != nil {
		return apperrors.Internal("error while submiting chunks", err)
	}

	isUpdated = true

	return c.JSON(fiber.Map{
		"material": material,
	})
}

func (h *V1Handler) ClearChunks(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	material, err := h.materialService.GetByID(c.Context(), materialID)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if err := h.checkLesson(c.Context(), claims.MiniAppID, material.LessonID); err != nil {
		return err
	}

	if err := h.chunkService.Delete(c.Context(), materialID); err != nil {
		return apperrors.Internal("failed to delete chunks", err)
	}

	if err := h.uploadService.ClearChunks(materialID); err != nil {
		h.logger.Error("failed to clear chunks", zap.String("err", err.Error()))
	}

	return nil
}
