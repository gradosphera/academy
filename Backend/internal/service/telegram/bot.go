package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type tgBotConfig struct {
	Response struct {
		Command    string `json:"command"`
		Message    string `json:"message"`
		ButtonText string `json:"button_text"`
		ButtonLink string `json:"button_link"`
	} `json:"response"`
}

func (s *Service) runBot(bot *tgbotapi.BotAPI) {
	s.logger.Info("authorized bot", zap.Int64("bot_id", bot.Self.ID))

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() && update.Message.Command() == s.botConfig.Response.Command {
			msg := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileBytes{
				Name: "image.png", Bytes: s.botImage,
			})

			msg.ParseMode = tgbotapi.ModeHTML

			msg.Caption = s.botConfig.Response.Message

			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL(
						s.botConfig.Response.ButtonText,
						s.botConfig.Response.ButtonLink,
					),
				),
			)

			if _, err := bot.Send(msg); err != nil {
				s.logger.Error("failed to send message", zap.Error(err))
			}
		}
	}
}
