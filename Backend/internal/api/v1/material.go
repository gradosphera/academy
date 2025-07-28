package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/model"
	"academy/internal/service/jwt"
	"academy/internal/service/upload"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (h *V1Handler) CreateMaterial(c fiber.Ctx) error {
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

	materials := mpForm.Value["material"]
	if len(materials) != 1 {
		return apperrors.BadRequest("materials not provided")
	}

	var req model.CreateMaterialRequest
	if err := json.Unmarshal([]byte(materials[0]), &req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	switch req.Category {
	case model.MaterialCategoryHomework:
		return apperrors.BadRequest("use a separate endpoint to upload homework")

	case model.MaterialCategorySlides:
		return apperrors.BadRequest("use a separate endpoint to upload slides")

	case model.MaterialCategoryLessonContent:
		if lessonTitleLimit < utf8.RuneCountInString(req.Title) {
			return apperrors.BadRequest("lesson title exceeds the limit")
		}

		switch req.ContentType {
		case model.MaterialTypeVideo, model.MaterialTypeCircleVideo:
			if videoLessonDescriptionLimit < utf8.RuneCountInString(req.Description) {
				return apperrors.BadRequest("lesson description exceeds the limit")
			}
		case model.MaterialTypeAudio:
			if audioLessonDescriptionLimit < utf8.RuneCountInString(req.Description) {
				return apperrors.BadRequest("lesson description exceeds the limit")
			}
		}
	case model.MaterialCategoryMaterials:
		if materialTitleLimit < utf8.RuneCountInString(req.Title) {
			return apperrors.BadRequest("material title exceeds the limit")
		}
	case model.MaterialCategoryBonus:
		if bonusMaterialTitleLimit < utf8.RuneCountInString(req.Title) {
			return apperrors.BadRequest("bonus material title exceeds the limit")
		}
	}

	files := mpForm.File["file"]
	if 1 < len(files) {
		return apperrors.BadRequest("only one file per material", err)
	}

	var materialPath string
	switch {
	case req.LessonID != uuid.Nil:
		lesson, err := h.lessonService.GetByID(c.Context(), req.LessonID, uuid.Nil)
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
		materialPath = lessonPath.String()
	case req.MiniAppID != uuid.Nil:
		if req.MiniAppID != claims.MiniAppID {
			return apperrors.Unauthorized("user is not permitted")
		}
		miniAppPath := upload.MaterialFilePath{
			MiniAppID: claims.MiniAppID,
		}
		materialPath = miniAppPath.String()
	case req.ProductLevelID != uuid.Nil:
		productLevel, err := h.productLevelService.GetByID(c.Context(), req.ProductLevelID)
		if err != nil {
			return apperrors.Internal("failed to get pluduct level", err)
		}
		if err := h.checkProduct(c.Context(), claims.MiniAppID, productLevel.ProductID); err != nil {
			return err
		}
		levelPath := upload.MaterialFilePath{
			MiniAppID:      claims.MiniAppID,
			ProductID:      productLevel.ProductID,
			ProductLevelID: productLevel.ID,
		}
		materialPath = levelPath.String()
	default:
		return apperrors.BadRequest("invalid material ids")
	}

	var isUpdated bool
	var newFiles []string
	var oldFiles []string
	defer func() {
		h.flushFiles(isUpdated, newFiles, oldFiles)
	}()

	var originalFilename string
	var filename string
	var fileSize int64
	if 0 < len(files) {
		fileExt := strings.ToLower(filepath.Ext(files[0].Filename))

		fileContentType := files[0].Header.Get("Content-Type")
		f, err := files[0].Open()
		if err != nil {
			return apperrors.BadRequest("error while opening file", err)
		}
		originalFilename = files[0].Filename

		filename, fileSize, err = h.uploadService.Upload(
			materialPath, f, fileExt)
		if err != nil {
			return apperrors.BadRequest("error while uploading file", err)
		}
		newFiles = append(newFiles, filename)

		if err := checkMaterialFile(
			req.Category, req.ContentType,
			fileSize, fileExt, fileContentType,
		); err != nil {
			return err
		}
	}

	material, err := req.ToMaterial(originalFilename, filename, fileSize)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	err = h.materialService.Create(c.Context(), material)
	if err != nil {
		return apperrors.Internal("failed to create material", err)
	}

	isUpdated = true

	return c.JSON(fiber.Map{
		"material": material,
	})
}

func (h *V1Handler) CreateHomework(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.CreateHomeworkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if err := h.checkLesson(c.Context(), claims.MiniAppID, req.LessonID); err != nil {
		return err
	}

	material, err := req.ToMaterial(homeworkQuestionLimit, homeworkQuestionLimit)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	err = h.materialService.Create(c.Context(), material)
	if err != nil {
		return apperrors.Internal("failed to create material", err)
	}

	return c.JSON(fiber.Map{
		"material": material,
	})
}

func (h *V1Handler) EditHomework(c fiber.Ctx) error {
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

	if material.Category != model.MaterialCategoryHomework {
		return apperrors.BadRequest("the material is not homework")
	}

	var req model.EditHomeworkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if err := h.checkLesson(c.Context(), claims.MiniAppID, material.LessonID); err != nil {
		return err
	}

	isChanged, err := req.UpdateMaterial(material, homeworkQuestionLimit, homeworkQuestionLimit)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if isChanged {
		err = h.materialService.Update(c.Context(), material)
		if err != nil {
			return apperrors.Internal("failed to update material", err)
		}
	}

	return c.JSON(fiber.Map{
		"material": material,
	})
}

func (h *V1Handler) EditMaterial(c fiber.Ctx) error {
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

	mpForm, err := c.MultipartForm()
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	materials := mpForm.Value["material"]
	if len(materials) != 1 {
		return apperrors.BadRequest("materials not provided")
	}

	var req model.EditMaterialRequest
	if err := json.Unmarshal([]byte(materials[0]), &req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	switch material.Category {
	case model.MaterialCategoryLessonContent:
		if lessonTitleLimit < utf8.RuneCountInString(req.Title) {
			return apperrors.BadRequest("lesson title exceeds the limit")
		}
		switch req.ContentType {
		case model.MaterialTypeVideo, model.MaterialTypeCircleVideo:
			if videoLessonDescriptionLimit < utf8.RuneCountInString(req.Description) {
				return apperrors.BadRequest("lesson description exceeds the limit")
			}
		case model.MaterialTypeAudio:
			if audioLessonDescriptionLimit < utf8.RuneCountInString(req.Description) {
				return apperrors.BadRequest("lesson description exceeds the limit")
			}
		}
	case model.MaterialCategoryMaterials:
		if materialTitleLimit < utf8.RuneCountInString(req.Title) {
			return apperrors.BadRequest("material title exceeds the limit")
		}
	case model.MaterialCategoryBonus:
		if bonusMaterialTitleLimit < utf8.RuneCountInString(req.Title) {
			return apperrors.BadRequest("bonus material title exceeds the limit")
		}
	}

	isChanged, err := req.UpdateMaterial(material)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	var materialPath string
	switch {
	case material.LessonID != uuid.Nil:
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
		materialPath = lessonPath.String()
	case material.MiniAppID != uuid.Nil:
		if material.MiniAppID != claims.MiniAppID {
			return apperrors.Unauthorized("user is not permitted")
		}
		miniAppPath := upload.MaterialFilePath{
			MiniAppID: claims.MiniAppID,
		}
		materialPath = miniAppPath.String()
	case material.ProductLevelID != uuid.Nil:
		productLevel, err := h.productLevelService.GetByID(c.Context(), material.ProductLevelID)
		if err != nil {
			return apperrors.Internal("failed to get pluduct level", err)
		}
		if err := h.checkProduct(c.Context(), claims.MiniAppID, productLevel.ProductID); err != nil {
			return err
		}
		levelPath := upload.MaterialFilePath{
			MiniAppID:      claims.MiniAppID,
			ProductID:      productLevel.ProductID,
			ProductLevelID: productLevel.ID,
		}
		materialPath = levelPath.String()
	default:
		return apperrors.BadRequest("invalid material ids")
	}

	var isUpdated bool
	var newFiles []string
	var oldFiles []string
	defer func() {
		h.flushFiles(isUpdated, newFiles, oldFiles)
	}()

	if files := mpForm.File["file"]; 0 < len(files) {
		if len(material.Metadata) != 0 {
			return apperrors.BadRequest("this material include metadata and do not accept new files")
		}

		fileExt := strings.ToLower(filepath.Ext(files[0].Filename))

		fileContentType := files[0].Header.Get("Content-Type")
		switch material.ContentType {
		case model.MaterialTypeQuiz, model.MaterialTypeOpenQuestion:
			return apperrors.BadRequest("this content type do not support files upload", err)
		}
		f, err := files[0].Open()
		if err != nil {
			return apperrors.BadRequest("error while opening file", err)
		}
		filename, fileSize, err := h.uploadService.Upload(
			materialPath, f, fileExt)
		if err != nil {
			return apperrors.BadRequest("error while uploading file", err)
		}

		oldFiles = append(oldFiles, material.Filename)
		newFiles = append(newFiles, filename)

		material.Filename = filename
		material.Size = fileSize

		isChanged = true

		if err := checkMaterialFile(
			material.Category, req.ContentType,
			fileSize, fileExt, fileContentType,
		); err != nil {
			return err
		}
	}

	if isChanged {
		err = h.materialService.Update(c.Context(), material)
		if err != nil {
			return apperrors.Internal("failed to update material", err)
		}
	}

	isUpdated = true

	return c.JSON(fiber.Map{
		"material": material,
	})
}

func (h *V1Handler) GetMaterialToken(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	material, err := h.materialService.GetByID(c.Context(), materialID)
	if err != nil {
		return apperrors.Internal("error while getting the material", err)
	}

	if material.LessonID == uuid.Nil ||
		material.Category != model.MaterialCategoryLessonContent ||
		len(material.Metadata) == 0 {

		return apperrors.BadRequest("this material do not support generating token")
	}

	if _, err := h.validateLessonAccess(c.Context(), &claims, material.LessonID); err != nil {
		return apperrors.Unauthorized("material access restricted", err)
	}

	var metadata model.MuxVideoMetadata
	if err := json.Unmarshal(material.Metadata, &metadata); err != nil {
		return apperrors.Internal("failed to load mux metadata", err)
	}

	token, err := h.uploadService.MuxSignPrivateVideo(metadata.PlaybackID)
	if err != nil {
		return apperrors.Internal("failed to sign video", err)
	}

	return c.JSON(fiber.Map{
		"playback_id": metadata.PlaybackID,
		"token":       token,
	})
}

func (h *V1Handler) DeleteMaterial(c fiber.Ctx) error {
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
		return apperrors.Internal("error while getting the material", err)
	}

	switch {
	case material.LessonID != uuid.Nil:
		if err := h.checkLesson(c.Context(), claims.MiniAppID, material.LessonID); err != nil {
			return err
		}
	case material.MiniAppID != uuid.Nil:
		if material.MiniAppID != claims.MiniAppID {
			return apperrors.Unauthorized("user is not permitted")
		}
	case material.ProductLevelID != uuid.Nil:
		productLevel, err := h.productLevelService.GetByID(c.Context(), material.ProductLevelID)
		if err != nil {
			return apperrors.Internal("failed to get pluduct level", err)
		}
		if err := h.checkProduct(c.Context(), claims.MiniAppID, productLevel.ProductID); err != nil {
			return err
		}
	default:
		return apperrors.Internal("unexpected case")
	}

	err = h.materialService.Delete(c.Context(), materialID)
	if err != nil {
		return apperrors.Internal("error while deleting the material", err)
	}

	h.flushFiles(true, nil, []string{material.Filename})

	return nil
}

func checkMaterialFile(
	category model.MaterialCategory,
	contentType model.MaterialType,
	fileSize int64, fileExt string, fileContentType string,
) error {

	fileExt = strings.ToLower(fileExt)

	switch category {
	case model.MaterialCategoryPrivacyPolicy:
		if privacyPolicySizeLimit < fileSize {
			return apperrors.BadRequest("privacy policy size exceeds the limit")
		}
	case model.MaterialCategoryTOS:
		if tosSizeLimit < fileSize {
			return apperrors.BadRequest("TOS size exceeds the limit")
		}
	case model.MaterialCategoryLessonContent:
		switch contentType {
		case model.MaterialTypeVideo, model.MaterialTypeCircleVideo:
			if _, ok := allowedVideoExt[fileExt]; !ok {
				return apperrors.BadRequest("this file extention is not allowed", fmt.Errorf("ext: %q", fileExt))
			}
			if videoLessonSizeLimit < fileSize {
				return apperrors.BadRequest("lesson size exceeds the limit")
			}
		case model.MaterialTypeAudio:
			if _, ok := allowedAudioExt[fileExt]; !ok {
				return apperrors.BadRequest("this file extention is not allowed", fmt.Errorf("ext: %q", fileExt))
			}
			if audioLessonSizeLimit < fileSize {
				return apperrors.BadRequest("lesson size exceeds the limit")
			}
		}
	case model.MaterialCategoryLessonCover:
		if lessonCoverSizeLimit < fileSize {
			return apperrors.BadRequest("lesson cover exceeds the limit")
		}
		if !isPictureAllowed(fileContentType, fileExt) {
			return apperrors.BadRequest("lesson cover file extention is not allowed", fmt.Errorf("ext: %q", fileExt))
		}
	case model.MaterialCategoryMaterials:
		if materialSizeLimit < fileSize {
			return apperrors.BadRequest("lesson material exceeds the limit")
		}
		if _, ok := allowedMaterialExt[fileExt]; !ok {
			return apperrors.BadRequest("this file extention is not allowed", fmt.Errorf("ext: %q", fileExt))
		}
	case model.MaterialCategoryBonus:
		if bonusMaterialSizeLimit < fileSize {
			return apperrors.BadRequest("bonus material exceeds the limit")
		}
		if _, ok := allowedMaterialExt[fileExt]; !ok {
			return apperrors.BadRequest("this file extention is not allowed", fmt.Errorf("ext: %q", fileExt))
		}
	}

	return nil
}
