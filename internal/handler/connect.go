package handler

import (
	"context"
	"fmt"
	"remnawave-tg-shop-bot/internal/config"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"

	"remnawave-tg-shop-bot/internal/database"
	"remnawave-tg-shop-bot/internal/translation"
	"remnawave-tg-shop-bot/utils"
)

func (h Handler) ConnectCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	customer, err := h.customerRepository.FindByTelegramId(ctx, update.Message.Chat.ID)
	if err != nil {
		slog.Error("Error finding customer", err)
		return
	}
	if customer == nil {
		slog.Error("customer not exist", "telegramId", utils.MaskHalfInt64(update.Message.Chat.ID), "error", err)
		return
	}

	langCode := update.Message.From.LanguageCode

	// Получаем список подписок и клавиатуру
	messageText, keyboard := h.GetSubscriptionsList(ctx, customer, langCode)

	isDisabled := true
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      messageText,
		ParseMode: models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: &isDisabled,
		},
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	})

	if err != nil {
		slog.Error("Error sending connect message", err)
	}
}

func (h Handler) ConnectCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery.Message.Message

	customer, err := h.customerRepository.FindByTelegramId(ctx, callback.Chat.ID)
	if err != nil {
		slog.Error("Error finding customer", err)
		return
	}
	if customer == nil {
		slog.Error("customer not exist", "telegramId", utils.MaskHalfInt64(callback.Chat.ID), "error", err)
		return
	}

	langCode := update.CallbackQuery.From.LanguageCode

	// Получаем список подписок и клавиатуру
	messageText, keyboard := h.GetSubscriptionsList(ctx, customer, langCode)

	isDisabled := true
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Chat.ID,
		MessageID: callback.ID,
		ParseMode: models.ParseModeHTML,
		Text:      messageText,
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: &isDisabled,
		},
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	})

	if err != nil {
		slog.Error("Error sending connect message", err)
	}
}

// Оставляем старую функцию для обратной совместимости
func buildConnectText(customer *database.Customer, langCode string, update *models.Update) string {
	var info strings.Builder

	tm := translation.GetInstance()

	if customer.ExpireAt != nil {
		currentTime := time.Now()

		if currentTime.Before(*customer.ExpireAt) {
			formattedDate := customer.ExpireAt.Format("02.01.2006 15:04")

			subscriptionActiveText := tm.GetText(langCode, "subscription_active")
			info.WriteString(fmt.Sprintf(subscriptionActiveText, formattedDate))

			if customer.SubscriptionLink != nil && *customer.SubscriptionLink != "" {
				platform := config.DetectPlatformFromUpdate(update)
				if config.IsWebAppLinkEnabledForPlatform(platform) {
					// Для мобильных устройств не показываем ссылку в тексте
				} else {
					subscriptionLinkText := tm.GetText(langCode, "subscription_link")
					info.WriteString(fmt.Sprintf(subscriptionLinkText, *customer.SubscriptionLink))
				}
			}
		} else {
			noSubscriptionText := tm.GetText(langCode, "no_subscription")
			info.WriteString(noSubscriptionText)
		}
	} else {
		noSubscriptionText := tm.GetText(langCode, "no_subscription")
		info.WriteString(noSubscriptionText)
	}

	return info.String()
}
