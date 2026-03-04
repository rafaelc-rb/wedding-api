package invitation

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type UseCase struct {
	invRepo  repository.InvitationRepository
	guestRepo repository.GuestRepository
}

func NewUseCase(ir repository.InvitationRepository, gr repository.GuestRepository) *UseCase {
	return &UseCase{invRepo: ir, guestRepo: gr}
}

type CreateInput struct {
	WeddingID string
	Code      string
	Label     string
	MaxGuests int
	Notes     string
	Guests    []string // nomes dos convidados
}

func (uc *UseCase) Create(ctx context.Context, input CreateInput) (*entity.Invitation, error) {
	now := time.Now()
	inv := &entity.Invitation{
		ID:        uuid.New().String(),
		WeddingID: input.WeddingID,
		Code:      input.Code,
		Label:     input.Label,
		MaxGuests: input.MaxGuests,
		Notes:     input.Notes,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.invRepo.Create(ctx, inv); err != nil {
		return nil, fmt.Errorf("invitation.Create: %w", err)
	}

	for _, name := range input.Guests {
		guest := &entity.Guest{
			ID:           uuid.New().String(),
			InvitationID: inv.ID,
			WeddingID:    input.WeddingID,
			Name:         name,
			Status:       entity.GuestStatusPending,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := uc.guestRepo.Create(ctx, guest); err != nil {
			return nil, fmt.Errorf("invitation.Create: create guest %q: %w", name, err)
		}
		inv.Guests = append(inv.Guests, *guest)
	}

	return inv, nil
}

func (uc *UseCase) FindByID(ctx context.Context, weddingID, id string) (*entity.Invitation, error) {
	inv, err := uc.invRepo.FindByID(ctx, weddingID, id)
	if err != nil {
		return nil, err
	}

	guests, err := uc.guestRepo.ListByInvitation(ctx, weddingID, id)
	if err != nil {
		return nil, fmt.Errorf("invitation.FindByID: list guests: %w", err)
	}
	inv.Guests = guests

	return inv, nil
}

func (uc *UseCase) List(ctx context.Context, weddingID string, page, perPage int, search string) ([]entity.Invitation, int, error) {
	return uc.invRepo.List(ctx, weddingID, page, perPage, search)
}

type UpdateInput struct {
	WeddingID string
	ID        string
	Code      string
	Label     string
	MaxGuests int
	Notes     string
}

func (uc *UseCase) Update(ctx context.Context, input UpdateInput) (*entity.Invitation, error) {
	inv, err := uc.invRepo.FindByID(ctx, input.WeddingID, input.ID)
	if err != nil {
		return nil, err
	}

	inv.Code = input.Code
	inv.Label = input.Label
	inv.MaxGuests = input.MaxGuests
	inv.Notes = input.Notes
	inv.UpdatedAt = time.Now()

	if err := uc.invRepo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("invitation.Update: %w", err)
	}
	return inv, nil
}

func (uc *UseCase) Delete(ctx context.Context, weddingID, id string) error {
	return uc.invRepo.Delete(ctx, weddingID, id)
}

// AddGuest adiciona um convidado a um convite existente.
func (uc *UseCase) AddGuest(ctx context.Context, weddingID, invitationID, name, phone, email string) (*entity.Guest, error) {
	if _, err := uc.invRepo.FindByID(ctx, weddingID, invitationID); err != nil {
		return nil, err
	}

	now := time.Now()
	guest := &entity.Guest{
		ID:           uuid.New().String(),
		InvitationID: invitationID,
		WeddingID:    weddingID,
		Name:         name,
		Phone:        phone,
		Email:        email,
		Status:       entity.GuestStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.guestRepo.Create(ctx, guest); err != nil {
		return nil, fmt.Errorf("invitation.AddGuest: %w", err)
	}
	return guest, nil
}
