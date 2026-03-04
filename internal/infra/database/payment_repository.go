package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/by-r2/weddo-api/internal/domain/entity"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) repository.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, p *entity.Payment) error {
	query := `
		INSERT INTO payments (id, gift_id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.GiftID, p.WeddingID, p.ProviderID, p.Amount, p.Status, p.PaymentMethod,
		p.PayerName, p.PayerEmail, p.Message, p.PixQRCode, p.PixExpiration,
		p.PaidAt, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("paymentRepository.Create: %w", err)
	}
	return nil
}

func (r *paymentRepository) FindByID(ctx context.Context, weddingID, id string) (*entity.Payment, error) {
	query := `
		SELECT id, gift_id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at
		FROM payments WHERE wedding_id = $1 AND id = $2`

	return r.scanPayment(r.db.QueryRowContext(ctx, query, weddingID, id))
}

func (r *paymentRepository) FindByProviderID(ctx context.Context, providerID string) (*entity.Payment, error) {
	query := `
		SELECT id, gift_id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at
		FROM payments WHERE provider_id = $1`

	return r.scanPayment(r.db.QueryRowContext(ctx, query, providerID))
}

func (r *paymentRepository) List(ctx context.Context, weddingID string, page, perPage int, status, giftID string) ([]entity.Payment, int, error) {
	countQuery := `SELECT COUNT(*) FROM payments WHERE wedding_id = $1`
	listQuery := `
		SELECT id, gift_id, wedding_id, provider_id, amount, status, payment_method,
			payer_name, payer_email, message, pix_qr_code, pix_expiration, paid_at, created_at, updated_at
		FROM payments WHERE wedding_id = $1`

	args := []any{weddingID}
	paramIdx := 2

	if status != "" {
		f := fmt.Sprintf(` AND status = $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, status)
		paramIdx++
	}
	if giftID != "" {
		f := fmt.Sprintf(` AND gift_id = $%d`, paramIdx)
		countQuery += f
		listQuery += f
		args = append(args, giftID)
		paramIdx++
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("paymentRepository.List: count: %w", err)
	}

	listQuery += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, paramIdx, paramIdx+1)
	offset := (page - 1) * perPage
	listArgs := append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("paymentRepository.List: query: %w", err)
	}
	defer rows.Close()

	var payments []entity.Payment
	for rows.Next() {
		var p entity.Payment
		if err := rows.Scan(
			&p.ID, &p.GiftID, &p.WeddingID, &p.ProviderID, &p.Amount, &p.Status, &p.PaymentMethod,
			&p.PayerName, &p.PayerEmail, &p.Message, &p.PixQRCode, &p.PixExpiration,
			&p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("paymentRepository.List: scan: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, total, nil
}

func (r *paymentRepository) Update(ctx context.Context, p *entity.Payment) error {
	query := `
		UPDATE payments SET provider_id = $1, status = $2, pix_qr_code = $3, pix_expiration = $4,
			paid_at = $5, updated_at = $6
		WHERE id = $7`

	res, err := r.db.ExecContext(ctx, query,
		p.ProviderID, p.Status, p.PixQRCode, p.PixExpiration, p.PaidAt, p.UpdatedAt,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("paymentRepository.Update: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *paymentRepository) SumByWedding(ctx context.Context, weddingID string) (totalRevenue float64, totalPayments int, err error) {
	query := `
		SELECT COALESCE(SUM(amount), 0), COUNT(*)
		FROM payments WHERE wedding_id = $1 AND status = 'approved'`

	err = r.db.QueryRowContext(ctx, query, weddingID).Scan(&totalRevenue, &totalPayments)
	if err != nil {
		err = fmt.Errorf("paymentRepository.SumByWedding: %w", err)
	}
	return
}

func (r *paymentRepository) scanPayment(row *sql.Row) (*entity.Payment, error) {
	var p entity.Payment
	err := row.Scan(
		&p.ID, &p.GiftID, &p.WeddingID, &p.ProviderID, &p.Amount, &p.Status, &p.PaymentMethod,
		&p.PayerName, &p.PayerEmail, &p.Message, &p.PixQRCode, &p.PixExpiration,
		&p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("paymentRepository.scanPayment: %w", err)
	}
	return &p, nil
}
