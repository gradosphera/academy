package service

import (
	repo "academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/storage/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type MiniAppService struct {
	miniAppRepository  *repository.MiniAppRepository
	productRepository  *repository.ProductRepository
	lessonRepository   *repository.LessonRepository
	materialRepository *repository.MaterialRepository
	userRepository     *repository.UserRepository
	transactionManager *repo.TransactionManager
}

func NewMiniAppService(
	miniAppRepository *repository.MiniAppRepository,
	productRepository *repository.ProductRepository,
	lessonRepository *repository.LessonRepository,
	materialRepository *repository.MaterialRepository,
	userRepository *repository.UserRepository,
	transactionManager *repo.TransactionManager,
) *MiniAppService {

	return &MiniAppService{
		miniAppRepository:  miniAppRepository,
		productRepository:  productRepository,
		lessonRepository:   lessonRepository,
		materialRepository: materialRepository,
		userRepository:     userRepository,
		transactionManager: transactionManager,
	}
}

func (s *MiniAppService) Create(
	ctx context.Context,
	miniApp *model.MiniApp,
	owner *model.User,
) error {

	err := s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err := s.miniAppRepository.WithTx(tx).Create(ctx, miniApp)
		if err != nil {
			return fmt.Errorf("failed to create mini app: %w", err)
		}

		err = s.userRepository.WithTx(tx).Create(ctx, owner)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		for _, product := range miniApp.Products {
			err := s.productRepository.WithTx(tx).Create(ctx, product)
			if err != nil {
				return fmt.Errorf("failed to create product: %w", err)
			}

			for _, lesson := range product.Lessons {
				err := s.lessonRepository.WithTx(tx).Create(ctx, lesson)
				if err != nil {
					return fmt.Errorf("failed to create lesson: %w", err)
				}

				for _, material := range lesson.Materials {
					err := s.materialRepository.WithTx(tx).Create(ctx, material)
					if err != nil {
						return fmt.Errorf("failed to create material: %w", err)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create mini app: %w", err)
	}

	return nil
}

func (s *MiniAppService) GetByID(ctx context.Context, id uuid.UUID) (*model.MiniApp, error) {
	miniApp, err := s.miniAppRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get mini_app by id: %w", err)
	}

	return miniApp, nil
}

func (s *MiniAppService) GetPlanByPlanID(ctx context.Context, planID model.PlanID) (*model.Plan, error) {
	plan, err := s.miniAppRepository.GetPlanByPlanID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mini_app plan by plan_id: %w", err)
	}

	return plan, nil
}

func (s *MiniAppService) GetByName(ctx context.Context, name string) (*model.MiniApp, error) {
	miniApp, err := s.miniAppRepository.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get mini_app by name: %w", err)
	}

	return miniApp, nil
}

func (s *MiniAppService) GetByOwnerTelegramID(ctx context.Context, ownerTelegramID int64) (*model.MiniApp, error) {
	miniApp, err := s.miniAppRepository.GetByOwnerTelegramID(ctx, ownerTelegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mini_app by owner tg id: %w", err)
	}

	return miniApp, nil
}

func (s *MiniAppService) GetByModTelegramID(ctx context.Context, modTelegramID int64) ([]*model.MiniApp, error) {
	miniApps, err := s.miniAppRepository.GetByModTelegramID(ctx, modTelegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mini_app by mod tg id: %w", err)
	}

	return miniApps, nil
}

func (s *MiniAppService) Update(ctx context.Context, model *model.MiniApp) error {
	err := s.miniAppRepository.Update(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to update mini_app: %w", err)
	}

	return nil
}

func (s *MiniAppService) UpdateWithUsers(
	ctx context.Context,
	miniApp *model.MiniApp,
	users ...*model.User,
) error {

	err := s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err := s.miniAppRepository.WithTx(tx).Update(ctx, miniApp)
		if err != nil {
			return fmt.Errorf("failed to update mini_app: %w", err)
		}

		for _, u := range users {
			err := s.userRepository.WithTx(tx).Update(ctx, u)
			if err != nil {
				return fmt.Errorf("failed to update user: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *MiniAppService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.miniAppRepository.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete mini_app by id: %w", err)
	}

	return nil
}

func (s *MiniAppService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := s.miniAppRepository.SoftDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete mini_app by id: %w", err)
	}

	return nil
}

func (s *MiniAppService) Restore(ctx context.Context, id uuid.UUID) error {
	err := s.miniAppRepository.Restore(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to restore mini_app by id: %w", err)
	}

	return nil
}

func (s *MiniAppService) DeleteOld(ctx context.Context, olderThan time.Time) ([]uuid.UUID, error) {
	ids, err := s.miniAppRepository.DeleteOld(ctx, olderThan)
	if err != nil {
		return nil, fmt.Errorf("failed to delete old mini_apps that older then %v: %w", olderThan, err)
	}

	return ids, nil
}

func (s *MiniAppService) ReorderSlides(
	ctx context.Context,
	miniAppID uuid.UUID,
	req *model.EditSlidesRequest,
	newSlide *model.Material,
) (*model.MiniApp, error) {

	miniApp, err := s.GetByID(ctx, miniAppID)
	if err != nil {
		return nil, fmt.Errorf("error getting mini-app: %w", err)
	}
	if miniApp == nil {
		return nil, nil
	}

	newSlides := make(map[uuid.UUID]*model.Material)
	for _, sl := range miniApp.Slides {
		newSlides[sl.ID] = sl
	}
	if newSlide != nil {
		newSlides[newSlide.ID] = newSlide
	}

	now := time.Now().UTC()
	currIndex := int64(0)
	for _, sl := range req.SlidesOrder {
		if newSlide != nil && currIndex == newSlide.Index {
			currIndex++
		}

		slide, ok := newSlides[sl]
		if !ok {
			return nil, fmt.Errorf("slide is not included in the mini-app: %v", s)
		}

		slide.Index = currIndex
		slide.UpdatedAt = now

		currIndex++
	}
	if newSlide != nil && currIndex <= newSlide.Index {
		newSlide.Index = currIndex
		currIndex++
	}

	if len(newSlides) != int(currIndex) {
		return nil, fmt.Errorf("new number of slides do not match with old")
	}

	miniApp.Slides = make([]*model.Material, len(newSlides))

	err = s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		for _, sl := range newSlides {
			miniApp.Slides[sl.Index] = sl

			if newSlide != nil && sl.ID == newSlide.ID {
				err := s.materialRepository.WithTx(tx).Create(ctx, sl)
				if err != nil {
					return fmt.Errorf("error saving new slide: %w", err)
				}
				continue
			}

			err := s.materialRepository.WithTx(tx).Update(ctx, sl)
			if err != nil {
				return fmt.Errorf("error updating slide: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return miniApp, nil
}

func (s *MiniAppService) Analytics(
	ctx context.Context,
	miniAppID uuid.UUID,
	req *model.AnalyticsRequest,
) (*model.Analytics, error) {

	analytics, err := s.miniAppRepository.Analytics(ctx, miniAppID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return analytics, nil
}

func (s *MiniAppService) GetInfo(ctx context.Context, id uuid.UUID) (*model.MiniAppInfo, error) {
	info, err := s.miniAppRepository.GetInfo(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get mini_app info: %w", err)
	}

	return info, nil
}

func (s *MiniAppService) CreateModInvite(
	ctx context.Context, miniAppID uuid.UUID, req *model.CreateModInviteRequest,
) (*model.ModInvite, error) {

	invite := model.NewModInvite(miniAppID)

	err := s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err := s.miniAppRepository.WithTx(tx).CreateModInvite(ctx, invite)
		if err != nil {
			return fmt.Errorf("failed to create mod invite: %w", err)
		}

		for _, p := range req.Permissions {
			mip := model.NewModInvitePermission(invite.ID, p)

			err := s.miniAppRepository.WithTx(tx).CreateModInvitePermission(ctx, mip)
			if err != nil {
				return fmt.Errorf("failed to create mod invite permission: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return invite, nil
}

func (s *MiniAppService) EditModInvite(
	ctx context.Context, inviteID uuid.UUID, req *model.EditModInviteRequest,
) error {

	err := s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err := s.miniAppRepository.WithTx(tx).DeleteModInvitePermissions(ctx, inviteID)
		if err != nil {
			return fmt.Errorf("failed to delete old permissions: %w", err)
		}

		for _, p := range req.Permissions {
			mip := model.NewModInvitePermission(inviteID, p)

			err := s.miniAppRepository.WithTx(tx).CreateModInvitePermission(ctx, mip)
			if err != nil {
				return fmt.Errorf("failed to create mod invite with new permission: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *MiniAppService) GetModInviteByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.ModInvite, error) {

	access, err := s.miniAppRepository.GetModInvite(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get mod invite by id: %w", err)
	}

	return access, nil
}

func (s *MiniAppService) ModInvites(
	ctx context.Context,
	miniAppID uuid.UUID,
	filter *model.FilterModInvitesRequest,
) ([]*model.ModInvite, int, error) {

	invites, total, err := s.miniAppRepository.ModInvites(ctx, miniAppID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get mod invites: %w", err)
	}

	return invites, total, nil
}

func (s *MiniAppService) UpdateModInvite(
	ctx context.Context,
	invite *model.ModInvite,
) error {

	err := s.miniAppRepository.UpdateModInvite(ctx, invite)
	if err != nil {
		return fmt.Errorf("failed to update mod invite: %w", err)
	}

	return nil
}

func (s *MiniAppService) DeleteModInvite(
	ctx context.Context,
	miniAppID, inviteID uuid.UUID,
) error {

	err := s.miniAppRepository.DeleteModInvite(ctx, miniAppID, inviteID)
	if err != nil {
		return fmt.Errorf("failed to delete mod invite: %w", err)
	}

	return nil
}

func (s *MiniAppService) GetPermissions(
	ctx context.Context, userID uuid.UUID,
) ([]*model.Permission, error) {

	permissions, err := s.miniAppRepository.GetPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permissions, nil
}

func (s *MiniAppService) CheckPermission(
	ctx context.Context, userID uuid.UUID, permissionName ...model.PermissionName,
) (bool, error) {

	ok, err := s.miniAppRepository.CheckPermission(ctx, userID, permissionName...)
	if err != nil {
		return false, fmt.Errorf("failed to get permission: %w", err)
	}

	return ok, nil
}

func (s *MiniAppService) ListPermissions(ctx context.Context) ([]*model.Permission, error) {
	permissions, err := s.miniAppRepository.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get mod permissions: %w", err)
	}

	return permissions, nil
}

func (s *MiniAppService) FindTonAddresses(ctx context.Context) ([]string, error) {
	addresses, err := s.miniAppRepository.FindTonAddresses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find TON adresses: %w", err)
	}

	return addresses, nil
}
