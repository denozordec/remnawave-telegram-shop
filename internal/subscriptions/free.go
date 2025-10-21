package subscriptions

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"remnawave-tg-shop-bot/internal/config"
	"remnawave-tg-shop-bot/internal/database"
	"remnawave-tg-shop-bot/internal/remnawave"
)

type Service struct {
	SubsRepo    *database.SubscriptionRepository
	Customers   *database.CustomerRepository
	RW          *remnawave.Client
	Bot         *bot.Bot
	Translate   interface{ GetText(lang, key string) string }
}

func (s *Service) ActivateFree(ctx context.Context, customerTelegramID int64) (string, error) {
	if config.TrialDays() == 0 {
		return "", nil
	}
	customer, err := s.Customers.FindByTelegramId(ctx, customerTelegramID)
	if err != nil { return "", err }
	if customer == nil { return "", fmt.Errorf("customer %d not found", customerTelegramID) }

	active, err := s.SubsRepo.GetActiveSubscriptions(ctx, customer.ID)
	if err != nil { return "", err }
	seq := len(active)+1

	user, err := s.RW.CreateUserForSubscription(ctx, customer.ID, customer.TelegramID, config.TrialTrafficLimit(), config.TrialDays(), seq)
	if err != nil { return "", err }

	sub := &database.Subscription{
		CustomerID:       customer.ID,
		SubscriptionLink: user.SubscriptionUrl,
		ExpireAt:         user.ExpireAt,
		IsActive:         true,
		Name:             fmt.Sprintf("%s #%d", s.Translate.GetText(customer.Language, "subscription_name"), seq),
		Description:      s.Translate.GetText(customer.Language, "trial_subscription_description"),
	}
	if _, err := s.SubsRepo.CreateSubscription(ctx, sub); err != nil { return "", err }
	return user.SubscriptionUrl, nil
}
