package handler

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"

	"remnawave-tg-shop-bot/internal/database"
)

// parseCallbackData parses callback data in format "action?key1=value1&key2=value2"
func parseCallbackData(callbackData string) map[string]string {
	result := make(map[string]string)
	
	parts := strings.SplitN(callbackData, "?", 2)
	if len(parts) < 2 {
		return result
	}
	
	queryString := parts[1]
	values, err := url.ParseQuery(queryString)
	if err != nil {
		return result
	}
	
	for key, vals := range values {
		if len(vals) > 0 {
			result[key] = vals[0]
		}
	}
	
	return result
}

// MySubscriptionsCallbackHandler обрабатывает запрос на показ всех подписок пользователя
func (h Handler) MySubscriptionsCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery.Message.Message
	langCode := update.CallbackQuery.From.LanguageCode
	chatID := callback.Chat.ID

	// Получаем клиента из базы данных
	customer, err := h.customerRepository.FindByTelegramId(ctx, chatID)
	if err != nil {
		slog.Error("Error finding customer", "error", err, "chatID", chatID)
		return
	}
	if customer == nil {
		slog.Error("Customer not found", "chatID", chatID)
		return
	}

	// Получаем все активные подписки
	activeSubscriptions, err := h.subscriptionRepository.GetActiveSubscriptions(ctx, customer.ID)
	if err != nil {
		slog.Error("Error getting active subscriptions", "error", err, "customerID", customer.ID)
		return
	}

	var messageText string
	var keyboard [][]models.InlineKeyboardButton

	if len(activeSubscriptions) == 0 {
		// Если нет активных подписок
		messageText = h.translation.GetText(langCode, "no_active_subscriptions")
		keyboard = [][]models.InlineKeyboardButton{
			{{Text: h.translation.GetText(langCode, "buy_button"), CallbackData: CallbackTrial}},
			{{Text: h.translation.GetText(langCode, "back_button"), CallbackData: CallbackStart}},
		}
	} else {
		// Формируем список подписок
		messageText = h.translation.GetText(langCode, "your_subscriptions")
		messageText += "\n\n"

		for _, sub := range activeSubscriptions {
			// Форматируем дату истечения
			expireDate := sub.ExpireAt.Format("02.01.2006 15:04")
			
			// Определяем статус (активна/истекает)
			status := "✅"
			if sub.ExpireAt.Before(time.Now().Add(24 * time.Hour)) {
				status = "⚠️" // Истекает в течение 24 часов
			}
			if sub.ExpireAt.Before(time.Now()) {
				status = "❌" // Истекла
			}

			messageText += fmt.Sprintf("%s <b>%s</b>\n%s\n%s %s\n\n",
				status,
				sub.Name,
				sub.Description,
				h.translation.GetText(langCode, "expires_at"),
				expireDate,
			)

			// Создаем кнопки для каждой подписки
			subscriptionButtons := []models.InlineKeyboardButton{
				{Text: fmt.Sprintf("🔗 %s", sub.Name), URL: sub.SubscriptionLink},
			}

			// Добавляем кнопку деактивации (только для активных)
			if sub.ExpireAt.After(time.Now()) {
				subscriptionButtons = append(subscriptionButtons,
					models.InlineKeyboardButton{
						Text:         fmt.Sprintf("🗑 %s", h.translation.GetText(langCode, "deactivate_button")),
						CallbackData: fmt.Sprintf("%s?id=%d", CallbackDeactivateSubscription, sub.ID),
					},
				)
			}

			keyboard = append(keyboard, subscriptionButtons)
		}

		// Добавляем общие кнопки управления
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{Text: h.translation.GetText(langCode, "add_subscription_button"), CallbackData: CallbackTrial},
		})
	}

	// Добавляем кнопку "Назад"
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: h.translation.GetText(langCode, "back_button"), CallbackData: CallbackStart},
	})

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Chat.ID,
		MessageID: callback.ID,
		ParseMode: models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
		Text: messageText,
	})

	if err != nil {
		slog.Error("Error editing message", "error", err)
	}
}

// DeactivateSubscriptionCallbackHandler обрабатывает деактивацию подписки
func (h Handler) DeactivateSubscriptionCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery.Message.Message
	langCode := update.CallbackQuery.From.LanguageCode
	chatID := callback.Chat.ID

	// Парсим ID подписки из callback data
	callbackQuery := parseCallbackData(update.CallbackQuery.Data)
	subscriptionIDStr, exists := callbackQuery["id"]
	if !exists {
		slog.Error("Subscription ID not found in callback data")
		return
	}

	subscriptionID, err := strconv.ParseInt(subscriptionIDStr, 10, 64)
	if err != nil {
		slog.Error("Error parsing subscription ID", "error", err)
		return
	}

	// Получаем клиента из базы данных
	customer, err := h.customerRepository.FindByTelegramId(ctx, chatID)
	if err != nil {
		slog.Error("Error finding customer", "error", err, "chatID", chatID)
		return
	}
	if customer == nil {
		slog.Error("Customer not found", "chatID", chatID)
		return
	}

	// Получаем подписку для проверки владения
	subscription, err := h.subscriptionRepository.GetSubscriptionByID(ctx, subscriptionID)
	if err != nil {
		slog.Error("Error getting subscription", "error", err, "subscriptionID", subscriptionID)
		return
	}
	if subscription == nil {
		slog.Error("Subscription not found", "subscriptionID", subscriptionID)
		return
	}

	// Проверяем, что подписка принадлежит этому клиенту
	if subscription.CustomerID != customer.ID {
		slog.Error("Subscription doesn't belong to this customer", "subscriptionID", subscriptionID, "customerID", customer.ID)
		return
	}

	// Деактивируем подписку
	err = h.subscriptionRepository.DeactivateSubscription(ctx, subscriptionID)
	if err != nil {
		slog.Error("Error deactivating subscription", "error", err, "subscriptionID", subscriptionID)
		return
	}

	// Отправляем сообщение об успешной деактивации
	successText := fmt.Sprintf(h.translation.GetText(langCode, "subscription_deactivated"), subscription.Name)

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Chat.ID,
		MessageID: callback.ID,
		ParseMode: models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: h.translation.GetText(langCode, "my_subscriptions_button"), CallbackData: CallbackMySubscriptions}},
				{{Text: h.translation.GetText(langCode, "back_button"), CallbackData: CallbackStart}},
			},
		},
		Text: successText,
	})

	if err != nil {
		slog.Error("Error editing message", "error", err)
	}

	slog.Info("Subscription deactivated", "subscriptionID", subscriptionID, "customerID", customer.ID)
}

// GetActiveSubscriptionsCount возвращает количество активных подписок для клиента
func (h Handler) GetActiveSubscriptionsCount(ctx context.Context, customerID int64) (int, error) {
	activeSubscriptions, err := h.subscriptionRepository.GetActiveSubscriptions(ctx, customerID)
	if err != nil {
		return 0, err
	}
	return len(activeSubscriptions), nil
}

// GetSubscriptionsList возвращает список подписок для отображения в connect handler
func (h Handler) GetSubscriptionsList(ctx context.Context, customer *database.Customer, langCode string) (string, [][]models.InlineKeyboardButton) {
	activeSubscriptions, err := h.subscriptionRepository.GetActiveSubscriptions(ctx, customer.ID)
	if err != nil {
		slog.Error("Error getting active subscriptions", "error", err)
		return h.translation.GetText(langCode, "error_getting_subscriptions"), [][]models.InlineKeyboardButton{
			{{Text: h.translation.GetText(langCode, "back_button"), CallbackData: CallbackStart}},
		}
	}

	if len(activeSubscriptions) == 0 {
		return h.translation.GetText(langCode, "no_active_subscriptions"), [][]models.InlineKeyboardButton{
			{{Text: h.translation.GetText(langCode, "buy_button"), CallbackData: CallbackTrial}},
			{{Text: h.translation.GetText(langCode, "back_button"), CallbackData: CallbackStart}},
		}
	}

	// Формируем текст со списком подписок
	messageText := h.translation.GetText(langCode, "your_active_subscriptions") + "\n\n"
	var keyboard [][]models.InlineKeyboardButton

	for _, sub := range activeSubscriptions {
		// Проверяем, не истекла ли подписка
		if sub.ExpireAt.After(time.Now()) {
			expireDate := sub.ExpireAt.Format("02.01.2006 15:04")
			messageText += fmt.Sprintf("🔗 <b>%s</b>\n%s %s\n\n", sub.Name, h.translation.GetText(langCode, "expires_at"), expireDate)
			
			// Добавляем кнопку для подключения к этой подписке
			keyboard = append(keyboard, []models.InlineKeyboardButton{
				{Text: fmt.Sprintf("📱 %s", sub.Name), URL: sub.SubscriptionLink},
			})
		}
	}

	// Добавляем кнопки управления
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: h.translation.GetText(langCode, "my_subscriptions_button"), CallbackData: CallbackMySubscriptions},
	})
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: h.translation.GetText(langCode, "add_subscription_button"), CallbackData: CallbackTrial},
	})
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{Text: h.translation.GetText(langCode, "back_button"), CallbackData: CallbackStart},
	})

	return messageText, keyboard
}
