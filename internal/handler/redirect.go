package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// afterSubscriptionCreated: без synthetic Update — напрямую рендерим табличку и делаем EditMessageText
func (h Handler) afterSubscriptionCreated(ctx context.Context, b *bot.Bot, chatID int64, messageID int) {
	// Получаем клиента и подписки
	customer, err := h.customerRepository.FindByTelegramId(ctx, chatID)
	if err != nil || customer == nil { return }
	subs, err := h.subscriptionRepository.GetActiveSubscriptions(ctx, customer.ID)
	if err != nil { return }

	lang := "ru"
	// Собираем текст таблички
	msg := "📋 <b>Ваши подписки:</b>\n\n"
	msg += "┌────────────────────────────────────┐\n"
	var keyboard [][]models.InlineKeyboardButton
	for i, sub := range subs {
		status := "✅"; statusText := "Активна"
		if sub.ExpireAt.Before(time.Now().Add(24*time.Hour)) { status = "⚠️"; statusText = "Истекает" }
		if sub.ExpireAt.Before(time.Now()) { status = "❌"; statusText = "Истекла" }
		msg += fmt.Sprintf("│ %s <b>%s</b>\n", status, sub.Name)
		msg += fmt.Sprintf("│ 📅 %s\n", sub.ExpireAt.Format("02.01.2006 15:04"))
		msg += fmt.Sprintf("│ 🟢 %s\n", statusText)
		if i < len(subs)-1 { msg += "├────────────────────────────────────┤\n" }
		row := []models.InlineKeyboardButton{{ Text: fmt.Sprintf("🔗 %s", sub.Name), URL: sub.SubscriptionLink }}
		if sub.ExpireAt.After(time.Now()) {
			row = append(row, models.InlineKeyboardButton{ Text: "✏️ Переименовать", CallbackData: fmt.Sprintf("%s?id=%d", CallbackRenameSubscription, sub.ID) })
			row = append(row, models.InlineKeyboardButton{ Text: "🗑 "+h.translation.GetText(lang, "deactivate_button"), CallbackData: fmt.Sprintf("%s?id=%d", CallbackDeactivateSubscription, sub.ID) })
		}
		keyboard = append(keyboard, row)
	}
	msg += "└────────────────────────────────────┘\n\n"
	keyboard = append(keyboard, []models.InlineKeyboardButton{{ Text: h.translation.GetText(lang, "add_subscription_button"), CallbackData: CallbackTrial }})
	keyboard = append(keyboard, []models.InlineKeyboardButton{{ Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackStart }})

	_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{ ChatID: chatID, MessageID: messageID, ParseMode: models.ParseModeHTML, Text: msg, ReplyMarkup: models.InlineKeyboardMarkup{ InlineKeyboard: keyboard } })
}
