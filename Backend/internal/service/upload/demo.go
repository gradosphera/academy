package upload

import (
	"academy/internal/model"
	"academy/internal/types"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const demoDir = "demo"

const productStructureFilename = "product.json"

const miniAppLogoFilename = "mini_app_logo.png"
const productCoverFilename = "product_cover.png"
const videoCoverFilename = "video_cover.png"
const audioCoverFilename = "audio_cover.png"
const circularVideoCoverFilename = "circular_video_cover.png"
const textCoverFilename = "text_cover.png"

const audioOriginalFilename = "audio_lesson_file.mp3"

// Moved to Mux video hosting service.
// const circularVideoOriginalFilename = "circular_video_lesson_file.mp4"

func (s *Service) AddDemoProduct(miniApp *model.MiniApp) error {
	var tempFiles []string
	defer func() {
		for _, filename := range tempFiles {
			if filename == "" {
				continue
			}
			err := s.Delete(filepath.Join(s.uploadDir, filename))
			if err != nil {
				s.logger.Error("error while deleting file", zap.String("err", err.Error()))
			}
		}
	}()

	now := time.Now()

	materialPath := MaterialFilePath{MiniAppID: miniApp.ID}

	if err := os.MkdirAll(filepath.Join(s.uploadDir, materialPath.String()), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	miniApp.Logo = filepath.Join(materialPath.String(), uuid.NewString()+filepath.Ext(miniAppLogoFilename))

	err := os.Link(
		filepath.Join(s.uploadDir, demoDir, miniAppLogoFilename),
		filepath.Join(s.uploadDir, miniApp.Logo),
	)
	if err != nil {
		return fmt.Errorf("failed to create link for default mini-app logo: %w", err)
	}
	tempFiles = append(tempFiles, miniApp.Logo)

	miniAppLogoStat, err := os.Stat(filepath.Join(s.uploadDir, miniApp.Logo))
	if err != nil {
		return fmt.Errorf("os.Stat: %w", err)
	}
	miniApp.LogoSize = miniAppLogoStat.Size()

	productFilename := filepath.Join(s.uploadDir, demoDir, productStructureFilename)

	productFile, err := os.Open(productFilename)
	if err != nil {
		return fmt.Errorf("failed to open demo product: %w", err)
	}
	defer productFile.Close()

	var product *model.Product
	err = json.NewDecoder(productFile).Decode(&product)
	if err != nil {
		return fmt.Errorf("failed to decode demo product: %w", err)
	}

	product.ID = uuid.New()
	product.MiniAppID = miniApp.ID
	product.UpdatedAt = now
	product.CreatedAt = now

	materialPath.ProductID = product.ID
	if err := os.MkdirAll(filepath.Join(s.uploadDir, materialPath.String()), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	product.Cover = filepath.Join(materialPath.String(), uuid.NewString()+filepath.Ext(productCoverFilename))

	err = os.Link(
		filepath.Join(s.uploadDir, demoDir, productCoverFilename),
		filepath.Join(s.uploadDir, product.Cover),
	)
	if err != nil {
		return fmt.Errorf("failed to create link for demo product cover: %w", err)
	}
	tempFiles = append(tempFiles, product.Cover)

	productCoverStat, err := os.Stat(filepath.Join(s.uploadDir, product.Cover))
	if err != nil {
		return fmt.Errorf("os.Stat: %w", err)
	}
	product.CoverSize = productCoverStat.Size()

	for _, lesson := range product.Lessons {
		lesson.ID = uuid.New()
		lesson.ProductID = product.ID
		lesson.UpdatedAt = now
		lesson.CreatedAt = now

		if lesson.ContentType == model.LessonTypeEvent {
			releaseDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).
				AddDate(0, 0, 1)
			lesson.ReleaseDate = types.NewTime(releaseDate)
			lesson.AccessTime = types.NewInterval(types.JsonInterval{Days: 1})
		}

		for _, material := range lesson.Materials {
			material.ID = uuid.New()
			material.LessonID = lesson.ID
			material.UpdatedAt = now
			material.CreatedAt = now

			if material.Category != model.MaterialCategoryLessonContent {
				continue
			}

			materialPath.LessonID = lesson.ID
			if err := os.MkdirAll(filepath.Join(s.uploadDir, materialPath.String()), 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}

			switch material.ContentType {
			case model.MaterialTypeVideo:
				coverMaterial := &model.Material{
					ID:               uuid.New(),
					LessonID:         lesson.ID,
					Category:         model.MaterialCategoryLessonCover,
					ContentType:      model.MaterialTypePicture,
					Title:            material.Title,
					Description:      material.Description,
					Filename:         filepath.Join(materialPath.String(), uuid.NewString()+filepath.Ext(videoCoverFilename)),
					OriginalFilename: videoCoverFilename,
					UpdatedAt:        now,
					CreatedAt:        now,
				}
				err = os.Link(
					filepath.Join(s.uploadDir, demoDir, videoCoverFilename),
					filepath.Join(s.uploadDir, coverMaterial.Filename),
				)
				if err != nil {
					return fmt.Errorf("failed to create link for demo video cover: %w", err)
				}
				lesson.Materials = append(lesson.Materials, coverMaterial)
				tempFiles = append(tempFiles, coverMaterial.Filename)
			case model.MaterialTypeAudio:
				audioFile, err := os.Open(filepath.Join(s.uploadDir, demoDir, audioOriginalFilename))
				if err != nil {
					return fmt.Errorf("failed to open demo audio: %w", err)
				}
				defer audioFile.Close()

				material.OriginalFilename = audioOriginalFilename
				material.Filename = filepath.Join(materialPath.String(), uuid.NewString()+filepath.Ext(audioOriginalFilename))

				err = os.Link(
					filepath.Join(s.uploadDir, demoDir, audioOriginalFilename),
					filepath.Join(s.uploadDir, material.Filename),
				)
				if err != nil {
					return fmt.Errorf("failed to create link for demo audio: %w", err)
				}

				coverMaterial := &model.Material{
					ID:               uuid.New(),
					LessonID:         lesson.ID,
					Category:         model.MaterialCategoryLessonCover,
					ContentType:      model.MaterialTypePicture,
					Title:            material.Title,
					Description:      material.Description,
					Filename:         filepath.Join(materialPath.String(), uuid.NewString()+filepath.Ext(audioCoverFilename)),
					OriginalFilename: audioCoverFilename,
					UpdatedAt:        now,
					CreatedAt:        now,
				}
				err = os.Link(
					filepath.Join(s.uploadDir, demoDir, audioCoverFilename),
					filepath.Join(s.uploadDir, coverMaterial.Filename),
				)
				if err != nil {
					return fmt.Errorf("failed to create link for demo audio cover: %w", err)
				}
				lesson.Materials = append(lesson.Materials, coverMaterial)
				tempFiles = append(tempFiles, coverMaterial.Filename)
			case model.MaterialTypeText:
				coverMaterial := &model.Material{
					ID:               uuid.New(),
					LessonID:         lesson.ID,
					Category:         model.MaterialCategoryLessonCover,
					ContentType:      model.MaterialTypePicture,
					Title:            material.Title,
					Description:      material.Description,
					Filename:         filepath.Join(materialPath.String(), uuid.NewString()+filepath.Ext(textCoverFilename)),
					OriginalFilename: textCoverFilename,
					UpdatedAt:        now,
					CreatedAt:        now,
				}
				err = os.Link(
					filepath.Join(s.uploadDir, demoDir, textCoverFilename),
					filepath.Join(s.uploadDir, coverMaterial.Filename),
				)
				if err != nil {
					return fmt.Errorf("failed to create link for demo text cover: %w", err)
				}
				lesson.Materials = append(lesson.Materials, coverMaterial)
				tempFiles = append(tempFiles, coverMaterial.Filename)
			case model.MaterialTypeCircleVideo:
				coverMaterial := &model.Material{
					ID:               uuid.New(),
					LessonID:         lesson.ID,
					Category:         model.MaterialCategoryLessonCover,
					ContentType:      model.MaterialTypePicture,
					Title:            material.Title,
					Description:      material.Description,
					Metadata:         material.Metadata,
					Filename:         filepath.Join(materialPath.String(), uuid.NewString()+filepath.Ext(circularVideoCoverFilename)),
					OriginalFilename: circularVideoCoverFilename,
					UpdatedAt:        now,
					CreatedAt:        now,
				}
				err = os.Link(
					filepath.Join(s.uploadDir, demoDir, circularVideoCoverFilename),
					filepath.Join(s.uploadDir, coverMaterial.Filename),
				)
				if err != nil {
					return fmt.Errorf("failed to create link for demo circular video cover: %w", err)
				}
				lesson.Materials = append(lesson.Materials, coverMaterial)
				tempFiles = append(tempFiles, coverMaterial.Filename)
			}
			tempFiles = append(tempFiles, material.Filename)
		}

		for _, m := range lesson.Materials {
			materialStat, err := os.Stat(filepath.Join(s.uploadDir, m.Filename))
			if err != nil {
				return fmt.Errorf("os.Stat: %w", err)
			}
			m.Size = materialStat.Size()
		}
	}

	miniApp.Products = append(miniApp.Products, product)

	tempFiles = nil

	return nil
}
