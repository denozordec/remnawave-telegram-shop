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
	_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{ChatID: callback.Chat.ID, MessageID: callback.ID, Text: h.translation.GetText(langCode, "trial_activated"), ParseMode: models.ParseModeHTML, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: h.createTrialSuccessKeyboard(langCode)}})
}

// createTrialSuccessKeyboard returns keyboard with Connect and Telegram buttons after trial activation
func (h Handler) createTrialSuccessKeyboard(lang string) [][]models.InlineKeyboardButton {
	var keyboard [][]models.InlineKeyboardButton

	// Кнопка "Подключиться"
	if config.GetMiniAppURL() != "" {
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{Text: h.translation.GetText(lang, "connect_button"), WebApp: &models.WebAppInfo{
				URL: config.GetMiniAppURL(),
			}},
		})
	} else {
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{Text: h.translation.GetText(lang, "connect_button"), CallbackData: CallbackConnect},
		})
	}

	// Кнопка "Оживить Telegram" если TG_PROXY_LINK задан
	if config.TgProxyLink() != "" {
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{Text: h.translation.GetText(lang, "tg_proxy_button"), URL: config.TgProxyLink()},
		})
	}

	// Кнопка "Мои подписки"
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: h.translation.GetText(lang, "my_subscriptions_button"), CallbackData: CallbackMySubscriptions},
	})

	// Кнопка "В главное меню"
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: h.translation.GetText(lang, "back_button"), CallbackData: CallbackStart},
	})

	return keyboard
}
