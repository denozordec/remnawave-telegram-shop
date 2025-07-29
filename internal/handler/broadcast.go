package handler

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"remnawave-tg-shop-bot/internal/config"
	"remnawave-tg-shop-bot/internal/database"
	"remnawave-tg-shop-bot/utils"
	"strings"
)

// BroadcastMenuHandler показывает меню рассылки для админов
func (h Handler) BroadcastMenuHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := "📢 <b>Меню рассылки</b>\n\nВыберите тип рассылки:"
	
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      message,
		ParseMode: models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "📤 Всем пользователям", CallbackData: CallbackBroadcastToAll},
				},
				{
					{Text: "👥 Только админам", CallbackData: CallbackBroadcastToAdmins},
				},
			},
		},
	})
	
	if err != nil {
		slog.Error("Error sending broadcast menu", err)
	}
}

// BroadcastTypeHandler обрабатывает выбор типа рассылки
func (h Handler) BroadcastTypeHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery
	callbackData := callback.Data
	
	var broadcastType int
	var buttonText string
	
	switch callbackData {
	case CallbackBroadcastToAll:
		broadcastType = 1
		buttonText = "📤 Всем пользователям"
	case CallbackBroadcastToAdmins:
		broadcastType = 2
		buttonText = "👥 Только админам"
	default:
		return
	}
	
	// Сохраняем тип рассылки в кэше
	h.cache.Set(callback.From.ID, broadcastType)
	
	message := fmt.Sprintf("📢 <b>Рассылка: %s</b>\n\nОтправьте сообщение для рассылки:", buttonText)
	
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Message.Message.Chat.ID,
		MessageID: callback.Message.Message.ID,
		Text:      message,
		ParseMode: models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "❌ Отмена", CallbackData: CallbackBroadcastCancel},
				},
			},
		},
	})
	
	if err != nil {
		slog.Error("Error editing broadcast type message", err)
	}
	
	// Отвечаем на callback query
	_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
	})
	if err != nil {
		slog.Error("Error answering callback query", err)
	}
}

// BroadcastMessageHandler обрабатывает сообщение для рассылки
func (h Handler) BroadcastMessageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Проверяем, что это админ
	if update.Message.From.ID != config.GetAdminTelegramId() {
		return
	}
	
	// Получаем тип рассылки из кэша
	broadcastType, exists := h.cache.Get(update.Message.From.ID)
	if !exists {
		return
	}
	
	message := update.Message.Text
	if message == "" {
		// Если нет текста, игнорируем
		return
	}
	
	// Удаляем тип рассылки из кэша (просто игнорируем)
	_ = broadcastType
	
	// Создаем клавиатуру подтверждения
	var broadcastTypeStr string
	switch broadcastType {
	case 1:
		broadcastTypeStr = "all"
	case 2:
		broadcastTypeStr = "admins"
	default:
		broadcastTypeStr = "all"
	}
	
	confirmKeyboard := [][]models.InlineKeyboardButton{
		{
			{Text: "✅ Подтвердить", CallbackData: CallbackBroadcastConfirm + ":" + broadcastTypeStr},
			{Text: "❌ Отмена", CallbackData: CallbackBroadcastCancel},
		},
	}
	
	var broadcastTypeText string
	switch broadcastType {
	case 1:
		broadcastTypeText = "📤 Всем пользователям"
	case 2:
		broadcastTypeText = "👥 Только админам"
	default:
		broadcastTypeText = "📤 Всем пользователям"
	}
	
	previewMessage := fmt.Sprintf("📢 <b>Предварительный просмотр рассылки</b>\n\n<b>Тип:</b> %s\n<b>Сообщение:</b>\n\n%s", broadcastTypeText, message)
	
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      previewMessage,
		ParseMode: models.ParseModeHTML,
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: confirmKeyboard,
		},
	})
	
	if err != nil {
		slog.Error("Error sending broadcast preview", err)
	}
}

// BroadcastConfirmHandler подтверждает и выполняет рассылку
func (h Handler) BroadcastConfirmHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery
	callbackData := callback.Data
	
	// Извлекаем тип рассылки из callback data
	parts := strings.Split(callbackData, ":")
	if len(parts) != 2 {
		return
	}
	
	broadcastTypeStr := parts[1]
	var broadcastType int
	switch broadcastTypeStr {
	case "all":
		broadcastType = 1
	case "admins":
		broadcastType = 2
	default:
		return
	}
	
	// Получаем сообщение из предыдущего сообщения
	messageText := ""
	if callback.Message.Message != nil && callback.Message.Message.Text != "" {
		// Извлекаем текст сообщения из предварительного просмотра
		lines := strings.Split(callback.Message.Message.Text, "\n")
		for i, line := range lines {
			if strings.Contains(line, "Сообщение:") {
				if i+1 < len(lines) {
					messageText = strings.Join(lines[i+1:], "\n")
					break
				}
			}
		}
	}
	
	if messageText == "" {
		_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "❌ Ошибка: не удалось получить текст сообщения",
		})
		if err != nil {
			slog.Error("Error answering callback query", err)
		}
		return
	}
	
	// Отправляем рассылку
	var err error
	switch broadcastType {
	case 1:
		err = h.sendBroadcastToAll(ctx, b, messageText, string(models.ParseModeHTML))
	case 2:
		err = h.sendBroadcastToAdmins(ctx, b, messageText, string(models.ParseModeHTML))
	}
	
	if err != nil {
		slog.Error("Error sending broadcast", err)
		_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "❌ Ошибка при отправке рассылки",
		})
		if err != nil {
			slog.Error("Error answering callback query", err)
		}
		return
	}
	
	// Обновляем сообщение с результатом
	var broadcastTypeText string
	switch broadcastType {
	case 1:
		broadcastTypeText = "📤 Всем пользователям"
	case 2:
		broadcastTypeText = "👥 Только админам"
	default:
		broadcastTypeText = "📤 Всем пользователям"
	}
	resultMessage := fmt.Sprintf("✅ <b>Рассылка выполнена успешно!</b>\n\nТип: %s\nСообщение отправлено.", broadcastTypeText)
	
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Message.Message.Chat.ID,
		MessageID: callback.Message.Message.ID,
		Text:      resultMessage,
		ParseMode: models.ParseModeHTML,
	})
	
	if err != nil {
		slog.Error("Error editing broadcast result message", err)
	}
	
	// Отвечаем на callback query
	_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
		Text:            "✅ Рассылка выполнена",
	})
	if err != nil {
		slog.Error("Error answering callback query", err)
	}
}

// BroadcastCancelHandler отменяет рассылку
func (h Handler) BroadcastCancelHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	callback := update.CallbackQuery
	
	// Удаляем тип рассылки из кэша (просто игнорируем)
	_ = callback.From.ID
	
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Message.Message.Chat.ID,
		MessageID: callback.Message.Message.ID,
		Text:      "❌ <b>Рассылка отменена</b>",
		ParseMode: models.ParseModeHTML,
	})
	
	if err != nil {
		slog.Error("Error editing broadcast cancel message", err)
	}
	
	// Отвечаем на callback query
	_, err = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
		Text:            "❌ Рассылка отменена",
	})
	if err != nil {
		slog.Error("Error answering callback query", err)
	}
}

// sendBroadcastToAll отправляет сообщение всем пользователям
func (h Handler) sendBroadcastToAll(ctx context.Context, b *bot.Bot, message string, parseMode string) error {
	customers, err := h.customerRepository.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all customers: %w", err)
	}

	slog.Info("Starting broadcast to all users", "total_users", len(customers))

	successCount := 0
	failedCount := 0

	for _, customer := range customers {
		err := h.sendMessageToCustomer(ctx, b, customer, message, parseMode)
		if err != nil {
			slog.Error("Failed to send broadcast message",
				"customer_id", utils.MaskHalfInt64(customer.ID),
				"telegram_id", utils.MaskHalfInt64(customer.TelegramID),
				"error", err)
			failedCount++
		} else {
			successCount++
		}
	}

	slog.Info("Broadcast completed",
		"total_users", len(customers),
		"success_count", successCount,
		"failed_count", failedCount)

	return nil
}

// sendBroadcastToAdmins отправляет сообщение только админам
func (h Handler) sendBroadcastToAdmins(ctx context.Context, b *bot.Bot, message string, parseMode string) error {
	adminTelegramID := config.GetAdminTelegramId()
	
	// Получаем админа
	adminCustomer, err := h.customerRepository.FindByTelegramId(ctx, adminTelegramID)
	if err != nil {
		return fmt.Errorf("failed to find admin customer: %w", err)
	}

	if adminCustomer == nil {
		return fmt.Errorf("admin customer not found")
	}

	err = h.sendMessageToCustomer(ctx, b, *adminCustomer, message, parseMode)
	if err != nil {
		return fmt.Errorf("failed to send message to admin: %w", err)
	}

	slog.Info("Broadcast to admins completed", "admin_id", utils.MaskHalfInt64(adminCustomer.ID))
	return nil
}

// sendMessageToCustomer отправляет сообщение конкретному пользователю
func (h Handler) sendMessageToCustomer(ctx context.Context, b *bot.Bot, customer database.Customer, message string, parseMode string) error {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    customer.TelegramID,
		Text:      message,
		ParseMode: models.ParseMode(parseMode),
	})

	return err
} 