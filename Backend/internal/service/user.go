package service

import (
	repo "academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/service/telegram"
	"academy/internal/service/upload"
	"academy/internal/storage/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type UserService struct {
	logger             *zap.Logger
	transactionManager *repo.TransactionManager

	miniAppRepository        *repository.MiniAppRepository
	userRepository           *repository.UserRepository
	paymentRepository        *repository.PaymentRepository
	lessonProgressRepository *repository.LessonProgressRepository

	uploadService   *upload.Service
	telegramService *telegram.Service
}

func NewUserService(
	logger *zap.Logger,
	transactionManager *repo.TransactionManager,

	miniAppRepository *repository.MiniAppRepository,
	userRepository *repository.UserRepository,
	paymentRepository *repository.PaymentRepository,
	lessonProgressRepository *repository.LessonProgressRepository,

	uploadService *upload.Service,
	telegramService *telegram.Service,
) *UserService {

	return &UserService{
		logger:             logger,
		transactionManager: transactionManager,

		miniAppRepository:        miniAppRepository,
		userRepository:           userRepository,
		paymentRepository:        paymentRepository,
		lessonProgressRepository: lessonProgressRepository,

		uploadService:   uploadService,
		telegramService: telegramService,
	}
}

func (s *UserService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.User, error) {

	user, err := s.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

func (s *UserService) SignInWithTelegramMiniApp(
	ctx context.Context,
	miniAppID uuid.UUID,
	initData *initdata.InitData,
	userRole model.UserRole,
) (*model.User, error) {

	user, err := s.userRepository.GetByTelegramID(ctx, initData.User.ID, miniAppID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by telegram id: %w", err)
	}

	// Keep old role if user still having it.
	if user != nil {
		if user.Role == model.UserRoleOwner && userRole != model.UserRoleOwner {
			miniApp, err := s.miniAppRepository.GetByID(ctx, miniAppID)
			if err != nil {
				return nil, fmt.Errorf("failed to get mini app: %w", err)
			}

			if miniApp.Owner.ID == user.ID {
				userRole = model.UserRoleOwner
			}
		}

		if user.Role == model.UserRoleModerator && userRole != model.UserRoleModerator {
			permissions, err := s.miniAppRepository.GetPermissions(ctx, user.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get permissions: %w", err)
			}

			if len(permissions) != 0 {
				userRole = model.UserRoleModerator
			}
		}
	}

	newUser := model.NewSignInWithTelegramMiniApp(initData, miniAppID, userRole)

	if user == nil {
		if initData.User.PhotoURL != "" && userRole == model.UserRoleStudent {
			avatarBytes, avatarExt, err := s.telegramService.DownloadAvatar(
				ctx, initData.User.PhotoURL)

			if err != nil {
				return nil, fmt.Errorf("failed to load telegram avatar: %w", err)
			}

			if avatarExt != ".svg" {
				materialPath := &upload.MaterialFilePath{
					MiniAppID: miniAppID,
				}
				newUser.Avatar, newUser.AvatarSize, err = s.uploadService.Upload(
					materialPath.String(), bytes.NewReader(avatarBytes), avatarExt)
				if err != nil {
					return nil, fmt.Errorf("failed to upload telegram avatar: %w", err)
				}
			}
		}

		err = s.userRepository.Create(ctx, newUser)
		if err != nil {
			if newUser.Avatar != "" {
				err2 := s.uploadService.Delete(newUser.Avatar)
				if err2 != nil {
					s.logger.Error("error while deleting file", zap.String("err", err2.Error()))
				}
			}

			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		return newUser, nil
	}

	if user.UpdateOnSignIn(newUser) {
		err = s.userRepository.Update(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, user *model.User) error {
	err := s.userRepository.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *UserService) StudentStats(
	ctx context.Context,
	miniAppID, userID uuid.UUID,
) ([]*model.StudentProductStats, error) {

	stats, err := s.userRepository.StudentStats(ctx, miniAppID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to student stats: %w", err)
	}

	return stats, nil
}

func (s *UserService) Ban(
	ctx context.Context,
	userID uuid.UUID, req *model.BanUserRequest,
) ([]string, error) {

	filesToDelete := make([]string, 0)
	err := s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err := s.userRepository.WithTx(tx).RestrictProductAccess(ctx, userID, req)
		if err != nil {
			return fmt.Errorf("failed to restrict user from accessing the product: %w", err)
		}

		progress, err := s.lessonProgressRepository.WithTx(tx).Delete(ctx, userID, req.ProductID)
		if err != nil {
			return fmt.Errorf("failed to delete user progress: %w", err)
		}

		for _, p := range progress {
			if p.Size == 0 {
				continue
			}
			var progressData model.LessonProgressData
			err := json.Unmarshal(p.Data, &progressData)
			if err != nil {
				s.logger.Error("failed to unmarshall lesson progress",
					zap.String("user_id", p.UserID.String()),
					zap.String("lesson_id", p.LessonID.String()),
					zap.Error(err),
				)
				continue
			}
			for _, f := range progressData.FilesMetadata {
				filesToDelete = append(filesToDelete, f.Filename)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return filesToDelete, nil
}

func (s *UserService) Unban(
	ctx context.Context,
	userID uuid.UUID, req *model.UnbanUserRequest,
) error {

	err := s.userRepository.AllowProductAccess(ctx, userID, req)
	if err != nil {
		return fmt.Errorf("failed to allow user access to product: %w", err)
	}

	return nil
}

func (s *UserService) ListBanned(
	ctx context.Context,
	miniAppID uuid.UUID, req *model.ListBannedUserRequest,
) ([]*model.User, int, error) {

	users, total, err := s.userRepository.ListRestrictedUsers(ctx, miniAppID, req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get banned users: %w", err)
	}

	return users, total, nil
}

func (s *UserService) LevelUp(
	ctx context.Context,
	userID uuid.UUID,
	product *model.Product,
	productLevels []*model.ProductLevel,
) error {

	freePayments := make([]*model.Payment, 0)

	for _, productLevel := range productLevels {
		freePayments = append(freePayments, model.NewFreePaymentForProductLevel(
			userID, product, productLevel))
	}

	err := s.paymentRepository.Create(ctx, freePayments...)
	if err != nil {
		return fmt.Errorf("failed to level up user: %w", err)
	}

	return nil
}

func (s *UserService) Levels(
	ctx context.Context,
	miniAppID, userID uuid.UUID, req *model.UserLevelsRequest,
) ([]*model.StudentProductLevels, error) {

	levels, err := s.userRepository.Levels(ctx, miniAppID, userID, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed get user levels: %w", err)
	}

	return levels, nil
}
