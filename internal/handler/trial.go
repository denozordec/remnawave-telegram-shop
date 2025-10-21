package handler

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"remnawave-tg-shop-bot/internal/subscriptions"
)

func (h Handler) TrialCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Всегда создаём бесплатную подписку через free service
	svc := &subscriptions.Service{
		SubsRepo:  h.subscriptionRepository,
		Customers: h.customerRepository,
		RW:        h.syncService.RemnawaveClient(),
		Bot:       b,
		Translate: h.translation,
	}
	callback := update.CallbackQuery.Message.Message
	_, err := svc.ActivateFree(context.WithValue(ctx, "username", update.CallbackQuery.From.Username), update.CallbackQuery.From.ID)
	langCode := update.CallbackQuery.From.LanguageCode
	if err != nil {
		slog.Error("Error activating free subscription", "err", err)
	}
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      callback.Chat.ID,
		MessageID:   callback.ID,
		Text:        h.translation.GetText(langCode, "trial_activated"),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: h.createConnectKeyboard(langCode)},
	})
	if err != nil {
		slog.Error("Error sending trial activation message", err)
	}
}
