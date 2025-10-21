package main

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/robfig/cron/v3"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"remnawave-tg-shop-bot/internal/config"
	"remnawave-tg-shop-bot/internal/database"
	"remnawave-tg-shop-bot/internal/handler"
	"remnawave-tg-shop-bot/internal/notification"
	"remnawave-tg-shop-bot/internal/remnawave"
	"remnawave-tg-shop-bot/internal/sync"
	"remnawave-tg-shop-bot/internal/translation"
	"remnawave-tg-shop-bot/internal/tribute"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config.InitConfig()

	tm := translation.GetInstance()
	if err := tm.InitTranslations("./translations"); err != nil { panic(err) }

	pool, err := initDatabase(ctx, config.DadaBaseUrl())
	if err != nil { panic(err) }
	if err := database.RunMigrations(ctx, &database.MigrationConfig{Direction: "up", MigrationsPath: "./db/migrations", Steps: 0}, pool); err != nil { panic(err) }

	customerRepository := database.NewCustomerRepository(pool)
	subscriptionRepository := database.NewSubscriptionRepository(pool)
	referralRepository := database.NewReferralRepository(pool)

	rw := remnawave.NewClient(config.RemnawaveUrl(), config.RemnawaveToken(), config.RemnawaveMode())
	b, err := bot.New(config.TelegramToken(), bot.WithWorkers(3))
	if err != nil { panic(err) }

	syncService := sync.NewSyncService(rw, customerRepository)
	h := handler.NewHandler(syncService, nil, tm, customerRepository, nil, subscriptionRepository, nil, nil, referralRepository, nil)

	me, err := b.GetMe(ctx); if err != nil { panic(err) }
	_, _ = b.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{ MenuButton: &models.MenuButtonCommands{ Type: models.MenuButtonTypeCommands } })
	_, _ = b.SetMyCommands(ctx, &bot.SetMyCommandsParams{ Commands: []models.BotCommand{{Command:"start", Description:"Начать работу с ботом"}}, LanguageCode: "ru"})
	_, _ = b.SetMyCommands(ctx, &bot.SetMyCommandsParams{ Commands: []models.BotCommand{{Command:"start", Description:"Start using the bot"}}, LanguageCode: "en"})
	config.SetBotURL(fmt.Sprintf("https://t.me/%s", me.Username))

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypePrefix, h.StartCommandHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/connect", bot.MatchTypeExact, h.ConnectCommandHandler, h.CreateCustomerIfNotExistMiddleware)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/sync", bot.MatchTypeExact, h.SyncUsersCommandHandler, isAdminMiddleware)

	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackReferral, bot.MatchTypeExact, h.ReferralCallbackHandler, h.CreateCustomerIfNotExistMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackTrial, bot.MatchTypeExact, h.TrialCallbackHandler, h.CreateCustomerIfNotExistMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackStart, bot.MatchTypeExact, h.StartCallbackHandler, h.CreateCustomerIfNotExistMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackConnect, bot.MatchTypeExact, h.ConnectCallbackHandler, h.CreateCustomerIfNotExistMiddleware)

	// Multiple subscriptions
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackMySubscriptions, bot.MatchTypeExact, h.MySubscriptionsCallbackHandler, h.CreateCustomerIfNotExistMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackDeactivateSubscription, bot.MatchTypePrefix, h.DeactivateSubscriptionCallbackHandler, h.CreateCustomerIfNotExistMiddleware)

	// Broadcast (admins only)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackBroadcastMenu, bot.MatchTypeExact, h.BroadcastMenuHandler, isAdminMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackBroadcastToAll, bot.MatchTypeExact, h.BroadcastTypeHandler, isAdminMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackBroadcastToAdmins, bot.MatchTypeExact, h.BroadcastTypeHandler, isAdminMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackBroadcastConfirm, bot.MatchTypePrefix, h.BroadcastConfirmHandler, isAdminMiddleware)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, handler.CallbackBroadcastCancel, bot.MatchTypeExact, h.BroadcastCancelHandler, isAdminMiddleware)
	b.RegisterHandlerMatchFunc(func(update *models.Update) bool { return update.Message != nil && update.Message.From.ID == config.GetAdminTelegramId() }, h.BroadcastMessageHandler)

	mux := http.NewServeMux()
	mux.Handle("/healthcheck", fullHealthHandler(pool, rw))
	if config.GetTributeWebHookUrl() != "" {
		tributeHandler := tribute.NewClient(nil, customerRepository)
		mux.Handle(config.GetTributeWebHookUrl(), tributeHandler.WebHookHandler())
	}

	srv := &http.Server{ Addr: fmt.Sprintf(":%d", config.GetHealthCheckPort()), Handler: mux }
	go func(){ log.Printf("Server listening on %s", srv.Addr); if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed { log.Fatalf("Server error: %v", err) } }()

	slog.Info("Bot is starting...")
	b.Start(ctx)

	log.Println("Shutting down health server…")
	shutdownCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second); defer shutCancel()
	_ = srv.Shutdown(shutdownCtx)
}

func fullHealthHandler(pool *pgxpool.Pool, rw *remnawave.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{ "status":"ok", "db":"ok", "rw":"ok", "time": time.Now().Format(time.RFC3339) }
		dbCtx, dbCancel := context.WithTimeout(r.Context(), 5*time.Second); defer dbCancel()
		if err := pool.Ping(dbCtx); err != nil { w.WriteHeader(http.StatusServiceUnavailable); status["status"] = "fail"; status["db"] = "error: "+err.Error() }
		rwCtx, rwCancel := context.WithTimeout(r.Context(), 5*time.Second); defer rwCancel()
		if err := rw.Ping(rwCtx); err != nil { w.WriteHeader(http.StatusServiceUnavailable); status["status"] = "fail"; status["rw"] = "error: "+err.Error() }
		if status["status"] == "ok" { w.WriteHeader(http.StatusOK) }
		w.Header().Set("Content-Type","application/json")
		fmt.Fprintf(w, `{"status":"%s","db":"%s","remnawave":"%s","time":"%s"}`, status["status"], status["db"], status["rw"], status["time"]) 
	})
}

func initDatabase(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString); if err != nil { return nil, err }
	cfg.MaxConns = 20; cfg.MinConns = 5
	return pgxpool.ConnectConfig(ctx, cfg)
}

func isAdminMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		var userID int64
		if update.Message != nil { userID = update.Message.From.ID } else if update.CallbackQuery != nil { userID = update.CallbackQuery.From.ID } else { return }
		if userID == config.GetAdminTelegramId() { next(ctx, b, update) }
	}
}
