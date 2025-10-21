package payment

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"remnawave-tg-shop-bot/internal/config"
	"remnawave-tg-shop-bot/internal/database"
	"remnawave-tg-shop-bot/utils"
)

type PaymentService struct {
	purchaseRepository     *database.PurchaseRepository
	remnawaveClient        interface{ CreateUserForSubscription(ctx context.Context, customerId int64, telegramId int64, trafficLimit int, days int, seq int) (user interface{ GetSubscriptionUrl() string; GetExpireAt() string }, err error) }
	customerRepository     *database.CustomerRepository
	subscriptionRepository *database.SubscriptionRepository
	telegramBot            *bot.Bot
	translation            interface{ GetText(lang, key string) string }
	cryptoPayClient        interface{}
	yookasaClient          interface{}
	referralRepository     *database.ReferralRepository
	cache                  interface{ Get(id int64) (int, bool) }
}

