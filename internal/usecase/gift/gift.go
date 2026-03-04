package gift

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type UseCase struct {
	giftRepo    repository.GiftRepository
	paymentRepo repository.PaymentRepository
}

func NewUseCase(gr repository.GiftRepository, pr repository.PaymentRepository) *UseCase {
	return &UseCase{giftRepo: gr, paymentRepo: pr}
}

type CreateInput struct {
	WeddingID   string
	Name        string
	Description string
	Price       float64
	ImageURL    string
	Category    string
}

func (uc *UseCase) Create(ctx context.Context, input CreateInput) (*entity.Gift, error) {
	now := time.Now()
	g := &entity.Gift{
		ID:          uuid.New().String(),
		WeddingID:   input.WeddingID,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		ImageURL:    input.ImageURL,
		Category:    input.Category,
		Status:      entity.GiftStatusAvailable,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.giftRepo.Create(ctx, g); err != nil {
		return nil, fmt.Errorf("gift.Create: %w", err)
	}
	return g, nil
}

func (uc *UseCase) FindByID(ctx context.Context, weddingID, id string) (*entity.Gift, error) {
	return uc.giftRepo.FindByID(ctx, weddingID, id)
}

func (uc *UseCase) List(ctx context.Context, weddingID string, page, perPage int, category, status, search string) ([]entity.Gift, int, error) {
	return uc.giftRepo.List(ctx, weddingID, page, perPage, category, status, search)
}

type UpdateInput struct {
	WeddingID   string
	ID          string
	Name        string
	Description string
	Price       float64
	ImageURL    string
	Category    string
	Status      string
}

func (uc *UseCase) Update(ctx context.Context, input UpdateInput) (*entity.Gift, error) {
	g, err := uc.giftRepo.FindByID(ctx, input.WeddingID, input.ID)
	if err != nil {
		return nil, err
	}

	g.Name = input.Name
	g.Description = input.Description
	g.Price = input.Price
	g.ImageURL = input.ImageURL
	g.Category = input.Category
	g.UpdatedAt = time.Now()

	if input.Status != "" {
		g.Status = entity.GiftStatus(input.Status)
	}

	if err := uc.giftRepo.Update(ctx, g); err != nil {
		return nil, fmt.Errorf("gift.Update: %w", err)
	}
	return g, nil
}

func (uc *UseCase) Delete(ctx context.Context, weddingID, id string) error {
	return uc.giftRepo.Delete(ctx, weddingID, id)
}

type DashboardStats struct {
	TotalGifts    int
	Purchased     int
	Available     int
	TotalRevenue  float64
	TotalPayments int
}

func (uc *UseCase) Dashboard(ctx context.Context, weddingID string) (*DashboardStats, error) {
	total, available, purchased, err := uc.giftRepo.CountByWedding(ctx, weddingID)
	if err != nil {
		return nil, fmt.Errorf("gift.Dashboard: count: %w", err)
	}

	revenue, payments, err := uc.paymentRepo.SumByWedding(ctx, weddingID)
	if err != nil {
		return nil, fmt.Errorf("gift.Dashboard: sum: %w", err)
	}

	return &DashboardStats{
		TotalGifts:    total,
		Purchased:     purchased,
		Available:     available,
		TotalRevenue:  revenue,
		TotalPayments: payments,
	}, nil
}
