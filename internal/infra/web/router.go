package web

import (
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/repository"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/infra/web/handler"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/infra/web/middleware"
	giftuc "github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/gift"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/guest"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/invitation"
	paymentuc "github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/payment"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/rsvp"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/usecase/wedding"
)

type RouterDeps struct {
	WeddingUC    *wedding.UseCase
	RSVPUC       *rsvp.UseCase
	InvitationUC *invitation.UseCase
	GuestUC      *guest.UseCase
	GiftUC       *giftuc.UseCase
	PaymentUC    *paymentuc.UseCase
	WeddingRepo  repository.WeddingRepository
	JWTSecret    string
	CORSOrigins  string
}

func NewRouter(deps RouterDeps) *chi.Mux {
	r := chi.NewRouter()

	origins := strings.Split(deps.CORSOrigins, ",")
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.RealIP)

	authHandler := handler.NewAuthHandler(deps.WeddingUC)
	rsvpHandler := handler.NewRSVPHandler(deps.RSVPUC)
	invHandler := handler.NewInvitationHandler(deps.InvitationUC)
	guestHandler := handler.NewGuestHandler(deps.GuestUC)
	giftHandler := handler.NewGiftHandler(deps.GiftUC)
	paymentHandler := handler.NewPaymentHandler(deps.PaymentUC)
	dashHandler := handler.NewDashboardHandler(deps.GuestUC, deps.GiftUC)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.Health)

		// Endpoints públicos (tenant via UUID na URL)
		r.Route("/w/{weddingId}", func(r chi.Router) {
			r.Use(middleware.TenantResolver(deps.WeddingRepo))

			r.Post("/rsvp", rsvpHandler.Confirm)
			r.Get("/rsvp/invitation", rsvpHandler.LookupInvitation)

			r.Get("/gifts", giftHandler.ListPublic)
			r.Get("/gifts/{id}", giftHandler.GetPublic)
			r.Post("/gifts/{id}/purchase", paymentHandler.Purchase)
			r.Get("/payments/{id}/status", paymentHandler.GetStatus)
		})

		// Webhook (sem auth — validação via assinatura do provider)
		r.Post("/payments/webhook", paymentHandler.Webhook)

		// Autenticação
		r.Post("/admin/auth", authHandler.Login)

		// Endpoints admin (tenant via JWT)
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))

			r.Get("/dashboard", dashHandler.Get)

			r.Route("/invitations", func(r chi.Router) {
				r.Get("/", invHandler.List)
				r.Post("/", invHandler.Create)
				r.Get("/{id}", invHandler.GetByID)
				r.Put("/{id}", invHandler.Update)
				r.Delete("/{id}", invHandler.Delete)
				r.Post("/{id}/guests", invHandler.AddGuest)
			})

			r.Route("/guests", func(r chi.Router) {
				r.Get("/", guestHandler.List)
				r.Get("/{id}", guestHandler.GetByID)
				r.Put("/{id}", guestHandler.Update)
				r.Delete("/{id}", guestHandler.Delete)
			})

			r.Route("/gifts", func(r chi.Router) {
				r.Get("/", giftHandler.List)
				r.Post("/", giftHandler.Create)
				r.Get("/{id}", giftHandler.GetByID)
				r.Put("/{id}", giftHandler.Update)
				r.Delete("/{id}", giftHandler.Delete)
			})

			r.Route("/payments", func(r chi.Router) {
				r.Get("/", paymentHandler.ListAdmin)
				r.Get("/{id}", paymentHandler.GetAdmin)
			})
		})
	})

	return r
}
