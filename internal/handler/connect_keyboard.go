package handler

import (
	"github.com/go-telegram/bot/models"
	"remnawave-tg-shop-bot/internal/config"
)

// createConnectKeyboard returns full keyboard with Connect, Telegram, and navigation buttons
func (h Handler) createConnectKeyboard(lang string) [][]models.InlineKeyboardButton {
	var keyboard [][]models.InlineKeyboardButton

	// Кнопка "Подключиться"
	if config.GetMiniAppURL() != "" {
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{Text: h.translation.GetText(lang, "connect_button"), WebApp: &models.WebAppInfo{
				URL: config.GetMiniAppURL(),
			}},
		})
	} else {
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{Text: h.translation.GetText(lang, "connect_button"), CallbackData: CallbackConnect},
		})
	}

	// Кнопка "Оживить Telegram" если TG_PROXY_LINK задан
	if config.TgProxyLink() != "" {
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{Text: h.translation.GetText(lang, "tg_proxy_button"), URL: config.TgProxyLink()},
		})
	}

	// Кнопка "Мои подписки"
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: h.translation.GetText(lang, "my_subscriptions_button"), CallbackData: CallbackMySubscriptions},
	})

	// Кнопка "В главное меню"
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackStart},
	})

	return keyboard
}
