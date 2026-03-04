package seed

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

func DevData(
	ctx context.Context,
	weddingID string,
	invRepo repository.InvitationRepository,
	guestRepo repository.GuestRepository,
	giftRepo repository.GiftRepository,
) error {
	existing, _ := invRepo.CountByWedding(ctx, weddingID)
	if existing > 0 {
		slog.Info("seed de desenvolvimento já aplicado, pulando")
		return nil
	}

	slog.Info("inserindo seed de desenvolvimento...")

	if err := seedInvitationsAndGuests(ctx, weddingID, invRepo, guestRepo); err != nil {
		return fmt.Errorf("seed invitations: %w", err)
	}

	if err := seedGifts(ctx, weddingID, giftRepo); err != nil {
		return fmt.Errorf("seed gifts: %w", err)
	}

	slog.Info("seed de desenvolvimento concluído")
	return nil
}

func seedInvitationsAndGuests(
	ctx context.Context,
	weddingID string,
	invRepo repository.InvitationRepository,
	guestRepo repository.GuestRepository,
) error {
	now := time.Now()

	invitations := []struct {
		code      string
		label     string
		maxGuests int
		notes     string
		guests    []struct {
			name   string
			phone  string
			email  string
			status entity.GuestStatus
		}
	}{
		{
			code: "FAM-SILVA", label: "Família Silva", maxGuests: 4, notes: "Mesa principal",
			guests: []struct {
				name   string
				phone  string
				email  string
				status entity.GuestStatus
			}{
				{name: "João Silva", phone: "11999990001", email: "joao@email.com", status: entity.GuestStatusConfirmed},
				{name: "Maria Silva", phone: "11999990002", email: "maria@email.com", status: entity.GuestStatusConfirmed},
				{name: "Pedro Silva", phone: "", email: "", status: entity.GuestStatusPending},
				{name: "Ana Silva", phone: "", email: "", status: entity.GuestStatusDeclined},
			},
		},
		{
			code: "FAM-SOUZA", label: "Família Souza", maxGuests: 3,
			guests: []struct {
				name   string
				phone  string
				email  string
				status entity.GuestStatus
			}{
				{name: "Carlos Souza", phone: "11999990003", email: "carlos@email.com", status: entity.GuestStatusConfirmed},
				{name: "Fernanda Souza", phone: "11999990004", email: "", status: entity.GuestStatusConfirmed},
				{name: "Lucas Souza", phone: "", email: "", status: entity.GuestStatusPending},
			},
		},
		{
			code: "AMIGOS-01", label: "Turma da Faculdade", maxGuests: 5,
			guests: []struct {
				name   string
				phone  string
				email  string
				status entity.GuestStatus
			}{
				{name: "Bruna Oliveira", phone: "11999990005", email: "bruna@email.com", status: entity.GuestStatusConfirmed},
				{name: "Diego Lima", phone: "11999990006", email: "diego@email.com", status: entity.GuestStatusConfirmed},
				{name: "Gabriela Costa", phone: "", email: "gabi@email.com", status: entity.GuestStatusPending},
				{name: "Henrique Pereira", phone: "", email: "", status: entity.GuestStatusPending},
			},
		},
		{
			code: "TRAB-01", label: "Colegas de Trabalho", maxGuests: 3,
			guests: []struct {
				name   string
				phone  string
				email  string
				status entity.GuestStatus
			}{
				{name: "Patrícia Almeida", phone: "11999990007", email: "patricia@empresa.com", status: entity.GuestStatusPending},
				{name: "Roberto Santos", phone: "11999990008", email: "roberto@empresa.com", status: entity.GuestStatusDeclined},
			},
		},
		{
			code: "PAD-01", label: "Padrinhos", maxGuests: 2, notes: "Reservar lugar especial",
			guests: []struct {
				name   string
				phone  string
				email  string
				status entity.GuestStatus
			}{
				{name: "Marcelo Ferreira", phone: "11999990009", email: "marcelo@email.com", status: entity.GuestStatusConfirmed},
				{name: "Juliana Ferreira", phone: "11999990010", email: "juliana@email.com", status: entity.GuestStatusConfirmed},
			},
		},
	}

	for _, inv := range invitations {
		invID := uuid.New().String()
		invitation := &entity.Invitation{
			ID:        invID,
			WeddingID: weddingID,
			Code:      inv.code,
			Label:     inv.label,
			MaxGuests: inv.maxGuests,
			Notes:     inv.notes,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := invRepo.Create(ctx, invitation); err != nil {
			return fmt.Errorf("create invitation %s: %w", inv.code, err)
		}

		for _, g := range inv.guests {
			var confirmedAt *time.Time
			if g.status == entity.GuestStatusConfirmed {
				t := now.Add(-24 * time.Hour)
				confirmedAt = &t
			}

			guest := &entity.Guest{
				ID:           uuid.New().String(),
				InvitationID: invID,
				WeddingID:    weddingID,
				Name:         g.name,
				Phone:        g.phone,
				Email:        g.email,
				Status:       g.status,
				ConfirmedAt:  confirmedAt,
				CreatedAt:    now,
				UpdatedAt:    now,
			}

			if err := guestRepo.Create(ctx, guest); err != nil {
				return fmt.Errorf("create guest %s: %w", g.name, err)
			}
		}
	}

	return nil
}

func seedGifts(ctx context.Context, weddingID string, giftRepo repository.GiftRepository) error {
	now := time.Now()

	gifts := []struct {
		name        string
		description string
		price       float64
		category    string
		imageURL    string
	}{
		{name: "Jogo de Panelas", description: "Jogo com 5 panelas antiaderentes de alta qualidade", price: 450.00, category: "Cozinha", imageURL: "https://placehold.co/400x300?text=Panelas"},
		{name: "Cafeteira Nespresso", description: "Cafeteira de cápsulas para cafés especiais", price: 350.00, category: "Cozinha", imageURL: "https://placehold.co/400x300?text=Cafeteira"},
		{name: "Jogo de Toalhas", description: "Kit com 8 toalhas de banho e rosto em algodão egípcio", price: 280.00, category: "Banheiro", imageURL: "https://placehold.co/400x300?text=Toalhas"},
		{name: "Aspirador Robô", description: "Aspirador robô inteligente com mapeamento", price: 1200.00, category: "Casa", imageURL: "https://placehold.co/400x300?text=Aspirador"},
		{name: "Jogo de Cama King", description: "Kit completo lençol, fronha e edredom king size 400 fios", price: 520.00, category: "Quarto", imageURL: "https://placehold.co/400x300?text=Cama"},
		{name: "Mixer", description: "Mixer de mão com acessórios para triturar e bater", price: 190.00, category: "Cozinha", imageURL: "https://placehold.co/400x300?text=Mixer"},
		{name: "Jogo de Taças", description: "12 taças de cristal para vinho tinto e branco", price: 320.00, category: "Cozinha", imageURL: "https://placehold.co/400x300?text=Tacas"},
		{name: "Smart TV 55\"", description: "Smart TV 4K 55 polegadas com sistema operacional atualizado", price: 2800.00, category: "Sala", imageURL: "https://placehold.co/400x300?text=TV"},
		{name: "Air Fryer", description: "Fritadeira elétrica digital 5.5L", price: 380.00, category: "Cozinha", imageURL: "https://placehold.co/400x300?text=AirFryer"},
		{name: "Jogo de Talheres", description: "Faqueiro completo 72 peças em aço inox", price: 350.00, category: "Cozinha", imageURL: "https://placehold.co/400x300?text=Talheres"},
		{name: "Travesseiros", description: "Par de travesseiros de pluma de ganso", price: 240.00, category: "Quarto", imageURL: "https://placehold.co/400x300?text=Travesseiros"},
		{name: "Lua de Mel", description: "Contribuição para a viagem dos noivos", price: 500.00, category: "Experiência", imageURL: "https://placehold.co/400x300?text=LuaDeMel"},
	}

	for _, g := range gifts {
		gift := &entity.Gift{
			ID:          uuid.New().String(),
			WeddingID:   weddingID,
			Name:        g.name,
			Description: g.description,
			Price:       g.price,
			ImageURL:    g.imageURL,
			Category:    g.category,
			Status:      entity.GiftStatusAvailable,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := giftRepo.Create(ctx, gift); err != nil {
			return fmt.Errorf("create gift %s: %w", g.name, err)
		}
	}

	return nil
}
