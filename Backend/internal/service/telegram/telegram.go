package telegram

import (
	"academy/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const baseURL = "https://api.telegram.org"

const defaultTimeout = 30 * time.Second

type Service struct {
	logger    *zap.Logger
	client    *http.Client
	botConfig *tgBotConfig
	botImage  []byte

	adminBotToken string
	tgTokenTTL    time.Duration
}

func NewService(logger *zap.Logger, cfg *config.Config) (*Service, error) {
	s := &Service{
		logger:        logger,
		adminBotToken: cfg.Auth.TelegramBotToken,
		tgTokenTTL:    cfg.Auth.TelegramTokenTTL,

		client: &http.Client{Timeout: defaultTimeout},
	}

	if cfg.App.TelegramBotConfig != "" && cfg.App.TelegramBotImage != "" {
		botImage, err := os.ReadFile(cfg.App.TelegramBotImage)
		if err != nil {
			return nil, fmt.Errorf("failed to read tg bot image: %w", err)
		}
		s.botImage = botImage

		f, err := os.Open(cfg.App.TelegramBotConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to open tg bot config: %w", err)
		}

		var botConfig tgBotConfig
		err = json.NewDecoder(f).Decode(&botConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to decode tg bot config: %w", err)
		}
		s.botConfig = &botConfig

		bot, err := tgbotapi.NewBotAPI(s.adminBotToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create bot: %w", err)
		}

		go s.runBot(bot)
	}

	return s, nil
}

func (s *Service) DownloadAvatar(
	ctx context.Context, url string,
) (rawBytes []byte, fileExt string, err error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, "", fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("client.Do: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("resp.StatusCode: %x", resp.Status)
	}

	defer resp.Body.Close()

	rawBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("io.ReadAll: %w", err)
	}

	return rawBytes, strings.ToLower(filepath.Ext(resp.Request.URL.Path)), nil
}
