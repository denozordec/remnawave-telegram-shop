package handler

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// afterSubscriptionCreated: просто вызывает handler, имитируя callback "Мои подписки"
func (h Handler) afterSubscriptionCreated(ctx context.Context, b *bot.Bot, chatID int64, messageID int) {
	// Вместо сборки сложных структур — напрямую вызовем handler с текущими данными, он сам перечитает БД и перерисует.
	update := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			From:    models.User{ID: chatID},
			Message: models.Message{Chat: &models.Chat{ID: chatID}, ID: messageID},
			Data:    CallbackMySubscriptions,
		},
	}
	h.MySubscriptionsCallbackHandler(ctx, b, update)
}
