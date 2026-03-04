package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/by-r2/weddo-api/internal/infra/config"
	"github.com/by-r2/weddo-api/internal/infra/database"
	"github.com/by-r2/weddo-api/internal/infra/gateway"
	"github.com/by-r2/weddo-api/internal/infra/security"
	"github.com/by-r2/weddo-api/internal/infra/seed"
	infraSheets "github.com/by-r2/weddo-api/internal/infra/sheets"
	"github.com/by-r2/weddo-api/internal/infra/web"
	giftuc "github.com/by-r2/weddo-api/internal/usecase/gift"
	"github.com/by-r2/weddo-api/internal/usecase/guest"
	"github.com/by-r2/weddo-api/internal/usecase/invitation"
	paymentuc "github.com/by-r2/weddo-api/internal/usecase/payment"
	"github.com/by-r2/weddo-api/internal/usecase/rsvp"
	sheetsuc "github.com/by-r2/weddo-api/internal/usecase/sheets"
	"github.com/by-r2/weddo-api/internal/usecase/wedding"
)

func main() {
	healthCheck := flag.Bool("health", false, "executa health check e encerra")
	seedDev := flag.Bool("seed-dev", false, "insere dados fictícios para desenvolvimento")
	flag.Parse()

	if *healthCheck {
		resp, err := http.Get("http://localhost:8080/api/v1/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	setupLogger(cfg.LogLevel, cfg.LogFormat)

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.RunMigrations(db, "migrations"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	weddingRepo := database.NewWeddingRepository(db)
	invitationRepo := database.NewInvitationRepository(db)
	guestRepo := database.NewGuestRepository(db)
	giftRepo := database.NewGiftRepository(db)
	paymentRepo := database.NewPaymentRepository(db)
	googleIntegrationRepo := database.NewGoogleIntegrationRepository(db)

	weddingUC := wedding.NewUseCase(weddingRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
	rsvpUC := rsvp.NewUseCase(guestRepo, invitationRepo)
	invitationUC := invitation.NewUseCase(invitationRepo, guestRepo)
	guestUC := guest.NewUseCase(guestRepo, invitationRepo)
	giftUC := giftuc.NewUseCase(giftRepo, paymentRepo)
	var sheetsUC *sheetsuc.UseCase

	var paymentUC *paymentuc.UseCase
	switch strings.ToLower(cfg.PaymentProvider) {
	case "infinitepay":
		if cfg.IPHandle == "" {
			slog.Error("PAYMENT_PROVIDER=infinitepay requer IP_HANDLE")
			os.Exit(1)
		}
		ipGateway := gateway.NewInfinitePayGateway(cfg.IPHandle, cfg.IPRedirectURL, cfg.IPWebhookURL)
		paymentUC = paymentuc.NewUseCase(paymentRepo, giftRepo, ipGateway)
		slog.Info("payment gateway initialized", "provider", "infinitepay")

	case "mercadopago":
		if cfg.MPAccessToken == "" {
			slog.Error("PAYMENT_PROVIDER=mercadopago requer MP_ACCESS_TOKEN")
			os.Exit(1)
		}
		mpGateway, err := gateway.NewMercadoPagoGateway(cfg.MPAccessToken, cfg.MPNotificationURL, cfg.MPPixExpirationMin)
		if err != nil {
			slog.Error("failed to init mercado pago gateway", "error", err)
			os.Exit(1)
		}
		paymentUC = paymentuc.NewUseCase(paymentRepo, giftRepo, mpGateway)
		slog.Info("payment gateway initialized", "provider", "mercadopago")

	case "":
		slog.Warn("PAYMENT_PROVIDER não definido — endpoints de pagamento retornarão 503")

	default:
		slog.Error("PAYMENT_PROVIDER inválido — use 'infinitepay' ou 'mercadopago'", "value", cfg.PaymentProvider)
		os.Exit(1)
	}

	if cfg.SeedAdminEmail != "" && cfg.SeedAdminPassword != "" {
		if err := seedWedding(weddingUC, cfg); err != nil {
			slog.Error("failed to seed wedding", "error", err)
			os.Exit(1)
		}
	}

	if cfg.GoogleOAuthClientID != "" && cfg.GoogleOAuthClientSecret != "" && cfg.GoogleOAuthRedirectURL != "" && cfg.GoogleOAuthTokenCipherKey != "" {
		oauthProvider := infraSheets.NewOAuthProvider(cfg.GoogleOAuthClientID, cfg.GoogleOAuthClientSecret, cfg.GoogleOAuthRedirectURL)
		tokenCipher, err := security.NewCipher(cfg.GoogleOAuthTokenCipherKey)
		if err != nil {
			slog.Error("failed to init google oauth token cipher", "error", err)
			os.Exit(1)
		}
		stateSecret := cfg.GoogleOAuthStateSecret
		if stateSecret == "" {
			stateSecret = cfg.JWTSecret
		}
		sheetsUC = sheetsuc.NewUseCase(
			invitationRepo,
			guestRepo,
			giftRepo,
			paymentRepo,
			weddingRepo,
			googleIntegrationRepo,
			oauthProvider,
			tokenCipher,
			stateSecret,
		)
		slog.Info("google sheets oauth sync enabled")
	} else {
		slog.Warn("google sheets oauth sync disabled — configure GOOGLE_OAUTH_CLIENT_ID, GOOGLE_OAUTH_CLIENT_SECRET, GOOGLE_OAUTH_REDIRECT_URL and GOOGLE_OAUTH_TOKEN_CIPHER_KEY")
	}

	if *seedDev {
		w, err := weddingRepo.FindBySlug(context.Background(), cfg.SeedWeddingSlug)
		if err != nil {
			slog.Error("wedding não encontrado para seed de dev", "slug", cfg.SeedWeddingSlug, "error", err)
			os.Exit(1)
		}
		if err := seed.DevData(context.Background(), w.ID, invitationRepo, guestRepo, giftRepo); err != nil {
			slog.Error("failed to seed dev data", "error", err)
			os.Exit(1)
		}
	}

	router := web.NewRouter(web.RouterDeps{
		WeddingUC:    weddingUC,
		RSVPUC:       rsvpUC,
		InvitationUC: invitationUC,
		GuestUC:      guestUC,
		GiftUC:       giftUC,
		PaymentUC:    paymentUC,
		SheetsUC:     sheetsUC,
		WeddingRepo:  weddingRepo,
		JWTSecret:    cfg.JWTSecret,
		CORSOrigins:  cfg.CORSAllowedOrigins,
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func setupLogger(level, format string) {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func seedWedding(uc *wedding.UseCase, cfg *config.Config) error {
	p1, p2 := parseSeedPartners(cfg.SeedWeddingTitle)

	err := uc.Seed(
		context.Background(),
		cfg.SeedWeddingSlug,
		cfg.SeedWeddingTitle,
		cfg.SeedWeddingDate,
		p1, p2,
		cfg.SeedAdminEmail,
		cfg.SeedAdminPassword,
	)
	if err != nil {
		return err
	}

	slog.Info("wedding seed checked", "slug", cfg.SeedWeddingSlug)
	return nil
}

// parseSeedPartners extrai nomes dos parceiros a partir do título.
// Formato esperado: "Casamento Manoela & Rafael" → ("Manoela", "Rafael")
// Fallback se não encontrar: ("Partner 1", "Partner 2")
func parseSeedPartners(title string) (string, string) {
	title = strings.TrimPrefix(title, "Casamento ")
	parts := strings.SplitN(title, " & ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "Partner 1", "Partner 2"
}
