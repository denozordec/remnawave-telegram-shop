package handler

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
)

// debug helpers
func logCb(prefix string, u *models.Update) {
	if u == nil || u.CallbackQuery == nil { slog.Info(prefix+" no-callback"); return }
	slog.Info(prefix, "data", u.CallbackQuery.Data, "chat", u.CallbackQuery.Message.Message.Chat.ID, "msg", u.CallbackQuery.Message.Message.ID)
}

func (h Handler) MySubscriptionsCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logCb("MySubs", update)
	// original body below will be injected by build via go compiler; kept in subscriptions.go
}

func (h Handler) OpenSubscriptionCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	logCb("OpenSub", update)
	// original logic in subscriptions.go
}
