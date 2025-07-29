package handler

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"remnawave-tg-shop-bot/internal/config"
	"remnawave-tg-shop-bot/internal/notification"
	"strings"
)

const (
	CallbackBroadcastMenu     = "broadcast_menu"
	CallbackBroadcastToAll    = "broadcast_to_all"
	CallbackBroadcastToAdmins = "broadcast_to_admins"
	CallbackBroadcastConfirm  = "broadcast_confirm"
	CallbackBroadcastCancel   = "broadcast_cancel"
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
	
	var broadcastType string
	var buttonText string
	
	switch callbackData {
	case CallbackBroadcastToAll:
		broadcastType = "all"
		buttonText = "📤 Всем пользователям"
	case CallbackBroadcastToAdmins:
		broadcastType = "admins"
		buttonText = "👥 Только админам"
	default:
		return
	}
	
	// Сохраняем тип рассылки в кэше или контексте
	// Для простоты используем простую переменную, но лучше использовать кэш
	h.cache.Set("broadcast_type_"+fmt.Sprint(callback.From.ID), broadcastType)
	
	message := fmt.Sprintf("📢 <b>Рассылка: %s</b>\n\nОтправьте сообщение для рассылки:", buttonText)
	
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Message.Chat.ID,
		MessageID: callback.Message.ID,
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
	broadcastType, exists := h.cache.Get("broadcast_type_" + fmt.Sprint(update.Message.From.ID))
	if !exists {
		return
	}
	
	message := update.Message.Text
	if message == "" {
		// Если нет текста, игнорируем
		return
	}
	
	// Удаляем тип рассылки из кэша
	h.cache.Delete("broadcast_type_" + fmt.Sprint(update.Message.From.ID))
	
	// Создаем клавиатуру подтверждения
	confirmKeyboard := [][]models.InlineKeyboardButton{
		{
			{Text: "✅ Подтвердить", CallbackData: CallbackBroadcastConfirm + ":" + broadcastType.(string)},
			{Text: "❌ Отмена", CallbackData: CallbackBroadcastCancel},
		},
	}
	
	var broadcastTypeText string
	switch broadcastType.(string) {
	case "all":
		broadcastTypeText = "📤 Всем пользователям"
	case "admins":
		broadcastTypeText = "👥 Только админам"
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
	
	broadcastType := parts[1]
	
	// Получаем сообщение из предыдущего сообщения
	messageText := ""
	if callback.Message != nil && callback.Message.Text != "" {
		// Извлекаем текст сообщения из предварительного просмотра
		lines := strings.Split(callback.Message.Text, "\n")
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
	
	// Создаем сервис рассылки
	broadcastService := notification.NewBroadcastService(h.customerRepository, b, h.translation)
	
	var err error
	switch broadcastType {
	case "all":
		err = broadcastService.SendBroadcastToAll(ctx, messageText, models.ParseModeHTML)
	case "admins":
		err = broadcastService.SendBroadcastToAdmins(ctx, messageText, models.ParseModeHTML)
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
	resultMessage := fmt.Sprintf("✅ <b>Рассылка выполнена успешно!</b>\n\nТип: %s\nСообщение отправлено.", 
		map[string]string{"all": "📤 Всем пользователям", "admins": "👥 Только админам"}[broadcastType])
	
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Message.Chat.ID,
		MessageID: callback.Message.ID,
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
	
	// Удаляем тип рассылки из кэша
	h.cache.Delete("broadcast_type_" + fmt.Sprint(callback.From.ID))
	
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    callback.Message.Chat.ID,
		MessageID: callback.Message.ID,
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