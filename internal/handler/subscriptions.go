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

// pendingRenames хранит соответствие chatID -> subscriptionID для ожидающего переименования
var pendingRenames = make(map[int64]int64)

// parseCallbackData parses callback data in format "action?key1=value1&key2=value2"
func parseCallbackData(callbackData string) map[string]string {
	result := make(map[string]string)
	parts := strings.SplitN(callbackData, "?", 2)
	if len(parts) < 2 { return result }
	values, err := url.ParseQuery(parts[1]); if err != nil { return result }
	for k, vals := range values { if len(vals) > 0 { result[k] = vals[0] } }
	return result
}

// MySubscriptionsCallbackHandler: компактный список (по одной кнопке на строку)
func (h Handler) MySubscriptionsCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery.Message.Message
	langCode := update.CallbackQuery.From.LanguageCode
	chatID := callback.Chat.ID

	customer, err := h.customerRepository.FindByTelegramId(ctx, chatID); if err != nil || customer == nil { return }
	activeSubscriptions, err := h.subscriptionRepository.GetActiveSubscriptions(ctx, customer.ID); if err != nil { return }

	messageText := "📋 <b>Ваши подписки</b>\n\n"
	var keyboard [][]models.InlineKeyboardButton
	for _, sub := range activeSubscriptions {
		label := fmt.Sprintf("📦 %s", sub.Name)
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{ Text: label, CallbackData: fmt.Sprintf("%s?id=%d", CallbackOpenSubscription, sub.ID) },
		})
	}
	keyboard = append(keyboard, []models.InlineKeyboardButton{{ Text: h.translation.GetText(langCode, "add_subscription_button"), CallbackData: CallbackTrial }})
	keyboard = append(keyboard, []models.InlineKeyboardButton{{ Text: h.translation.GetText(langCode, "back_button"), CallbackData: CallbackStart }})

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{ ChatID: chatID, MessageID: callback.ID, ParseMode: models.ParseModeHTML, ReplyMarkup: models.InlineKeyboardMarkup{ InlineKeyboard: keyboard }, Text: messageText })
	if err != nil { slog.Error("Error editing message", "error", err) }
}

// OpenSubscriptionCallbackHandler: карточка подписки с действиями
func (h Handler) OpenSubscriptionCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery.Message.Message
	langCode := update.CallbackQuery.From.LanguageCode
	chatID := callback.Chat.ID
	q := parseCallbackData(update.CallbackQuery.Data)
	idStr, ok := q["id"]; if !ok { return }
	subID, err := strconv.ParseInt(idStr, 10, 64); if err != nil { return }

	subscription, err := h.subscriptionRepository.GetSubscriptionByID(ctx, subID); if err != nil || subscription == nil { return }

	status := "✅ Активна"
	if subscription.ExpireAt.Before(time.Now()) { status = "❌ Истекла" } else if subscription.ExpireAt.Before(time.Now().Add(24*time.Hour)) { status = "⚠️ Истекает" }
	messageText := fmt.Sprintf("<b>%s</b>\n📅 %s\n%s", subscription.Name, subscription.ExpireAt.Format("02.01.2006 15:04"), status)

	var keyboard [][]models.InlineKeyboardButton
	keyboard = append(keyboard, []models.InlineKeyboardButton{{ Text: fmt.Sprintf("📱 %s", subscription.Name), URL: subscription.SubscriptionLink }})
	keyboard = append(keyboard, []models.InlineKeyboardButton{{ Text: "✏️ Переименовать", CallbackData: fmt.Sprintf("%s?id=%d", CallbackRenameSubscription, subscription.ID) }})
	if subscription.ExpireAt.After(time.Now()) {
		keyboard = append(keyboard, []models.InlineKeyboardButton{{ Text: fmt.Sprintf("🗑 %s", h.translation.GetText(langCode, "deactivate_button")), CallbackData: fmt.Sprintf("%s?id=%d", CallbackDeactivateSubscription, subscription.ID) }})
	}
	keyboard = append(keyboard, []models.InlineKeyboardButton{{ Text: "⬅️ "+h.translation.GetText(langCode, "my_subscriptions_button"), CallbackData: CallbackMySubscriptions }})

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{ ChatID: chatID, MessageID: callback.ID, ParseMode: models.ParseModeHTML, Text: messageText, ReplyMarkup: models.InlineKeyboardMarkup{ InlineKeyboard: keyboard } })
	if err != nil { slog.Error("Error editing message", "error", err) }
}

// DeactivateSubscriptionCallbackHandler и Rename остаются без изменений ниже...

func (h Handler) RenameSubscriptionCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery.Message.Message
	chatID := callback.Chat.ID
	callbackQuery := parseCallbackData(update.CallbackQuery.Data)
	subscriptionIDStr, exists := callbackQuery["id"]; if !exists { slog.Error("Subscription ID not found in callback data"); return }
	subscriptionID, err := strconv.ParseInt(subscriptionIDStr, 10, 64); if err != nil { slog.Error("Error parsing subscription ID", "error", err); return }
	pendingRenames[chatID] = subscriptionID
	text := "✏️ <b>Переименование подписки</b>\n\nОтправьте новое имя одним сообщением (до 50 символов).\n\n❕ Спецсимволы &lt; &gt; \" ' & запрещены."
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{ ChatID: callback.Chat.ID, MessageID: callback.ID, ParseMode: models.ParseModeHTML, ReplyMarkup: models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{ {{Text: "❌ Отмена", CallbackData: CallbackMySubscriptions}}, }}, Text: text })
	if err != nil { slog.Error("Error editing rename prompt", "error", err) }
}

func (h Handler) DeactivateSubscriptionCallbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery.Message.Message
	langCode := update.CallbackQuery.From.LanguageCode
	chatID := callback.Chat.ID
	callbackQuery := parseCallbackData(update.CallbackQuery.Data)
	subscriptionIDStr, exists := callbackQuery["id"]; if !exists { slog.Error("Subscription ID not found in callback data"); return }
	subscriptionID, err := strconv.ParseInt(subscriptionIDStr, 10, 64); if err != nil { slog.Error("Error parsing subscription ID", "error", err); return }
	customer, err := h.customerRepository.FindByTelegramId(ctx, chatID); if err != nil || customer == nil { slog.Error("Customer not found", "chatID", chatID); return }
	subscription, err := h.subscriptionRepository.GetSubscriptionByID(ctx, subscriptionID); if err != nil || subscription == nil { slog.Error("Subscription not found", "subscriptionID", subscriptionID); return }
	if subscription.CustomerID != customer.ID { slog.Error("Subscription doesn't belong to this customer", "subscriptionID", subscriptionID, "customerID", customer.ID); return }
	if err = h.subscriptionRepository.DeactivateSubscription(ctx, subscriptionID); err != nil { slog.Error("Error deactivating subscription", "error", err, "subscriptionID", subscriptionID); return }
	successText := fmt.Sprintf("✅ <b>Подписка деактивирована</b>\n\n🗑 %s успешно отключена", subscription.Name)
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{ ChatID: callback.Chat.ID, MessageID: callback.ID, ParseMode: models.ParseModeHTML, ReplyMarkup: models.InlineKeyboardMarkup{ InlineKeyboard: [][]models.InlineKeyboardButton{ {{Text: h.translation.GetText(langCode, "my_subscriptions_button"), CallbackData: CallbackMySubscriptions}}, {{Text: h.translation.GetText(langCode, "back_button"), CallbackData: CallbackStart}}, } }, Text: successText })
	if err != nil { slog.Error("Error editing message", "error", err) }
}

func (h Handler) GetActiveSubscriptionsCount(ctx context.Context, customerID int64) (int, error) {
	activeSubscriptions, err := h.subscriptionRepository.GetActiveSubscriptions(ctx, customerID); if err != nil { return 0, err }
	return len(activeSubscriptions), nil
}

func (h Handler) GetSubscriptionsList(ctx context.Context, customer *database.Customer, langCode string) (string, [][]models.InlineKeyboardButton) {
	messageText := h.translation.GetText(langCode, "your_active_subscriptions") + "\n\n"
	return messageText, [][]models.InlineKeyboardButton{{{ Text: h.translation.GetText(langCode, "back_button"), CallbackData: CallbackStart }}}
}

func max(a, b int) int { if a > b { return a }; return b }
