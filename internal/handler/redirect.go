package handler

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// afterSubscriptionCreated: напрямую вызывает рендер через существующий handler
func (h Handler) afterSubscriptionCreated(ctx context.Context, b *bot.Bot, chatID int64, messageID int) {
	upd := &models.Update{CallbackQuery: &models.CallbackQuery{From: models.User{ID: chatID}, Message: models.Message{Chat: &models.Chat{ID: chatID}, ID: messageID}, Data: CallbackMySubscriptions}}
	h.MySubscriptionsCallbackHandler(ctx, b, upd)
}
