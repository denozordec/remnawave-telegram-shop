package handler

import (
	"github.com/go-telegram/bot/models"
)

// createConnectKeyboard minimal replacement used by trial fallback
func (h Handler) createConnectKeyboard(lang string) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{{Text: h.translation.GetText(lang, "my_subscriptions_button"), CallbackData: CallbackMySubscriptions}},
		{{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackStart}},
	}
}
