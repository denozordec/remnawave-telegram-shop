package notification

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"remnawave-tg-shop-bot/internal/config"
	"remnawave-tg-shop-bot/internal/database"
	"remnawave-tg-shop-bot/internal/translation"
	"remnawave-tg-shop-bot/utils"
)

type BroadcastService struct {
	customerRepository *database.CustomerRepository
	telegramBot        *bot.Bot
	tm                 *translation.Manager
}

func NewBroadcastService(customerRepository *database.CustomerRepository, telegramBot *bot.Bot, tm *translation.Manager) *BroadcastService {
	return &BroadcastService{
		customerRepository: customerRepository,
		telegramBot:        telegramBot,
		tm:                 tm,
	}
}

// SendBroadcastToAll отправляет сообщение всем пользователям бота
func (s *BroadcastService) SendBroadcastToAll(ctx context.Context, message string, parseMode string) error {
	customers, err := s.customerRepository.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all customers: %w", err)
	}

	slog.Info("Starting broadcast to all users", "total_users", len(customers))

	successCount := 0
	failedCount := 0

	for _, customer := range customers {
		err := s.sendMessageToCustomer(ctx, customer, message, parseMode)
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

// SendBroadcastToAdmins отправляет сообщение только админам
func (s *BroadcastService) SendBroadcastToAdmins(ctx context.Context, message string, parseMode string) error {
	adminTelegramID := config.GetAdminTelegramId()
	
	// Получаем всех админов (в данном случае только одного, но можно расширить)
	adminCustomer, err := s.customerRepository.FindByTelegramId(ctx, adminTelegramID)
	if err != nil {
		return fmt.Errorf("failed to find admin customer: %w", err)
	}

	if adminCustomer == nil {
		return fmt.Errorf("admin customer not found")
	}

	err = s.sendMessageToCustomer(ctx, *adminCustomer, message, parseMode)
	if err != nil {
		return fmt.Errorf("failed to send message to admin: %w", err)
	}

	slog.Info("Broadcast to admins completed", "admin_id", utils.MaskHalfInt64(adminCustomer.ID))
	return nil
}

// sendMessageToCustomer отправляет сообщение конкретному пользователю
func (s *BroadcastService) sendMessageToCustomer(ctx context.Context, customer database.Customer, message string, parseMode string) error {
	_, err := s.telegramBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    customer.TelegramID,
		Text:      message,
		ParseMode: parseMode,
	})

	return err
} 