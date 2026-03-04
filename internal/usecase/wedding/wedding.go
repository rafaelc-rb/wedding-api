package wedding

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

type UseCase struct {
	repo      repository.WeddingRepository
	jwtSecret string
	jwtExpH   int
}

func NewUseCase(repo repository.WeddingRepository, jwtSecret string, jwtExpH int) *UseCase {
	return &UseCase{repo: repo, jwtSecret: jwtSecret, jwtExpH: jwtExpH}
}

// Authenticate valida email+senha e retorna um JWT com wedding_id nos claims.
func (uc *UseCase) Authenticate(ctx context.Context, email, password string) (string, *entity.Wedding, error) {
	w, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, entity.ErrUnauthorized
	}

	if !w.Active {
		return "", nil, entity.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(w.AdminPassHash), []byte(password)); err != nil {
		return "", nil, entity.ErrUnauthorized
	}

	claims := jwt.MapClaims{
		"wedding_id": w.ID,
		"email":      w.AdminEmail,
		"exp":        time.Now().Add(time.Duration(uc.jwtExpH) * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return "", nil, fmt.Errorf("wedding.Authenticate: sign token: %w", err)
	}

	return signed, w, nil
}

// Seed cria um wedding se não existir (usado no boot para o primeiro tenant).
func (uc *UseCase) Seed(ctx context.Context, slug, title, date, partner1, partner2, email, password string) error {
	_, err := uc.repo.FindByEmail(ctx, email)
	if err == nil {
		return nil // já existe
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("wedding.Seed: hash password: %w", err)
	}

	now := time.Now()
	w := &entity.Wedding{
		ID:            uuid.New().String(),
		Slug:          slug,
		Title:         title,
		Date:          date,
		Partner1Name:  partner1,
		Partner2Name:  partner2,
		AdminEmail:    email,
		AdminPassHash: string(hash),
		Active:        true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := uc.repo.Create(ctx, w); err != nil {
		return fmt.Errorf("wedding.Seed: %w", err)
	}

	return nil
}
