package upload

import (
	"academy/internal/model"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ErrHashNotMatch = errors.New("hashsum not match")

func (s *Service) AddChunk(
	materialID uuid.UUID, chunkIndex int64,
	file io.Reader, hashsum string,
) (size int64, err error) {

	materialDir := filepath.Join(s.tempDir, materialID.String())

	if err := os.MkdirAll(materialDir, 0755); err != nil {
		return size, fmt.Errorf("failed to create directory: %v", err)
	}

	filePath := filepath.Join(materialDir, fmt.Sprintf("%d", chunkIndex))

	dst, err := os.Create(filePath)
	if err != nil {
		return size, fmt.Errorf("os.Create: %w", err)
	}
	defer dst.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(dst, hasher)

	size, err = io.Copy(writer, file)
	if err != nil {
		return size, err
	}

	if hashsum != hex.EncodeToString(hasher.Sum(nil)) {
		err := os.Remove(filePath)
		if err != nil {
			s.logger.Error("failed to remove chunk with wrong hashsum", zap.Error(err))
		}

		return size, ErrHashNotMatch
	}

	return size, nil
}

func (s *Service) UploadChunks(
	ctx context.Context,
	materialFilenameDir string, fileExtension string,
	chunks []*model.Chunk,
) (filename string, size int64, err error) {

	if len(chunks) == 0 {
		return filename, size, fmt.Errorf("no chunks provided")
	}

	if err := os.MkdirAll(filepath.Join(s.uploadDir, materialFilenameDir), 0755); err != nil {
		return filename, size, fmt.Errorf("failed to create directory: %v", err)
	}

	filename = filepath.Join(materialFilenameDir, uuid.New().String()+fileExtension)

	dst, err := os.Create(filepath.Join(s.uploadDir, filename))
	if err != nil {
		return filename, size, fmt.Errorf("os.Create: %w", err)
	}
	defer dst.Close()

	var materialID uuid.UUID
	for _, chunk := range chunks {
		select {
		case <-ctx.Done():
			return filename, size, ctx.Err()
		default:
		}

		materialID = chunk.MaterialID
		chunkPath := filepath.Join(s.tempDir, chunk.MaterialID.String(), strconv.Itoa(int(chunk.Index)))

		src, err := os.Open(chunkPath)
		if err != nil {
			return filename, size, fmt.Errorf("failed to open file %s: %w", chunkPath, err)
		}

		chunkSize, err := io.Copy(dst, src)

		src.Close()

		if err != nil {
			return filename, size, fmt.Errorf("failed to copy file %s: %w", chunkPath, err)
		}

		size += chunkSize
	}

	dst.Close()

	if err := s.ClearChunks(materialID); err != nil {
		return filename, size, fmt.Errorf("error deleting chunks: %w", err)
	}

	return filename, size, nil
}

func (s *Service) ClearChunks(materialID uuid.UUID) error {
	materialDir := filepath.Join(s.tempDir, materialID.String())

	return os.RemoveAll(materialDir)
}

func (s *Service) ClearOldChunks(ctx context.Context) error {
	err := s.chunkRepository.DeleteOlderThen(ctx, time.Now().Add(-MaxChunkAge))
	if err != nil {
		return fmt.Errorf("failed to delete old chunks: %w", err)
	}

	err = filepath.WalkDir(s.tempDir, func(chunkPath string, info os.DirEntry, err error) error {
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
			if chunkPath == s.tempDir {
				return nil
			}
			isEmpty, err := isDirectoryEmpty(chunkPath)
			if err != nil {
				return fmt.Errorf("error checking directory %s: %w", chunkPath, err)
			}
			if !isEmpty {
				return nil
			}
			err = os.Remove(chunkPath)
			if err != nil {
				s.logger.Error("error deleting empty directory",
					zap.String("directory", chunkPath),
					zap.Error(err),
				)
			} else {
				s.logger.Info("deleted empty directory", zap.String("directory", chunkPath))
			}

			return nil
		}

		fullInfo, err := info.Info()
		if err != nil {
			return fmt.Errorf("info.Info: %w", err)
		}

		modifiedAt := fullInfo.ModTime()

		if MaxChunkAge < time.Since(modifiedAt) {
			err := os.Remove(chunkPath)
			if err != nil {
				return fmt.Errorf("error deleting file %s: %v", chunkPath, err)
			}
			s.logger.Info("deleted file",
				zap.String("file", chunkPath),
				zap.String("modified_at", modifiedAt.String()),
			)
		}

		return nil
	})

	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("failed to delete old chunk files: %w", err)
	}

	return nil
}
