package handler

import (
	"remnawave-tg-shop-bot/internal/cache"
	"remnawave-tg-shop-bot/internal/cryptopay"
	"remnawave-tg-shop-bot/internal/database"
	"remnawave-tg-shop-bot/internal/sync"
	"remnawave-tg-shop-bot/internal/translation"
	"remnawave-tg-shop-bot/internal/yookasa"
)

type Handler struct {
	customerRepository     *database.CustomerRepository
	purchaseRepository     *database.PurchaseRepository
	subscriptionRepository *database.SubscriptionRepository
	cryptoPayClient        *cryptopay.Client
	yookasaClient          *yookasa.Client
	translation            *translation.Manager
	syncService            *sync.SyncService
	referralRepository     *database.ReferralRepository
	cache                  *cache.Cache
}

func NewHandler(
	syncService *sync.SyncService,
	_ interface{},
	translation *translation.Manager,
	customerRepository *database.CustomerRepository,
	purchaseRepository *database.PurchaseRepository,
	subscriptionRepository *database.SubscriptionRepository,
	cryptoPayClient *cryptopay.Client,
	yookasaClient *yookasa.Client,
	referralRepository *database.ReferralRepository,
	cache *cache.Cache) *Handler {
	return &Handler{
		syncService:            syncService,
		customerRepository:     customerRepository,
		purchaseRepository:     purchaseRepository,
		subscriptionRepository: subscriptionRepository,
		cryptoPayClient:        cryptoPayClient,
		yookasaClient:          yookasaClient,
		translation:            translation,
		referralRepository:     referralRepository,
		cache:                  cache,
	}
}
