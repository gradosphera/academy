package upload

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) ConvertVideoForFastStart(ctx context.Context, filename string) error {
	s.toFastStartMu.Lock()
	defer s.toFastStartMu.Unlock()

	materialFilenameDir := strings.TrimSuffix(filename, filepath.Base(filename))

	tempFilename := filepath.Join(materialFilenameDir, uuid.New().String()+filepath.Ext(filename))

	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-f", "mp4",
		"-i", filepath.Join(s.uploadDir, filename),
		"-c:v", "copy",
		"-c:a", "copy",
		"-movflags", "+faststart",
		filepath.Join(s.uploadDir, tempFilename),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg error: %w: %v", err, stderr.String())
	}

	err := os.Rename(filepath.Join(s.uploadDir, tempFilename), filepath.Join(s.uploadDir, filename))
	if err != nil {
		return fmt.Errorf("failed to replace file: %w", err)
	}

	return nil
}

func (s *Service) CompressVideo(ctx context.Context, filename string) (int64, error) {
	s.compressMu.Lock()
	defer s.compressMu.Unlock()

	materialFilenameDir := strings.TrimSuffix(filename, filepath.Base(filename))

	tempFilename := filepath.Join(materialFilenameDir, uuid.New().String()+".mp4")

	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-f", "mp4",
		"-i", filepath.Join(s.uploadDir, filename),
		"-c:v", "libx264",
		"-crf", "21",
		"-maxrate", "6M", "-bufsize", "12M",
		"-preset", "faster",
		"-c:a", "aac", "-b:a", "128k",
		"-movflags", "+faststart",
		filepath.Join(s.uploadDir, tempFilename),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	// cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return 0, fmt.Errorf("ffmpeg error: %w: %v", err, stderr.String())
		// return fmt.Errorf("ffmpeg error: %w", err)
	}

	err := os.Rename(filepath.Join(s.uploadDir, tempFilename), filepath.Join(s.uploadDir, filename))
	if err != nil {
		return 0, fmt.Errorf("failed to replace file: %w", err)
	}

	info, err := os.Stat(filepath.Join(s.uploadDir, filename))
	if err != nil {
		return 0, fmt.Errorf("failed to get new file info: %w", err)
	}

	return info.Size(), nil
}
