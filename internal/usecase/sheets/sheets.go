package sheets

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/by-r2/weddo-api/internal/domain/entity"
	gw "github.com/by-r2/weddo-api/internal/domain/gateway"
	"github.com/by-r2/weddo-api/internal/domain/repository"
)

var ErrNotConfigured = errors.New("google sheets não configurado")

type UseCase struct {
	invRepo         repository.InvitationRepository
	guestRepo       repository.GuestRepository
	giftRepo        repository.GiftRepository
	payRepo         repository.PaymentRepository
	weddingRepo     repository.WeddingRepository
	integrationRepo repository.GoogleIntegrationRepository
	oauthProvider   gw.GoogleSheetsOAuthProvider
	cipher          TokenCipher
	stateSecret     string
}

type SyncResult struct {
	Invitations int `json:"invitations"`
	Guests      int `json:"guests"`
	Gifts       int `json:"gifts"`
	Payments    int `json:"payments"`
}

type PullResult struct {
	InvitationsUpdated int `json:"invitations_updated"`
	InvitationsCreated int `json:"invitations_created"`
	GuestsUpdated      int `json:"guests_updated"`
	GuestsCreated      int `json:"guests_created"`
	Skipped            int `json:"skipped"`
}

type ConnectStartResult struct {
	AuthURL string `json:"auth_url"`
}

type ConnectCallbackResult struct {
	WeddingID      string `json:"wedding_id"`
	SpreadsheetID  string `json:"spreadsheet_id"`
	SpreadsheetURL string `json:"spreadsheet_url"`
}

type TokenCipher interface {
	Encrypt(plain string) (string, error)
	Decrypt(encrypted string) (string, error)
}

func NewUseCase(
	invRepo repository.InvitationRepository,
	guestRepo repository.GuestRepository,
	giftRepo repository.GiftRepository,
	payRepo repository.PaymentRepository,
	weddingRepo repository.WeddingRepository,
	integrationRepo repository.GoogleIntegrationRepository,
	oauthProvider gw.GoogleSheetsOAuthProvider,
	cipher TokenCipher,
	stateSecret string,
) *UseCase {
	return &UseCase{
		invRepo:         invRepo,
		guestRepo:       guestRepo,
		giftRepo:        giftRepo,
		payRepo:         payRepo,
		weddingRepo:     weddingRepo,
		integrationRepo: integrationRepo,
		oauthProvider:   oauthProvider,
		cipher:          cipher,
		stateSecret:     stateSecret,
	}
}

func (uc *UseCase) Push(ctx context.Context, weddingID string) (*SyncResult, error) {
	client, _, err := uc.clientForWedding(ctx, weddingID)
	if err != nil {
		return nil, err
	}

	invitations, err := uc.listAllInvitations(ctx, weddingID)
	if err != nil {
		return nil, err
	}
	guests, err := uc.listAllGuests(ctx, weddingID)
	if err != nil {
		return nil, err
	}
	gifts, err := uc.listAllGifts(ctx, weddingID)
	if err != nil {
		return nil, err
	}
	payments, err := uc.listAllPayments(ctx, weddingID)
	if err != nil {
		return nil, err
	}

	invByID := make(map[string]entity.Invitation, len(invitations))
	invCodeByID := make(map[string]string, len(invitations))
	for _, inv := range invitations {
		invByID[inv.ID] = inv
		invCodeByID[inv.ID] = inv.Code
	}

	giftByID := make(map[string]entity.Gift, len(gifts))
	for _, g := range gifts {
		giftByID[g.ID] = g
	}

	tabGuests, tabInvitations, tabGifts, tabPayments := defaultTabs()

	if err := client.WriteTab(ctx, tabInvitations, invitationsToSheet(invitations)); err != nil {
		return nil, err
	}
	if err := client.WriteTab(ctx, tabGuests, guestsToSheet(guests, invCodeByID)); err != nil {
		return nil, err
	}
	if err := client.WriteTab(ctx, tabGifts, giftsToSheet(gifts)); err != nil {
		return nil, err
	}
	if err := client.WriteTab(ctx, tabPayments, paymentsToSheet(payments, giftByID)); err != nil {
		return nil, err
	}

	return &SyncResult{
		Invitations: len(invitations),
		Guests:      len(guests),
		Gifts:       len(gifts),
		Payments:    len(payments),
	}, nil
}

func (uc *UseCase) Pull(ctx context.Context, weddingID string) (*PullResult, error) {
	client, _, err := uc.clientForWedding(ctx, weddingID)
	if err != nil {
		return nil, err
	}

	result := &PullResult{}
	tabGuests, tabInvitations, _, _ := defaultTabs()

	invRows, err := client.ReadTab(ctx, tabInvitations)
	if err != nil {
		return nil, err
	}
	if err := uc.pullInvitations(ctx, weddingID, invRows, result); err != nil {
		return nil, err
	}

	guestRows, err := client.ReadTab(ctx, tabGuests)
	if err != nil {
		return nil, err
	}
	if err := uc.pullGuests(ctx, weddingID, guestRows, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (uc *UseCase) StartConnect(ctx context.Context, weddingID string) (*ConnectStartResult, error) {
	if uc.oauthProvider == nil || uc.cipher == nil || uc.integrationRepo == nil {
		return nil, ErrNotConfigured
	}
	state, err := uc.buildState(weddingID)
	if err != nil {
		return nil, err
	}
	return &ConnectStartResult{AuthURL: uc.oauthProvider.AuthCodeURL(state)}, nil
}

func (uc *UseCase) HandleOAuthCallback(ctx context.Context, code, state string) (*ConnectCallbackResult, error) {
	if uc.oauthProvider == nil || uc.cipher == nil || uc.integrationRepo == nil {
		return nil, ErrNotConfigured
	}
	payload, err := uc.parseState(state)
	if err != nil {
		return nil, err
	}

	token, err := uc.oauthProvider.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	w, err := uc.weddingRepo.FindByID(ctx, payload.WeddingID)
	title := "Wedding Guests"
	if err == nil && w != nil && strings.TrimSpace(w.Title) != "" {
		title = "Wedding Guests - " + w.Title
	}

	spreadsheetID, spreadsheetURL, err := uc.oauthProvider.CreateSpreadsheet(ctx, token, title)
	if err != nil {
		return nil, err
	}

	encAccess, err := uc.cipher.Encrypt(token.AccessToken)
	if err != nil {
		return nil, err
	}
	encRefresh, err := uc.cipher.Encrypt(token.RefreshToken)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var expiry *time.Time
	if !token.Expiry.IsZero() {
		t := token.Expiry
		expiry = &t
	}

	if err := uc.integrationRepo.Upsert(ctx, &entity.GoogleIntegration{
		WeddingID:             payload.WeddingID,
		SpreadsheetID:         spreadsheetID,
		SpreadsheetURL:        spreadsheetURL,
		EncryptedAccessToken:  encAccess,
		EncryptedRefreshToken: encRefresh,
		TokenExpiry:           expiry,
		CreatedAt:             now,
		UpdatedAt:             now,
	}); err != nil {
		return nil, err
	}

	return &ConnectCallbackResult{
		WeddingID:      payload.WeddingID,
		SpreadsheetID:  spreadsheetID,
		SpreadsheetURL: spreadsheetURL,
	}, nil
}

func (uc *UseCase) pullInvitations(ctx context.Context, weddingID string, rows [][]string, result *PullResult) error {
	if len(rows) <= 1 {
		return nil
	}
	for _, row := range rows[1:] {
		id := col(row, 0)
		code := strings.TrimSpace(col(row, 1))
		label := strings.TrimSpace(col(row, 2))
		maxGuests := intOrDefault(col(row, 3), 1)
		notes := col(row, 4)

		if code == "" || label == "" {
			result.Skipped++
			continue
		}

		now := time.Now()
		if id == "" {
			inv := &entity.Invitation{
				ID:        uuid.New().String(),
				WeddingID: weddingID,
				Code:      code,
				Label:     label,
				MaxGuests: maxGuests,
				Notes:     notes,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := uc.invRepo.Create(ctx, inv); err != nil {
				result.Skipped++
				continue
			}
			result.InvitationsCreated++
			continue
		}

		inv, err := uc.invRepo.FindByID(ctx, weddingID, id)
		if err != nil {
			result.Skipped++
			continue
		}
		inv.Code = code
		inv.Label = label
		inv.MaxGuests = maxGuests
		inv.Notes = notes
		inv.UpdatedAt = now
		if err := uc.invRepo.Update(ctx, inv); err != nil {
			result.Skipped++
			continue
		}
		result.InvitationsUpdated++
	}
	return nil
}

func (uc *UseCase) pullGuests(ctx context.Context, weddingID string, rows [][]string, result *PullResult) error {
	if len(rows) <= 1 {
		return nil
	}
	for _, row := range rows[1:] {
		id := col(row, 0)
		invCode := strings.TrimSpace(col(row, 1))
		name := strings.TrimSpace(col(row, 2))
		phone := strings.TrimSpace(col(row, 3))
		email := strings.TrimSpace(col(row, 4))
		status := sanitizeGuestStatus(col(row, 5))
		confirmedAtRaw := strings.TrimSpace(col(row, 6))

		if name == "" || invCode == "" {
			result.Skipped++
			continue
		}

		inv, err := uc.invRepo.FindByCode(ctx, weddingID, invCode)
		if err != nil {
			result.Skipped++
			continue
		}

		now := time.Now()
		confirmedAt := parseConfirmedAt(confirmedAtRaw)
		if status != entity.GuestStatusConfirmed {
			confirmedAt = nil
		} else if confirmedAt == nil {
			confirmedAt = &now
		}

		if id == "" {
			guest := &entity.Guest{
				ID:           uuid.New().String(),
				InvitationID: inv.ID,
				WeddingID:    weddingID,
				Name:         name,
				Phone:        phone,
				Email:        email,
				Status:       status,
				ConfirmedAt:  confirmedAt,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := uc.guestRepo.Create(ctx, guest); err != nil {
				result.Skipped++
				continue
			}
			result.GuestsCreated++
			continue
		}

		g, err := uc.guestRepo.FindByID(ctx, weddingID, id)
		if err != nil {
			result.Skipped++
			continue
		}
		g.InvitationID = inv.ID
		g.Name = name
		g.Phone = phone
		g.Email = email
		g.Status = status
		g.ConfirmedAt = confirmedAt
		g.UpdatedAt = now
		if err := uc.guestRepo.Update(ctx, g); err != nil {
			result.Skipped++
			continue
		}
		result.GuestsUpdated++
	}

	return nil
}

func invitationsToSheet(items []entity.Invitation) [][]any {
	rows := [][]any{
		{"ID", "Código", "Grupo", "Max Convidados", "Notas"},
	}
	for _, inv := range items {
		rows = append(rows, []any{inv.ID, inv.Code, inv.Label, inv.MaxGuests, inv.Notes})
	}
	return rows
}

func guestsToSheet(items []entity.Guest, invCodeByID map[string]string) [][]any {
	rows := [][]any{
		{"ID", "Convite", "Nome", "Telefone", "Email", "Status", "Confirmado em"},
	}
	for _, g := range items {
		confirmedAt := ""
		if g.ConfirmedAt != nil {
			confirmedAt = g.ConfirmedAt.Format(time.RFC3339)
		}
		rows = append(rows, []any{
			g.ID,
			invCodeByID[g.InvitationID],
			g.Name,
			g.Phone,
			g.Email,
			string(g.Status),
			confirmedAt,
		})
	}
	return rows
}

func giftsToSheet(items []entity.Gift) [][]any {
	rows := [][]any{
		{"ID", "Nome", "Categoria", "Preço", "Status"},
	}
	for _, g := range items {
		rows = append(rows, []any{g.ID, g.Name, g.Category, g.Price, string(g.Status)})
	}
	return rows
}

func paymentsToSheet(items []entity.Payment, giftByID map[string]entity.Gift) [][]any {
	rows := [][]any{
		{"ID", "Presente", "Quem presenteou", "Valor", "Método", "Status", "Pago em"},
	}
	for _, p := range items {
		paidAt := ""
		if p.PaidAt != nil {
			paidAt = p.PaidAt.Format(time.RFC3339)
		}
		rows = append(rows, []any{
			p.ID,
			giftByID[p.GiftID].Name,
			p.PayerName,
			p.Amount,
			string(p.PaymentMethod),
			string(p.Status),
			paidAt,
		})
	}
	return rows
}

func (uc *UseCase) listAllInvitations(ctx context.Context, weddingID string) ([]entity.Invitation, error) {
	var out []entity.Invitation
	page, perPage := 1, 100
	for {
		items, total, err := uc.invRepo.List(ctx, weddingID, page, perPage, "")
		if err != nil {
			return nil, fmt.Errorf("sheets.listAllInvitations: %w", err)
		}
		out = append(out, items...)
		if len(out) >= total || len(items) == 0 {
			break
		}
		page++
	}
	return out, nil
}

func (uc *UseCase) listAllGuests(ctx context.Context, weddingID string) ([]entity.Guest, error) {
	var out []entity.Guest
	page, perPage := 1, 100
	for {
		items, total, err := uc.guestRepo.List(ctx, weddingID, page, perPage, "", "")
		if err != nil {
			return nil, fmt.Errorf("sheets.listAllGuests: %w", err)
		}
		out = append(out, items...)
		if len(out) >= total || len(items) == 0 {
			break
		}
		page++
	}
	return out, nil
}

func (uc *UseCase) listAllGifts(ctx context.Context, weddingID string) ([]entity.Gift, error) {
	var out []entity.Gift
	page, perPage := 1, 100
	for {
		items, total, err := uc.giftRepo.List(ctx, weddingID, page, perPage, "", "", "")
		if err != nil {
			return nil, fmt.Errorf("sheets.listAllGifts: %w", err)
		}
		out = append(out, items...)
		if len(out) >= total || len(items) == 0 {
			break
		}
		page++
	}
	return out, nil
}

func (uc *UseCase) listAllPayments(ctx context.Context, weddingID string) ([]entity.Payment, error) {
	var out []entity.Payment
	page, perPage := 1, 100
	for {
		items, total, err := uc.payRepo.List(ctx, weddingID, page, perPage, "", "")
		if err != nil {
			return nil, fmt.Errorf("sheets.listAllPayments: %w", err)
		}
		out = append(out, items...)
		if len(out) >= total || len(items) == 0 {
			break
		}
		page++
	}
	return out, nil
}

func col(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func intOrDefault(v string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func sanitizeGuestStatus(v string) entity.GuestStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case string(entity.GuestStatusConfirmed):
		return entity.GuestStatusConfirmed
	case string(entity.GuestStatusDeclined):
		return entity.GuestStatusDeclined
	default:
		return entity.GuestStatusPending
	}
}

func parseConfirmedAt(v string) *time.Time {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return &t
	}
	return nil
}

func defaultTabs() (string, string, string, string) {
	return "Convidados", "Convites", "Presentes", "Pagamentos"
}

type oauthStatePayload struct {
	WeddingID string `json:"wedding_id"`
	Exp       int64  `json:"exp"`
	Nonce     string `json:"nonce"`
}

func (uc *UseCase) buildState(weddingID string) (string, error) {
	payload := oauthStatePayload{
		WeddingID: weddingID,
		Exp:       time.Now().Add(15 * time.Minute).Unix(),
		Nonce:     uuid.New().String(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sig := uc.sign(raw)
	return base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (uc *UseCase) parseState(state string) (*oauthStatePayload, error) {
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("estado OAuth inválido")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("estado OAuth inválido")
	}
	gotSig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("estado OAuth inválido")
	}
	wantSig := uc.sign(raw)
	if !hmac.Equal(gotSig, wantSig) {
		return nil, fmt.Errorf("assinatura do estado OAuth inválida")
	}
	var payload oauthStatePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("estado OAuth inválido")
	}
	if payload.WeddingID == "" || payload.Exp < time.Now().Unix() {
		return nil, fmt.Errorf("estado OAuth expirado")
	}
	return &payload, nil
}

func (uc *UseCase) sign(raw []byte) []byte {
	mac := hmac.New(sha256.New, []byte(uc.stateSecret))
	mac.Write(raw)
	return mac.Sum(nil)
}

func (uc *UseCase) clientForWedding(ctx context.Context, weddingID string) (gw.GoogleSheetsClient, *entity.GoogleIntegration, error) {
	if uc.oauthProvider == nil || uc.cipher == nil || uc.integrationRepo == nil {
		return nil, nil, ErrNotConfigured
	}
	gi, err := uc.integrationRepo.FindByWeddingID(ctx, weddingID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, nil, entity.ErrNotFound
		}
		return nil, nil, err
	}

	access, err := uc.cipher.Decrypt(gi.EncryptedAccessToken)
	if err != nil {
		return nil, nil, err
	}
	refresh, err := uc.cipher.Decrypt(gi.EncryptedRefreshToken)
	if err != nil {
		return nil, nil, err
	}
	expiry := time.Now().Add(-1 * time.Minute)
	if gi.TokenExpiry != nil {
		expiry = *gi.TokenExpiry
	}

	client, err := uc.oauthProvider.NewClient(ctx, &gw.GoogleToken{
		AccessToken:  access,
		RefreshToken: refresh,
		Expiry:       expiry,
	}, gi.SpreadsheetID)
	if err != nil {
		return nil, nil, err
	}
	return client, gi, nil
}
