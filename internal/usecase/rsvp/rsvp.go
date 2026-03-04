package rsvp

import (
	"context"
	"fmt"
	"time"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type UseCase struct {
	guestRepo      repository.GuestRepository
	invitationRepo repository.InvitationRepository
}

func NewUseCase(gr repository.GuestRepository, ir repository.InvitationRepository) *UseCase {
	return &UseCase{guestRepo: gr, invitationRepo: ir}
}

// Confirm busca o convidado por nome e registra a confirmação.
// Retorna o guest atualizado e o invitation ao qual pertence.
func (uc *UseCase) Confirm(ctx context.Context, weddingID, name string) (*entity.Guest, *entity.Invitation, bool, error) {
	guest, err := uc.guestRepo.FindByName(ctx, weddingID, name)
	if err != nil {
		return nil, nil, false, err
	}

	alreadyConfirmed := guest.Status == entity.GuestStatusConfirmed

	if !alreadyConfirmed {
		now := time.Now()
		guest.Status = entity.GuestStatusConfirmed
		guest.ConfirmedAt = &now
		guest.UpdatedAt = now

		if err := uc.guestRepo.Update(ctx, guest); err != nil {
			return nil, nil, false, fmt.Errorf("rsvp.Confirm: %w", err)
		}
	}

	inv, err := uc.invitationRepo.FindByID(ctx, weddingID, guest.InvitationID)
	if err != nil {
		return nil, nil, false, fmt.Errorf("rsvp.Confirm: find invitation: %w", err)
	}

	return guest, inv, alreadyConfirmed, nil
}

// LookupInvitation busca o convite de um convidado pelo nome e lista todos os guests do convite.
func (uc *UseCase) LookupInvitation(ctx context.Context, weddingID, name string) (*entity.Invitation, []entity.Guest, error) {
	guest, err := uc.guestRepo.FindByName(ctx, weddingID, name)
	if err != nil {
		return nil, nil, err
	}

	inv, err := uc.invitationRepo.FindByID(ctx, weddingID, guest.InvitationID)
	if err != nil {
		return nil, nil, fmt.Errorf("rsvp.LookupInvitation: find invitation: %w", err)
	}

	guests, err := uc.guestRepo.ListByInvitation(ctx, weddingID, inv.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("rsvp.LookupInvitation: list guests: %w", err)
	}

	return inv, guests, nil
}
