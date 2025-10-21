package handler

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h Handler) ConnectCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil { return }
	// Прямой вызов логики рендера подписок для чата
	h.renderMySubscriptionsForChat(ctx, b, update.Message.Chat.ID, update.Message.ID, update.Message.From.LanguageCode)
}

func (h Handler) ConnectCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil { return }
	msg := update.CallbackQuery.Message.Message
	h.renderMySubscriptionsForChat(ctx, b, msg.Chat.ID, msg.ID, update.CallbackQuery.From.LanguageCode)
}
