package v1

import (
	"academy/internal/api/apperrors"
	"academy/internal/model"
	"academy/internal/service/jwt"
	"academy/internal/service/upload"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (h *V1Handler) CreateLesson(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.CreateLessonRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if moduleNameLimit < utf8.RuneCountInString(req.ModuleName) {
		return apperrors.BadRequest("module name exceeds the limit")
	}

	lesson, err := req.ToLesson()
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	product, err := h.productService.GetByID(c.Context(), req.ProductID, true)
	if err != nil {
		return apperrors.Internal("failed to get product", err)
	}

	if product.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	err = h.lessonService.Create(c.Context(), lesson, req.Index, product, req.ProductLevelID)
	if err != nil {
		return apperrors.Internal("failed to create lesson", err)
	}

	return c.JSON(fiber.Map{
		"lesson": lesson,
	})
}

func (h *V1Handler) GetLesson(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	lessonID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	lesson, err := h.validateLessonAccess(c.Context(), &claims, lessonID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"lesson": lesson,
	})
}

func (h *V1Handler) EditLesson(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	lessonID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	var req model.EditLessonRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if moduleNameLimit < utf8.RuneCountInString(req.ModuleName) {
		return apperrors.BadRequest("module name exceeds the limit")
	}

	lesson, err := h.lessonService.GetByID(c.Context(), lessonID, uuid.Nil)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if err := h.checkProduct(c.Context(), claims.MiniAppID, lesson.ProductID); err != nil {
		return err
	}

	isChanged, err := req.UpdateLesson(lesson)
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if isChanged {
		err = h.lessonService.Update(c.Context(), lesson)
		if err != nil {
			return apperrors.Internal("failed to update lesson", err)
		}
	}

	return c.JSON(fiber.Map{
		"lesson": lesson,
	})
}

func (h *V1Handler) DeleteLesson(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionProductsControl) {
		return apperrors.Unauthorized("user is not permitted")
	}

	lessonID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if lessonID == uuid.Nil {
		return apperrors.BadRequest("invalid lesson id")
	}

	lesson, err := h.lessonService.GetByID(c.Context(), lessonID, uuid.Nil)
	if err != nil {
		return apperrors.Internal("error while getting the lesson", err)
	}

	product, err := h.productService.GetByID(c.Context(), lesson.ProductID, true)
	if err != nil {
		return apperrors.Internal("failed to get product", err)
	}

	if product.MiniAppID != claims.MiniAppID {
		return apperrors.Unauthorized("user is not permitted")
	}

	err = h.lessonService.Delete(c.Context(), lessonID, product)
	if err != nil {
		return apperrors.Internal("error while deleting the lesson", err)
	}

	pathToDelete := upload.MaterialFilePath{
		MiniAppID: claims.MiniAppID,
		ProductID: lesson.ProductID,
		LessonID:  lesson.ID,
	}
	err = h.uploadService.Delete(pathToDelete.String())
	if err != nil {
		h.logger.Error("error while deleting lesson", zap.String("err", err.Error()))
	}

	return nil
}

func (h *V1Handler) SubmitLesson(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if claims.IsOwner || claims.IsMod {
		return apperrors.Unauthorized("only students can submit the lesson")
	}

	lessonID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if lessonID == uuid.Nil {
		return apperrors.BadRequest("invalid lesson id")
	}

	lesson, err := h.validateLessonAccess(c.Context(), &claims, lessonID)
	if err != nil {
		return err
	}

	homeworks := make([]*model.Material, 0)
	for _, m := range lesson.Materials {
		if m.Category != model.MaterialCategoryHomework {
			continue
		}
		homeworks = append(homeworks, m)
	}

	if len(homeworks) == 0 ||
		(homeworks[0].ContentType != model.MaterialTypeQuiz &&
			homeworks[0].ContentType != model.MaterialTypeOpenQuestion) {

		lessonProgress := model.NewLessonProgressFromEmptyHomework(claims.UserID, lessonID)

		err := h.lessonProgressService.CreateOrUpdate(c.Context(), lessonProgress)
		if err != nil {
			return apperrors.Internal("error while creating progress", err)
		}

		return c.JSON(model.LessonSubmitionResponce{
			LessonResult: lessonProgress,
		})
	}

	mpForm, err := c.MultipartForm()
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	submits := mpForm.Value["submit"]
	if len(submits) != 1 {
		return apperrors.BadRequest("submit not provided")
	}

	var req model.LessonSubmitionRequest
	if err := json.Unmarshal([]byte(submits[0]), &req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}
	if submitLessonLinksLimit < len(req.Links) {
		return apperrors.BadRequest("number of links exceeds the limit")
	}

	var lessonProgress *model.LessonProgress

	if homeworks[0].ContentType == model.MaterialTypeQuiz {
		if len(req.QuizAnswers) == 0 {
			return apperrors.BadRequest("no quiz answers provided")
		}

		var quizMetadata model.QuizMetadata
		err := json.Unmarshal(homeworks[0].Metadata, &quizMetadata)
		if err != nil {
			return apperrors.Internal("lesson include invalid homework", err)
		}

		var quizAnswers model.QuizHiddenMetadata
		err = json.Unmarshal(homeworks[0].HiddenMetadata, &quizAnswers)
		if err != nil {
			return apperrors.Internal("lesson include invalid homework", err)
		}

		quizResult, score, err := quizMetadata.ToQuizResults(&quizAnswers, req.QuizAnswers)
		if err != nil {
			return apperrors.BadRequest("error while calculating the result", err)
		}

		var data model.LessonProgressData
		data.QuizResults = quizResult

		progressData, err := json.Marshal(data)
		if err != nil {
			return apperrors.Internal("error while saving the result", err)
		}

		lessonProgress = model.NewLessonProgressFromQuiz(
			claims.UserID,
			lessonID,
			progressData,
			score,
		)

		err = h.lessonProgressService.CreateOrUpdate(c.Context(), lessonProgress)
		if err != nil {
			return apperrors.Internal("error while creating progress", err)
		}
	}

	var isUpdated bool
	var newFiles []string
	var oldFiles []string
	defer func() {
		h.flushFiles(isUpdated, newFiles, oldFiles)
	}()

	if homeworks[0].ContentType == model.MaterialTypeOpenQuestion {
		if req.Text == "" {
			return apperrors.BadRequest("answer not provided")
		}

		var openQuestionMetadata model.OpenQuestionMetadata
		err := json.Unmarshal(homeworks[0].Metadata, &openQuestionMetadata)
		if err != nil {
			return apperrors.Internal("lesson include invalid homework", err)
		}

		var progressData model.LessonProgressData
		progressData.OpenAnswer = req.Text
		progressData.Links = req.Links

		var totalFilesSize int64
		if openQuestionMetadata.AllowFileAnswer {
			files := mpForm.File["file"]
			if submitLessonFilesLimit < len(files) {
				return apperrors.BadRequest("number of files exceeds the limit")
			}
			for _, f := range files {
				fileExt := strings.ToLower(filepath.Ext(f.Filename))

				if _, ok := allowedMaterialExt[fileExt]; !ok {
					return apperrors.BadRequest("this file extention is not allowed")
				}

				if submitLessonFileSizeLimit < f.Size {
					return apperrors.BadRequest("size of file exceeds the limit")
				}

				file, err := f.Open()
				if err != nil {
					return apperrors.BadRequest("error while opening file", err)
				}

				userPath := upload.MaterialFilePath{
					MiniAppID: claims.MiniAppID,
					ProductID: lesson.ProductID,
					LessonID:  lesson.ID,
					UserID:    claims.UserID,
				}

				fileURL, fileSize, err := h.uploadService.Upload(userPath.String(), file, fileExt)
				if err != nil {
					return apperrors.BadRequest("error while uploading file", err)
				}

				newFiles = append(newFiles, fileURL)

				progressData.FilesMetadata = append(progressData.FilesMetadata, &model.FileMetadata{
					Filename:         fileURL,
					OriginalFilename: f.Filename,
					Size:             fileSize,
				})
				totalFilesSize += fileSize
			}
		}

		data, err := json.Marshal(progressData)
		if err != nil {
			return apperrors.Internal("error while saving the result", err)
		}

		lessonProgress = model.NewLessonProgressFromOpenQuestion(
			claims.UserID,
			lessonID,
			data, totalFilesSize,
		)

		err = h.lessonProgressService.CreateOrUpdate(c.Context(), lessonProgress)
		if err != nil {
			return apperrors.Internal("error while creating progress", err)
		}
	}

	isUpdated = true

	return c.JSON(model.LessonSubmitionResponce{
		LessonResult: lessonProgress,
	})
}

func (h *V1Handler) SubmitLessonQuestion(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	lessonID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	lesson, err := h.validateLessonAccess(c.Context(), &claims, lessonID)
	if err != nil {
		return err
	}

	var homework *model.Material
	for _, m := range lesson.Materials {
		if m.Category != model.MaterialCategoryHomework {
			continue
		}

		homework = m
		break
	}

	if homework == nil || homework.ContentType != model.MaterialTypeQuiz {
		return apperrors.BadRequest("partial submition not suported for this lesson", err)
	}

	var req model.QuestionSubmitionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if len(req.QuestionAnswer) == 0 {
		return apperrors.BadRequest("no question answers provided")
	}

	var quizMetadata model.QuizMetadata
	err = json.Unmarshal(homework.Metadata, &quizMetadata)
	if err != nil {
		return apperrors.Internal("lesson include invalid homework", err)
	}

	var quizAnswers model.QuizHiddenMetadata
	err = json.Unmarshal(homework.HiddenMetadata, &quizAnswers)
	if err != nil {
		return apperrors.Internal("lesson include invalid homework", err)
	}

	questionResult, err := quizMetadata.ToQuestionResult(&quizAnswers, req.QuestionIndex, req.QuestionAnswer)
	if err != nil {
		return apperrors.BadRequest("error while calculating the result", err)
	}

	return c.JSON(model.QuestionSubmitionResponce{
		QuestionResult: questionResult,
	})
}

func (h *V1Handler) FeedbackHomework(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(jwt.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !h.isPermitted(c.Context(), &claims, model.PermissionStudentInteraction) {
		return apperrors.Unauthorized("user is not permitted")
	}

	var req model.FeedbackHomeworkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data", err)
	}

	if err := h.checkLesson(c.Context(), claims.MiniAppID, req.LessonID); err != nil {
		return err
	}

	err := h.lessonProgressService.FeedbackHomework(c.Context(), &req)
	if err != nil {
		return apperrors.Internal("error while getting homework by product", err)
	}

	return nil
}

// validateLessonAccess checks following rules:
// 1. Is lesson exists.
// 2. Is lesson related to user's mini-app.
// 3. Is product active.
// 4. Is user deleted from access the product.
// 5. Is user completed previous lesson (for lesson with set dependence).
// 6. Is lesson active for access.
// 7. Is lesson released and accessible.
// 8. Is lesson paid by the student (for paid lesson) or unlocked by invite.
func (h *V1Handler) validateLessonAccess(
	ctx context.Context,
	claims *jwt.TokenClaims,
	lessonID uuid.UUID,
) (*model.Lesson, error) {

	lesson, err := h.lessonService.GetByID(ctx, lessonID, claims.UserID)
	if err != nil {
		return nil, apperrors.Internal("failed to get lesson", err)
	}

	product, err := h.productService.GetByID(ctx, lesson.ProductID, false)
	if err != nil {
		return nil, apperrors.Internal("failed to get product", err)
	}

	if product.MiniAppID != claims.MiniAppID {
		return nil, apperrors.Unauthorized("user is not permitted")
	}

	if h.isPermitted(ctx, claims, model.PermissionProductsControl) {
		return lesson, nil
	}

	if !product.IsActive {
		return nil, apperrors.BadRequest("product not accessible")
	}

	// Do not check product.ReleaseDate because now it is included in payments
	// that checked in h.lessonService.IsLessonUnlocked().

	productAccess := model.NewProductAccess(claims.UserID, product.ID)
	productAccess, err = h.productService.CheckProductAccess(ctx, productAccess)
	if err != nil {
		return nil, apperrors.Internal("failed to get product access", err)
	}
	if productAccess.DeletedAt != nil {
		return nil, apperrors.Unauthorized("user deleted from accessing the product")
	}

	if lesson.PreviousLessonID != uuid.Nil {
		if len(lesson.PrevLessonProgress) == 0 ||
			lesson.PrevLessonProgress[0].Status != model.LessonProgressStatusAccepted {

			return nil, apperrors.BadRequest("complete previous lesson first")
		}
	}

	if !lesson.IsActive {
		return nil, apperrors.BadRequest("lesson not found")
	}

	if lesson.ReleaseDate.Valid {
		err := isAccessible(lesson.ReleaseDate, lesson.AccessTime)
		if err != nil {
			return nil, apperrors.Unauthorized("lesson not accessible", err)
		}
	}

	isUnlocked, err := h.lessonService.IsLessonUnlocked(ctx, lessonID, claims.UserID)
	if err != nil {
		return nil, apperrors.Internal("failed to check if lesson is unlocked", err)
	}

	if !isUnlocked {
		return nil, apperrors.Unauthorized("the lesson is locked for the user")
	}

	return lesson, nil
}
