package handler

import (
	"context"
	"regexp"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// TextMessageHandler: если ожидается переименование, принимает новое имя; иначе ничего не делает
func (h Handler) TextMessageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil { return }
	chatID := update.Message.Chat.ID
	newName := strings.TrimSpace(update.Message.Text)

	// Если нет ожидания — выходим
	subID, ok := pendingRenames[chatID]
	if !ok { return }

	// Валидация
	if len(newName) < 1 || len(newName) > 50 { 
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ ChatID: chatID, ParseMode: models.ParseModeHTML, Text: "⚠️ Имя должно быть от 1 до 50 символов. Попробуйте снова." })
		return 
	}
	if regexp.MustCompile(`[<>"'&]`).MatchString(newName) {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ ChatID: chatID, ParseMode: models.ParseModeHTML, Text: "⚠️ Имя не должно содержать символы: < > \" ' &" })
		return
	}

	// Получаем клиента и подписку, обновляем имя
	customer, err := h.customerRepository.FindByTelegramId(ctx, chatID); if err != nil || customer == nil { delete(pendingRenames, chatID); return }
	sub, err := h.subscriptionRepository.GetSubscriptionByID(ctx, subID); if err != nil || sub == nil { delete(pendingRenames, chatID); return }
	if sub.CustomerID != customer.ID { delete(pendingRenames, chatID); return }
	_ = h.subscriptionRepository.UpdateSubscriptionName(ctx, subID, newName)
	delete(pendingRenames, chatID)
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{ ChatID: chatID, ParseMode: models.ParseModeHTML, Text: "✅ Имя обновлено!", ReplyMarkup: models.InlineKeyboardMarkup{ InlineKeyboard: [][]models.InlineKeyboardButton{ { {Text:"📋 Мои подписки", CallbackData: CallbackMySubscriptions} }, { {Text:"⬅️ Назад", CallbackData: CallbackStart} }, } } })
}
