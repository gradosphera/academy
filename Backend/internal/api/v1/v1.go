package v1

import (
	"academy/internal/config"
	"academy/internal/model"
	"academy/internal/service"
	"academy/internal/service/jwt"
	"academy/internal/service/security"
	"academy/internal/service/telegram"
	"academy/internal/service/upload"
	"academy/internal/types"
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"go.uber.org/zap"
)

const uploadPath = "/upload/"

type V1Handler struct {
	config *config.Config
	logger *zap.Logger

	userService           *service.UserService
	miniAppService        *service.MiniAppService
	productService        *service.ProductService
	lessonService         *service.LessonService
	materialService       *service.MaterialService
	chunkService          *service.ChunkService
	lessonProgressService *service.LessonProgressService
	productLevelService   *service.ProductLevelService
	reviewService         *service.ReviewService

	jwtService      *service.JWTService
	telegramService *telegram.Service
	paymentService  *service.PaymentService
	uploadService   *upload.Service
	securityService *security.Service
}

func NewV1Handler(
	cfg *config.Config,
	logger *zap.Logger,

	userService *service.UserService,
	miniAppService *service.MiniAppService,
	productService *service.ProductService,
	lessonService *service.LessonService,
	materialService *service.MaterialService,
	chunkService *service.ChunkService,

	lessonProgressService *service.LessonProgressService,
	productLevelService *service.ProductLevelService,
	reviewService *service.ReviewService,

	jwtService *service.JWTService,
	tgService *telegram.Service,
	uploadService *upload.Service,
	paymentService *service.PaymentService,
	securityService *security.Service,
) *V1Handler {

	return &V1Handler{
		config: cfg,
		logger: logger,

		userService:     userService,
		miniAppService:  miniAppService,
		productService:  productService,
		lessonService:   lessonService,
		materialService: materialService,
		chunkService:    chunkService,

		lessonProgressService: lessonProgressService,
		productLevelService:   productLevelService,
		reviewService:         reviewService,

		jwtService:      jwtService,
		telegramService: tgService,
		uploadService:   uploadService,
		paymentService:  paymentService,
		securityService: securityService,
	}
}

func (h *V1Handler) RegisterRoutes(app *fiber.App) {
	v1Group := app.Group("/v1")

	authGroup := v1Group.Group("/auth")
	authGroup.Post("/admin/signin", h.AdminSignIn)
	authGroup.Post("/mod/list", h.ListMiniApps)
	authGroup.Post("/mod/signin", h.ModSignIn)
	authGroup.Post("/signin", h.SignIn)
	authGroup.Post("/signin/invite", h.SignInWithInvite)
	authGroup.Post("/refresh", h.RefreshTokens)

	userGroup := v1Group.Group("/user")
	userGroup.Use(h.JWTAuthMiddleware)
	userGroup.Get("/me", h.GetUser)
	userGroup.Post("/edit", h.EditUser)
	userGroup.Post("/homeworks", h.UserHomeworks)
	userGroup.Get("/:id/stats", h.StudentStats)
	userGroup.Post("/:id/ban", h.BanUser)
	userGroup.Post("/:id/unban", h.UnbanUser)
	userGroup.Post("/banlist", h.ListBannedUser)
	userGroup.Post("/:id/levelup", h.LevelUpUser)
	userGroup.Post("/:id/levels", h.UserLevels)

	modGroup := v1Group.Group("/mod")
	modGroup.Use(h.JWTAuthMiddleware)
	modGroup.Post("/invite", h.CreateModInvite)
	modGroup.Delete("/invite/:id", h.DeleteModInvite)
	modGroup.Post("/invite/:id/edit", h.EditModInvite)
	modGroup.Post("/invites", h.ModInvites)
	modGroup.Get("/permissions", h.ModPermissions)

	v1Group.Post("/app", h.CreateMiniApp)
	appGroup := v1Group.Group("/app")
	appGroup.Use(h.JWTAuthMiddleware)
	appGroup.Get("/", h.GetMiniApp)
	appGroup.Delete("/", h.DeleteMiniApp)
	appGroup.Post("/edit/account", h.EditMiniAppAccount)
	appGroup.Post("/edit/branding", h.EditMiniAppBranding)
	appGroup.Post("/edit/analytics", h.EditMiniAppAnalytics)
	appGroup.Get("/archive", h.ArchiveMiniApp)
	appGroup.Get("/unarchive", h.UnarchiveMiniApp)
	appGroup.Post("/slides", h.EditSlides)
	appGroup.Post("/analytics", h.Analytics)
	appGroup.Get("/info", h.Info)
	appGroup.Get("/payment_metadata", h.PaymentMetadata)

	appGroup.Post("/product", h.CreateProduct)
	appGroup.Get("/product/:id", h.GetProduct)
	appGroup.Post("/product/:id/edit", h.EditProduct)
	appGroup.Post("/product/:id/reorder/lessons", h.ReorderProductLessons)
	appGroup.Post("/product/:id/reorder/levels", h.ReorderProductLevels)
	appGroup.Delete("/product/:id", h.DeleteProduct)

	appGroup.Post("/product/:id/invites", h.ProductInvites)
	appGroup.Post("/product/:id/homeworks", h.ProductHomeworks)
	appGroup.Get("/product/:id/feedback", h.ProductFeedback)
	appGroup.Get("/product/:id/students", h.ProductStudents)
	appGroup.Post("/product/:id/students/export/excel", h.ExportProductStudents)

	appGroup.Post("/lesson", h.CreateLesson)
	appGroup.Get("/lesson/:id", h.GetLesson)
	appGroup.Post("/lesson/:id/edit", h.EditLesson)
	appGroup.Post("/lesson/:id/submit", h.SubmitLesson)
	appGroup.Post("/lesson/:id/submit/question", h.SubmitLessonQuestion)
	appGroup.Post("/lesson/:id/review", h.ReviewLesson)
	appGroup.Delete("/lesson/:id", h.DeleteLesson)
	appGroup.Post("/homework/feedback", h.FeedbackHomework)

	appGroup.Post("/homework", h.CreateHomework)
	appGroup.Post("/homework/:id/edit", h.EditHomework)

	appGroup.Post("/material", h.CreateMaterial)
	appGroup.Post("/material/:id/edit", h.EditMaterial)
	appGroup.Get("/material/:id/token", h.GetMaterialToken)
	appGroup.Post("/material/:id/chunk/:index", h.AddChunk)
	appGroup.Post("/material/:id/chunks/submit", h.SubmitChunks)
	appGroup.Post("/material/:id/chunks/clear", h.ClearChunks)
	appGroup.Delete("/material/:id", h.DeleteMaterial)

	appGroup.Post("/level", h.CreateProductLevel)
	appGroup.Post("/level/:id/edit", h.EditProductLevel)
	appGroup.Get("/level/:id/invite", h.CreateProductLevelInvite)
	appGroup.Delete("/level/:id", h.DeleteProductLevel)
	appGroup.Get("/level/:id/buy/ton", h.BuyProductLevelWithTON)
	appGroup.Get("/level/:id/buy/wayforpay", h.BuyProductLevelWithWayForPay)

	appGroup.Get("/payment/:id", h.GetPayment)
	appGroup.Post("/payments", h.GetPayments)
	appGroup.Post("/students/payments", h.GetStudentsPayments)
	appGroup.Post("/students/payments/export/excel", h.ExportStudentsPayments)

	v1Group.Post("/wayforpay/update", h.WayForPayUpdate)

	v1Group.Get("/static/*", static.New("./resources/static"))
	v1Group.Get("/swagger/*", static.New("./resources/swagger"))

	var staticConfig = static.Config{
		IndexNames:    []string{"index.html"},
		CacheDuration: 10 * time.Second,
		ByteRange:     true,
		Compress:      true,
	}

	app.Get(uploadPath+"*", static.New(
		h.config.App.UploadDirectory,
		staticConfig,
	),
		h.MaterialAuthMiddleware,
		// h.HandleMaterialHeaders,
	)
}

func (h *V1Handler) isPermitted(
	ctx context.Context,
	claims *jwt.TokenClaims,
	permissionName ...model.PermissionName,
) bool {

	if claims.IsOwner {
		return true
	}

	if !claims.IsMod {
		return false
	}

	isPermitted, err := h.miniAppService.CheckPermission(ctx, claims.UserID, permissionName...)
	if err != nil {
		h.logger.Error("failed to check permissions", zap.Error(err))
		return false
	}

	return isPermitted
}

// flushFiles takes list of newly uploaded files and files that gets replaced
// so that one of the collection gets deleted depending on success of operation.
func (h *V1Handler) flushFiles(isUpdated bool, newFiles []string, oldFiles []string) {
	filesToDelete := newFiles
	if isUpdated {
		filesToDelete = oldFiles
	}

	for _, filename := range filesToDelete {
		if filename == "" {
			continue
		}

		err := h.uploadService.Delete(filename)
		if err != nil {
			h.logger.Error("error while deleting file", zap.String("err", err.Error()))
		}
	}
}

func isAccessible(releaseDate types.Time, accessTime types.Interval) error {
	timeNow := time.Now()

	if timeNow.Before(releaseDate.Time) {
		return errors.New("content not released yet")
	}

	if accessTime.Valid {
		expTime := releaseDate.Time.AddDate(
			0,
			int(accessTime.Months),
			int(accessTime.Days),
		).Add(
			time.Duration(accessTime.Microseconds * 1000),
		)

		if expTime.Before(timeNow) {
			return errors.New("content not accessible anymore")
		}
	}

	return nil
}

func validateLimit(limit uint) uint {
	if limit == 0 {
		return defaultLimit
	}

	if maxLimit < limit {
		return maxLimit
	}

	return limit
}
