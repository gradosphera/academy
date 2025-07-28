package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type BotInfo struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      *struct {
		ID    int64 `json:"id"`
		IsBot bool  `json:"is_bot"`
	} `json:"result"`
}

func (s *Service) ValidateBotToken(ctx context.Context, token string) (int64, error) {
	u, err := url.JoinPath(baseURL, "bot"+token, "getMe")
	if err != nil {
		return 0, fmt.Errorf("url.JoinPath: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return 0, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("s.client.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status: %v", resp.Status)
	}

	var botInfo BotInfo
	if err := json.NewDecoder(resp.Body).Decode(&botInfo); err != nil {
		return 0, fmt.Errorf("failed to decode response: %v", err)
	}

	if !botInfo.OK || botInfo.Result == nil {
		return 0, fmt.Errorf("error in response: %v", botInfo.Description)
	}

	if !botInfo.Result.IsBot {
		return 0, fmt.Errorf("not a bot")
	}

	return botInfo.Result.ID, nil
}
