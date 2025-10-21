package handler

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
)

func (h Handler) ConnectCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Редиректим в раздел Мои подписки
	if update.Message == nil { return }
	u := &models.Update{ CallbackQuery: &models.CallbackQuery{ From: *update.Message.From, Message: *update.Message, Data: CallbackMySubscriptions } }
	h.MySubscriptionsCallbackHandler(ctx, b, u)
}

func (h Handler) ConnectCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Проксируем в раздел Мои подписки
	if update.CallbackQuery == nil { slog.Error("ConnectCallback without CallbackQuery") ; return }
	update.CallbackQuery.Data = CallbackMySubscriptions
	h.MySubscriptionsCallbackHandler(ctx, b, update)
}

// Оставляем старую функцию для обратной совместимости
func buildConnectText(customer *database.Customer, langCode string, update *models.Update) string { return "" }
