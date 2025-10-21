package handler

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h Handler) ConnectCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil { return }
	upd := &models.Update{ CallbackQuery: &models.CallbackQuery{ From: *update.Message.From, Message: *update.Message, Data: CallbackMySubscriptions } }
	h.MySubscriptionsCallbackHandler(ctx, b, upd)
}

func (h Handler) ConnectCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil { return }
	update.CallbackQuery.Data = CallbackMySubscriptions
	h.MySubscriptionsCallbackHandler(ctx, b, update)
}
