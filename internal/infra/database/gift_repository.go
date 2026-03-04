package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/entity"
	"github.com/rafaeljurkfitz/mr-wedding-api/internal/domain/repository"
)

type giftRepository struct {
	db *sql.DB
}

func NewGiftRepository(db *sql.DB) repository.GiftRepository {
	return &giftRepository{db: db}
}

func (r *giftRepository) Create(ctx context.Context, g *entity.Gift) error {
	query := `
		INSERT INTO gifts (id, wedding_id, name, description, price, image_url, category, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		g.ID, g.WeddingID, g.Name, g.Description, g.Price, g.ImageURL,
		g.Category, g.Status, g.CreatedAt, g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("giftRepository.Create: %w", err)
	}
	return nil
}

func (r *giftRepository) FindByID(ctx context.Context, weddingID, id string) (*entity.Gift, error) {
	query := `
		SELECT id, wedding_id, name, description, price, image_url, category, status, created_at, updated_at
		FROM gifts WHERE wedding_id = ? AND id = ?`

	var g entity.Gift
	err := r.db.QueryRowContext(ctx, query, weddingID, id).Scan(
		&g.ID, &g.WeddingID, &g.Name, &g.Description, &g.Price, &g.ImageURL,
		&g.Category, &g.Status, &g.CreatedAt, &g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("giftRepository.FindByID: %w", err)
	}
	return &g, nil
}

func (r *giftRepository) List(ctx context.Context, weddingID string, page, perPage int, category, status, search string) ([]entity.Gift, int, error) {
	countQuery := `SELECT COUNT(*) FROM gifts WHERE wedding_id = ?`
	listQuery := `
		SELECT id, wedding_id, name, description, price, image_url, category, status, created_at, updated_at
		FROM gifts WHERE wedding_id = ?`

	args := []any{weddingID}

	if category != "" {
		f := ` AND category = ?`
		countQuery += f
		listQuery += f
		args = append(args, category)
	}
	if status != "" {
		f := ` AND status = ?`
		countQuery += f
		listQuery += f
		args = append(args, status)
	}
	if search != "" {
		f := ` AND name LIKE ?`
		countQuery += f
		listQuery += f
		args = append(args, "%"+search+"%")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("giftRepository.List: count: %w", err)
	}

	listQuery += ` ORDER BY category, name LIMIT ? OFFSET ?`
	offset := (page - 1) * perPage
	listArgs := append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("giftRepository.List: query: %w", err)
	}
	defer rows.Close()

	var gifts []entity.Gift
	for rows.Next() {
		var g entity.Gift
		if err := rows.Scan(
			&g.ID, &g.WeddingID, &g.Name, &g.Description, &g.Price, &g.ImageURL,
			&g.Category, &g.Status, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("giftRepository.List: scan: %w", err)
		}
		gifts = append(gifts, g)
	}
	return gifts, total, nil
}

func (r *giftRepository) Update(ctx context.Context, g *entity.Gift) error {
	query := `
		UPDATE gifts SET name = ?, description = ?, price = ?, image_url = ?, category = ?, status = ?, updated_at = ?
		WHERE wedding_id = ? AND id = ?`

	res, err := r.db.ExecContext(ctx, query,
		g.Name, g.Description, g.Price, g.ImageURL, g.Category, g.Status, g.UpdatedAt,
		g.WeddingID, g.ID,
	)
	if err != nil {
		return fmt.Errorf("giftRepository.Update: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *giftRepository) Delete(ctx context.Context, weddingID, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM gifts WHERE wedding_id = ? AND id = ?`, weddingID, id)
	if err != nil {
		return fmt.Errorf("giftRepository.Delete: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *giftRepository) CountByWedding(ctx context.Context, weddingID string) (total, available, purchased int, err error) {
	query := `
		SELECT
			COUNT(*),
			COUNT(CASE WHEN status = 'available' THEN 1 END),
			COUNT(CASE WHEN status = 'purchased' THEN 1 END)
		FROM gifts WHERE wedding_id = ?`

	err = r.db.QueryRowContext(ctx, query, weddingID).Scan(&total, &available, &purchased)
	if err != nil {
		err = fmt.Errorf("giftRepository.CountByWedding: %w", err)
	}
	return
}
