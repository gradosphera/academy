package upload

import (
	"academy/internal/config"
	"academy/internal/model"
	"academy/internal/storage/repository"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const MaxChunkAge = 2 * time.Hour
const MaxExcelAge = 10 * time.Minute

type Service struct {
	logger *zap.Logger

	client       *http.Client
	muxConfig    *config.MuxConfig
	assetsToKeep map[string]struct{}

	uploadDir string
	tempDir   string

	toFastStartMu sync.Mutex
	compressMu    sync.Mutex

	userRepository         *repository.UserRepository
	miniAppRepository      *repository.MiniAppRepository
	productRepository      *repository.ProductRepository
	lessonRepository       *repository.LessonRepository
	materialRepository     *repository.MaterialRepository
	chunkRepository        *repository.ChunkRepository
	productLevelRepository *repository.ProductLevelRepository
}

func NewService(
	cfg *config.Config,
	logger *zap.Logger,
	userRepository *repository.UserRepository,
	miniAppRepository *repository.MiniAppRepository,
	productRepository *repository.ProductRepository,
	lessonRepository *repository.LessonRepository,
	materialRepository *repository.MaterialRepository,
	chunkRepository *repository.ChunkRepository,
	productLevelRepository *repository.ProductLevelRepository,
) (*Service, error) {

	dirInfo, err := os.Stat(cfg.App.UploadDirectory)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("directory not exist: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("os.Stat: %w", err)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("provided path is file")
	}

	dirInfo, err = os.Stat(cfg.App.TempUploadDirectory)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("directory not exist: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("os.Stat: %w", err)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("provided path is file")
	}

	demoAssets := make(map[string]struct{}, 0)
	if cfg.App.EnableDemoProduct {
		demoProductFilename := filepath.Join(cfg.App.UploadDirectory, demoDir, productStructureFilename)

		demoProductFile, err := os.Open(demoProductFilename)
		if err != nil {
			return nil, fmt.Errorf("failed to open demo product: %w", err)
		}
		defer demoProductFile.Close()

		var demoProduct *model.Product
		if err := json.NewDecoder(demoProductFile).Decode(&demoProduct); err != nil {
			return nil, fmt.Errorf("failed to decode demo product: %w", err)
		}

		for _, lesson := range demoProduct.Lessons {
			for _, material := range lesson.Materials {
				if len(material.Metadata) == 0 {
					continue
				}

				var metadata model.MuxVideoMetadata
				if err := json.Unmarshal(material.Metadata, &metadata); err != nil {
					continue
				}

				demoAssets[metadata.AssetID] = struct{}{}
			}
		}
	}

	return &Service{
		logger: logger,

		client:       &http.Client{Timeout: cfg.HTTP.Timeout},
		muxConfig:    &cfg.Mux,
		assetsToKeep: demoAssets,

		uploadDir: cfg.App.UploadDirectory,
		tempDir:   cfg.App.TempUploadDirectory,

		userRepository:         userRepository,
		miniAppRepository:      miniAppRepository,
		productRepository:      productRepository,
		lessonRepository:       lessonRepository,
		materialRepository:     materialRepository,
		chunkRepository:        chunkRepository,
		productLevelRepository: productLevelRepository,
	}, nil
}

func (s *Service) Upload(
	path string,
	file io.Reader,
	fileExtension string,
) (filename string, size int64, err error) {

	if err := os.MkdirAll(filepath.Join(s.uploadDir, path), 0755); err != nil {
		return filename, size, fmt.Errorf("failed to create directory: %v", err)
	}

	filename = filepath.Join(path, uuid.New().String()+fileExtension)
	fullFilePath := filepath.Join(s.uploadDir, filename)

	dst, err := os.Create(fullFilePath)
	if err != nil {
		return "", 0, fmt.Errorf("os.Create: %w", err)
	}
	defer dst.Close()

	size, err = io.Copy(dst, file)
	if err != nil {
		return "", 0, err
	}

	return filename, size, nil
}

func (s *Service) UploadExcel(
	path string,
	rawBytes []byte, filenameBase string,
) (filename string, size int64, err error) {

	if err := os.MkdirAll(filepath.Join(s.uploadDir, path), 0755); err != nil {
		return filename, size, fmt.Errorf("failed to create directory: %v", err)
	}

	filename = filepath.Join(path, filenameBase)

	dst, err := os.Create(filepath.Join(s.uploadDir, filename))
	if err != nil {
		return "", 0, fmt.Errorf("os.Create: %w", err)
	}
	defer dst.Close()

	n, err := dst.Write(rawBytes)
	size = int64(n)
	if err != nil {
		return "", size, err
	}

	return filename, size, nil
}

func (s *Service) Delete(filePath string) error {
	if filePath == "" {
		return nil
	}

	fullFilePath := filepath.Join(s.uploadDir, filePath)

	_, err := os.Stat(fullFilePath)
	if err != nil {
		return fmt.Errorf("os.Stat: %w", err)
	}

	err = os.RemoveAll(fullFilePath)
	if err != nil {
		return fmt.Errorf("os.RemoveAll: %w", err)
	}

	return nil
}

func (s *Service) ClearDanglingUploads(ctx context.Context) error {
	err := filepath.WalkDir(s.uploadDir, func(
		uploadFilePath string, info os.DirEntry, err error,
	) error {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}

		if info.IsDir() {
			if uploadFilePath == s.uploadDir {
				return nil
			}
			if uploadFilePath == filepath.Join(s.uploadDir, demoDir) {
				return nil
			}
			isEmpty, err := isDirectoryEmpty(uploadFilePath)
			if err != nil {
				return fmt.Errorf("error checking directory %s: %w", uploadFilePath, err)
			}
			if !isEmpty {
				return nil
			}
			err = os.Remove(uploadFilePath)
			if err != nil {
				s.logger.Error("error deleting empty directory",
					zap.String("directory", uploadFilePath),
					zap.Error(err),
				)
			} else {
				s.logger.Info("deleted empty directory", zap.String("directory", uploadFilePath))
			}

			return nil
		}

		materialFilename := strings.TrimPrefix(uploadFilePath, s.uploadDir)
		materialFilename = strings.TrimPrefix(materialFilename, "/")

		if strings.HasPrefix(materialFilename, demoDir) {
			return nil
		}

		materialPath, err := ParseMaterialFilePath(materialFilename)
		if err != nil {
			s.logger.Error("error in ParseMaterialFilePath",
				zap.String("directory", uploadFilePath),
				zap.Error(err),
			)
			return nil
		}

		// Delete all top level files. Files should be groupped only by mini-apps.
		if materialPath.MiniAppID == uuid.Nil {
			return s.removeFile(uploadFilePath)
		}

		miniApp, err := s.miniAppRepository.GetByID(ctx, materialPath.MiniAppID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to get mini-app: %w", err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return s.removeFile(uploadFilePath)
		}

		if filepath.Ext(materialFilename) == ".xlsx" {
			fullInfo, err := info.Info()
			if err != nil {
				return fmt.Errorf("info.Info: %w", err)
			}

			if MaxExcelAge < time.Since(fullInfo.ModTime()) {
				return s.removeFile(uploadFilePath)
			}

			return nil
		}

		if materialPath.ProductID == uuid.Nil {
			miniAppMaterials := make(map[string]struct{})
			miniAppMaterials[miniApp.Logo] = struct{}{}
			miniAppMaterials[miniApp.TeacherAvatar] = struct{}{}
			for _, s := range miniApp.Slides {
				miniAppMaterials[s.Filename] = struct{}{}
			}
			for _, s := range miniApp.TOS {
				miniAppMaterials[s.Filename] = struct{}{}
			}
			for _, s := range miniApp.PrivacyPolicy {
				miniAppMaterials[s.Filename] = struct{}{}
			}

			if _, ok := miniAppMaterials[materialFilename]; ok {
				return nil
			}

			_, err := s.userRepository.GetByAvatarFilename(ctx, miniApp.ID, materialFilename)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("failed to get user by avatar: %w", err)
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return nil
			}

			return s.removeFile(uploadFilePath)
		}

		product, err := s.productRepository.GetByID(ctx, materialPath.ProductID, false)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to get product: %w", err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return s.removeFile(uploadFilePath)
		}

		if materialPath.LessonID == uuid.Nil {
			if materialFilename == product.Cover {
				return nil
			}

			if materialPath.ProductLevelID != uuid.Nil {
				material, err := s.materialRepository.GetByFilename(ctx, materialFilename)
				if err != nil {
					return fmt.Errorf("failed to get bonus material: %w", err)
				}
				if material == nil {
					return s.removeFile(uploadFilePath)
				}
				return nil
			}

			return s.removeFile(uploadFilePath)
		}

		lesson, err := s.lessonRepository.GetByID(ctx, materialPath.LessonID, materialPath.UserID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to get lesson: %w", err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return s.removeFile(uploadFilePath)
		}

		if materialPath.UserID == uuid.Nil {
			material, err := s.materialRepository.GetByFilename(ctx, materialFilename)
			if err != nil {
				return fmt.Errorf("failed to get material: %w", err)
			}
			if material == nil {
				return s.removeFile(uploadFilePath)
			}

			return nil
		}

		for _, progress := range lesson.Progress {
			var progressData model.LessonProgressData
			err := json.Unmarshal(progress.Data, &progressData)
			if err != nil {
				return fmt.Errorf("json.Unmarshal: %w", err)
			}
			for _, progressFile := range progressData.FilesMetadata {
				if progressFile.Filename == materialFilename {
					return nil
				}
			}
		}

		return s.removeFile(uploadFilePath)
	})

	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("error deleting files: %w", err)
	}

	return nil
}

func isDirectoryEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if errors.Is(err, io.EOF) {
		return true, nil
	}

	return false, nil
}

func (s *Service) removeFile(filename string) error {
	err := os.Remove(filename)
	if err != nil {
		return fmt.Errorf("error deleting file %s: %v", filename, err)
	}
	s.logger.Info("deleted file", zap.String("file", filename))
	return nil
}
