package payment

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"remnawave-tg-shop-bot/internal/cache"
	"remnawave-tg-shop-bot/internal/config"
	"remnawave-tg-shop-bot/internal/cryptopay"
	"remnawave-tg-shop-bot/internal/database"
	"remnawave-tg-shop-bot/internal/remnawave"
	"remnawave-tg-shop-bot/internal/translation"
	"remnawave-tg-shop-bot/internal/yookasa"
	"remnawave-tg-shop-bot/utils"
	"time"
)

// ... (structs and constructor unchanged)

func (s PaymentService) ProcessPurchaseById(ctx context.Context, purchaseId int64) error {
	purchase, err := s.purchaseRepository.FindById(ctx, purchaseId)
	if err != nil { return err }
	if purchase == nil { return fmt.Errorf("purchase with crypto invoice id %d not found", utils.MaskHalfInt64(purchaseId)) }

	customer, err := s.customerRepository.FindById(ctx, purchase.CustomerID)
	if err != nil { return err }
	if customer == nil { return fmt.Errorf("customer %s not found", utils.MaskHalfInt64(purchase.CustomerID)) }

	if messageId, b := s.cache.Get(purchase.ID); b {
		_, _ = s.telegramBot.DeleteMessage(ctx, &bot.DeleteMessageParams{ ChatID: customer.TelegramID, MessageID: messageId })
	}

	// Find active subscriptions to determine sequence and create a fresh RW user for this subscription
	activeSubscriptions, err := s.subscriptionRepository.GetActiveSubscriptions(ctx, customer.ID)
	if err != nil { return fmt.Errorf("failed to get active subscriptions: %w", err) }
	seq := len(activeSubscriptions) + 1

	user, err := s.remnawaveClient.CreateUserForSubscription(ctx, customer.ID, customer.TelegramID, config.TrafficLimit(), purchase.Month*config.DaysInMonth(), seq)
	if err != nil { return err }

	// Mark purchase as paid
	if err := s.purchaseRepository.MarkAsPaid(ctx, purchase.ID); err != nil { return err }

	// Name and create subscription record
	subscriptionName := fmt.Sprintf("%s #%d", s.translation.GetText(customer.Language, "subscription_name"), seq)
	subscriptionDescription := fmt.Sprintf("%d %s", purchase.Month, s.translation.GetText(customer.Language, "months_word"))
	newSubscription := &database.Subscription{ CustomerID: customer.ID, SubscriptionLink: user.SubscriptionUrl, ExpireAt: user.ExpireAt, IsActive: true, Name: subscriptionName, Description: subscriptionDescription }
	if _, err := s.subscriptionRepository.CreateSubscription(ctx, newSubscription); err != nil { return fmt.Errorf("failed to create subscription: %w", err) }

	// Backward compatibility fields on customer
	_ = s.customerRepository.UpdateFields(ctx, customer.ID, map[string]interface{}{ "subscription_link": user.SubscriptionUrl, "expire_at": user.ExpireAt })

	// Notify
	_, err = s.telegramBot.SendMessage(ctx, &bot.SendMessageParams{ ChatID: customer.TelegramID, Text: fmt.Sprintf(s.translation.GetText(customer.Language, "subscription_activated_multiple"), subscriptionName), ReplyMarkup: models.InlineKeyboardMarkup{ InlineKeyboard: s.createSubscriptionsKeyboard(customer) } })
	if err != nil { return err }

	// Referral bonus unchanged
	ctxReferee := context.Background()
	referee, err := s.referralRepository.FindByReferee(ctxReferee, customer.TelegramID)
	if referee == nil || err != nil { if err == nil { return nil } ; return err }
	if referee.BonusGranted { return nil }

	refereeCustomer, err := s.customerRepository.FindByTelegramId(ctxReferee, referee.ReferrerID)
	if err != nil { return err }
	refSeqSubs, err := s.subscriptionRepository.GetActiveSubscriptions(ctxReferee, refereeCustomer.ID)
	if err != nil { return err }
	refSeq := len(refSeqSubs) + 1
	refereeUser, err := s.remnawaveClient.CreateUserForSubscription(ctxReferee, refereeCustomer.ID, refereeCustomer.TelegramID, config.TrafficLimit(), config.GetReferralDays(), refSeq)
	if err != nil { return err }

	bonusSubscription := &database.Subscription{ CustomerID: refereeCustomer.ID, SubscriptionLink: refereeUser.SubscriptionUrl, ExpireAt: refereeUser.ExpireAt, IsActive: true, Name: s.translation.GetText(refereeCustomer.Language, "referral_bonus_subscription"), Description: s.translation.GetText(refereeCustomer.Language, "referral_bonus_description") }
	if _, err := s.subscriptionRepository.CreateSubscription(ctxReferee, bonusSubscription); err != nil { return err }
	_ = s.customerRepository.UpdateFields(ctxReferee, refereeCustomer.ID, map[string]interface{}{ "subscription_link": refereeUser.SubscriptionUrl, "expire_at": refereeUser.ExpireAt })
	if err := s.referralRepository.MarkBonusGranted(ctxReferee, referee.ID); err != nil { return err }
	_, _ = s.telegramBot.SendMessage(ctxReferee, &bot.SendMessageParams{ ChatID: refereeCustomer.TelegramID, ParseMode: models.ParseModeHTML, Text: s.translation.GetText(refereeCustomer.Language, "referral_bonus_granted"), ReplyMarkup: models.InlineKeyboardMarkup{ InlineKeyboard: s.createSubscriptionsKeyboard(refereeCustomer) } })
	return nil
}

func (s PaymentService) ActivateTrial(ctx context.Context, telegramId int64) (string, error) {
	if config.TrialDays() == 0 { return "", nil }
	customer, err := s.customerRepository.FindByTelegramId(ctx, telegramId)
	if err != nil { slog.Error("Error finding customer", err); return "", err }
	if customer == nil { return "", fmt.Errorf("customer %d not found", telegramId) }

	// Determine seq and create a fresh RW user for this trial subscription
	activeSubscriptions, err := s.subscriptionRepository.GetActiveSubscriptions(ctx, customer.ID)
	if err != nil { return "", fmt.Errorf("failed to get active subscriptions: %w", err) }
	seq := len(activeSubscriptions) + 1
	user, err := s.remnawaveClient.CreateUserForSubscription(ctx, customer.ID, telegramId, config.TrialTrafficLimit(), config.TrialDays(), seq)
	if err != nil { slog.Error("Error creating user", err); return "", err }

	trialSubscription := &database.Subscription{ CustomerID: customer.ID, SubscriptionLink: user.SubscriptionUrl, ExpireAt: user.ExpireAt, IsActive: true, Name: fmt.Sprintf("%s #%d", s.translation.GetText(customer.Language, "subscription_name"), seq), Description: s.translation.GetText(customer.Language, "trial_subscription_description") }
	if _, err := s.subscriptionRepository.CreateSubscription(ctx, trialSubscription); err != nil { return "", fmt.Errorf("failed to create trial subscription: %w", err) }
	_ = s.customerRepository.UpdateFields(ctx, customer.ID, map[string]interface{}{ "subscription_link": user.SubscriptionUrl, "expire_at": user.ExpireAt })
	return user.SubscriptionUrl, nil
}
