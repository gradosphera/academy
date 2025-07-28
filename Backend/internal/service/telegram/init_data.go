package telegram

import (
	"fmt"

	initdata "github.com/telegram-mini-apps/init-data-golang"
)

func (s *Service) ParseToken(
	botID int64,
	rawInitData string,
	validate bool,
) (*initdata.InitData, error) {

	if validate {
		if err := initdata.ValidateThirdParty(rawInitData, botID, s.tgTokenTTL); err != nil {
			return nil, fmt.Errorf("initdata.ValidateThirdParty: %w", err)
		}
	}

	initData, err := initdata.Parse(rawInitData)
	if err != nil {
		return nil, fmt.Errorf("initdata.Parse: %w", err)
	}

	return &initData, nil
}

func (s *Service) ParseAdminToken(
	rawInitData string,
	validate bool,
) (*initdata.InitData, error) {

	if validate {
		if err := initdata.Validate(rawInitData, s.adminBotToken, s.tgTokenTTL); err != nil {
			return nil, fmt.Errorf("initdata.Validate: %w", err)
		}
	}

	initData, err := initdata.Parse(rawInitData)
	if err != nil {
		return nil, fmt.Errorf("initdata.Parse: %w", err)
	}

	return &initData, nil
}
