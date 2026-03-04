package guest

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

func (uc *UseCase) FindByID(ctx context.Context, weddingID, id string) (*entity.Guest, error) {
	return uc.guestRepo.FindByID(ctx, weddingID, id)
}

func (uc *UseCase) List(ctx context.Context, weddingID string, page, perPage int, status, search string) ([]entity.Guest, int, error) {
	return uc.guestRepo.List(ctx, weddingID, page, perPage, status, search)
}

type UpdateInput struct {
	WeddingID string
	ID        string
	Name      string
	Phone     string
	Email     string
	Status    string
}

func (uc *UseCase) Update(ctx context.Context, input UpdateInput) (*entity.Guest, error) {
	g, err := uc.guestRepo.FindByID(ctx, input.WeddingID, input.ID)
	if err != nil {
		return nil, err
	}

	g.Name = input.Name
	g.Phone = input.Phone
	g.Email = input.Email
	g.UpdatedAt = time.Now()

	if input.Status != "" && entity.GuestStatus(input.Status) != g.Status {
		g.Status = entity.GuestStatus(input.Status)
		if g.Status == entity.GuestStatusConfirmed && g.ConfirmedAt == nil {
			now := time.Now()
			g.ConfirmedAt = &now
		}
		if g.Status != entity.GuestStatusConfirmed {
			g.ConfirmedAt = nil
		}
	}

	if err := uc.guestRepo.Update(ctx, g); err != nil {
		return nil, fmt.Errorf("guest.Update: %w", err)
	}
	return g, nil
}

func (uc *UseCase) Delete(ctx context.Context, weddingID, id string) error {
	return uc.guestRepo.Delete(ctx, weddingID, id)
}

type DashboardStats struct {
	TotalInvitations int
	TotalGuests      int
	Confirmed        int
	Pending          int
	Declined         int
}

func (uc *UseCase) Dashboard(ctx context.Context, weddingID string) (*DashboardStats, error) {
	totalInv, err := uc.invitationRepo.CountByWedding(ctx, weddingID)
	if err != nil {
		return nil, fmt.Errorf("guest.Dashboard: count invitations: %w", err)
	}

	total, confirmed, pending, declined, err := uc.guestRepo.CountByWedding(ctx, weddingID)
	if err != nil {
		return nil, fmt.Errorf("guest.Dashboard: count guests: %w", err)
	}

	return &DashboardStats{
		TotalInvitations: totalInv,
		TotalGuests:      total,
		Confirmed:        confirmed,
		Pending:          pending,
		Declined:         declined,
	}, nil
}
