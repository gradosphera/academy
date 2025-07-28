package upload

import (
	"academy/internal/model"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	muxgo "github.com/muxinc/mux-go"
	"go.uber.org/zap"
)

const muxBaseURL = "https://api.mux.com"
const muxTokenExpiration = 12 * time.Hour
const chunkSize = 40 * 1024 * 1024 // 40 MiB.

var ErrNotFound = errors.New("asset not found")

func (s *Service) MuxUpload(ctx context.Context, filename string) (*model.MuxVideoMetadata, error) {
	uploadResp, err := s.muxCreateUpload(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create mux upload: %w", err)
	}
	uploadID := uploadResp.Data.Id
	uploadURL := uploadResp.Data.Url

	// err = s.muxUploadFile(ctx, filename, uploadURL)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to upload file to mux: %w", err)
	// }

	err = s.muxChunkedUploadFile(ctx, filename, uploadURL, chunkSize)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to mux by chunks: %w", err)
	}

	var assetID string
	for range 5 {
		updatedUploadResp, err := s.muxGetUpload(ctx, uploadID)
		if err != nil {
			return nil, fmt.Errorf("failed to get updated mux upload: %w", err)
		}

		if updatedUploadResp.Data.Status == "asset_created" {
			assetID = updatedUploadResp.Data.AssetId
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	if assetID == "" {
		return nil, fmt.Errorf("failed waiting for asset creation")
	}

	asset, err := s.MuxGetAsset(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	if len(asset.Data.PlaybackIds) == 0 {
		return nil, fmt.Errorf("no playback ids included in asset")
	}

	playback := asset.Data.PlaybackIds[0]

	if playback.Policy != muxgo.SIGNED {
		return nil, fmt.Errorf("wrong playback policy: %v", playback.Policy)
	}

	// playbackURL := fmt.Sprintf("https://stream.mux.com/%s.m3u8?token=required", playback.Id)

	if err := os.Remove(filepath.Join(s.uploadDir, filename)); err != nil {
		s.logger.Error("failed to delete file", zap.String("file", filename), zap.Error(err))
	} else {
		s.logger.Info("deleted file", zap.String("file", filename))
	}

	return &model.MuxVideoMetadata{
		AssetID:    assetID,
		PlaybackID: playback.Id,
	}, nil
}

func (s *Service) MuxDeleteAsset(ctx context.Context, assetID string) error {
	if assetID == "" {
		return fmt.Errorf("no asset id provided")
	}

	if _, ok := s.assetsToKeep[assetID]; ok {
		s.logger.Warn("skipped deletion of asset", zap.String("asset_id", assetID))
		return nil
	}

	u, err := url.JoinPath(muxBaseURL, "video/v1/assets", assetID)
	if err != nil {
		return fmt.Errorf("failed to create url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.muxConfig.MuxTokenID, s.muxConfig.MuxTokenSecret)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return ErrNotFound
	default:
		return fmt.Errorf("request failed with status: %v", resp.Status)
	}
}

func (s *Service) muxCreateUpload(ctx context.Context) (*muxgo.UploadResponse, error) {
	payload := map[string]any{
		"timeout":     3600,
		"cors_origin": "*",
		"new_asset_settings": map[string]any{
			"playback_policies": []muxgo.PlaybackPolicy{muxgo.SIGNED},
			// "advanced_playback_policies": []map[string]any{{ // Use instead of playback_policies.
			// 	"policy":               "drm",
			// 	"drm_configuration_id": "",
			// }},
			// "video_quality":     "basic", // "basic" | "plus" | "premium"
		},
		// "test": true, // Uncomment to create temporary asset.
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	u, err := url.JoinPath(muxBaseURL, "video/v1/uploads")
	if err != nil {
		return nil, fmt.Errorf("failed to create url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.muxConfig.MuxTokenID, s.muxConfig.MuxTokenSecret)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("request failed with status: %v", resp.Status)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var uploadResp muxgo.UploadResponse
	err = json.Unmarshal(rawBody, &uploadResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload response: %w", err)
	}

	return &uploadResp, nil
}

func (s *Service) muxUploadFile(ctx context.Context, filename, uploadURL string) error {
	file, err := os.Open(filepath.Join(s.uploadDir, filename))
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}

	req.Header.Set("Content-Type", "video/mp4")

	// Create custom http client with timeout to process large files.
	client := &http.Client{Timeout: 20 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload to mux: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mux upload error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *Service) muxChunkedUploadFile(
	ctx context.Context, filename, uploadURL string, chunkSize int64,
) error {

	file, err := os.Open(filepath.Join(s.uploadDir, filename))
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := fileInfo.Size()

	// Create custom http client with timeout to process large files.
	client := &http.Client{Timeout: 20 * time.Minute}

	offset := int64(0)
loop:
	for offset < fileSize {
		s.logger.Info("uploading chunk",
			zap.Int64("offset", offset),
			zap.Int64("total_size", fileSize),
			zap.String("file", filepath.Base(filename)),
		)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if fileSize < offset+chunkSize {
			chunkSize = fileSize - offset
		}

		chunk := make([]byte, chunkSize)
		chunkLen, err := file.ReadAt(chunk, offset)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read chunk at offset %d: %w", offset, err)
		}
		chunk = chunk[:chunkLen]

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(chunk))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "video/mp4")
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d",
			offset, offset+int64(chunkLen)-1, fileSize))
		req.Header.Set("Content-Length", strconv.Itoa(chunkLen))

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to upload chunk to mux: %v", err)
		}

		offset += int64(chunkLen)

		switch resp.StatusCode {
		case http.StatusPermanentRedirect:
			continue
		case http.StatusCreated, http.StatusOK:
			break loop
		default:
			return fmt.Errorf("chunk upload failed with status %s", resp.Status)
		}
	}

	return nil
}

func (s *Service) MuxGetAsset(ctx context.Context, assetID string) (*muxgo.AssetResponse, error) {
	if assetID == "" {
		return nil, fmt.Errorf("no asset id provided")
	}

	u, err := url.JoinPath(muxBaseURL, "video/v1/assets", assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to create url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.muxConfig.MuxTokenID, s.muxConfig.MuxTokenSecret)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %v", resp.Status)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var assetResp muxgo.AssetResponse
	err = json.Unmarshal(rawBody, &assetResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal asset response: %w", err)
	}

	return &assetResp, nil
}

func (s *Service) muxGetUpload(ctx context.Context, uploadID string) (*muxgo.UploadResponse, error) {
	if uploadID == "" {
		return nil, fmt.Errorf("no upload id provided")
	}

	u, err := url.JoinPath(muxBaseURL, "video/v1/uploads", uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to create url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.muxConfig.MuxTokenID, s.muxConfig.MuxTokenSecret)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %v", resp.Status)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var uploadResp muxgo.UploadResponse
	err = json.Unmarshal(rawBody, &uploadResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload response: %w", err)
	}

	return &uploadResp, nil
}

func (s *Service) MuxSignPrivateVideo(playbackID string) (string, error) {
	decodedKey, err := base64.StdEncoding.DecodeString(s.muxConfig.MuxSigningPrivateKey)
	if err != nil {
		return "", fmt.Errorf("could not decode base64 private key: %w", err)
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(decodedKey)
	if err != nil {
		return "", fmt.Errorf("could not parse RSA private key: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": playbackID,
		"aud": "v",
		"exp": time.Now().Add(muxTokenExpiration).Unix(),
		"kid": s.muxConfig.MuxSigningKey,
	})

	tokenString, err := token.SignedString(signKey)
	if err != nil {
		return "", fmt.Errorf("could not generate token: %w", err)
	}

	return tokenString, nil
}
