package handler

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"remnawave-tg-shop-bot/internal/config"
	"remnawave-tg-shop-bot/internal/subscriptions"
)

func (h Handler) TrialCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if config.TrialDays() == 0 {
		return
	}
	// Всегда создаём бесплатную подписку через free service
	svc := &subscriptions.Service{SubsRepo: h.subscriptionRepository, Customers: h.customerRepository, RW: h.syncService.GetClient(), Translate: h.translation}
	callback := update.CallbackQuery.Message.Message
	_, err := svc.ActivateFree(context.WithValue(ctx, "username", update.CallbackQuery.From.Username), update.CallbackQuery.From.ID)
	langCode := update.CallbackQuery.From.LanguageCode
	if err != nil {
		slog.Error("Error activating free subscription", "err", err)
	}
	// сразу рендерим красивую таблицу
	h.afterSubscriptionCreated(ctx, b, callback.Chat.ID, callback.ID)
	// если что-то пойдёт не так, покажем запасной текст
	_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{ChatID: callback.Chat.ID, MessageID: callback.ID, Text: h.translation.GetText(langCode, "trial_activated"), ParseMode: models.ParseModeHTML, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: h.createConnectKeyboard(langCode)}})
}

