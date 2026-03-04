package sheets

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/by-r2/weddo-api/internal/domain/gateway"
)

type OAuthProvider struct {
	config *oauth2.Config
}

func NewOAuthProvider(clientID, clientSecret, redirectURL string) *OAuthProvider {
	return &OAuthProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/spreadsheets",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (p *OAuthProvider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (p *OAuthProvider) Exchange(ctx context.Context, code string) (*gateway.GoogleToken, error) {
	tok, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("sheets.OAuthProvider.Exchange: %w", err)
	}
	return toDomainToken(tok), nil
}

func (p *OAuthProvider) NewClient(ctx context.Context, token *gateway.GoogleToken, spreadsheetID string) (gateway.GoogleSheetsClient, error) {
	ts := p.config.TokenSource(ctx, toOAuthToken(token))
	return NewClientFromTokenSource(ctx, spreadsheetID, ts)
}

func (p *OAuthProvider) CreateSpreadsheet(ctx context.Context, token *gateway.GoogleToken, title string) (id string, url string, err error) {
	ts := p.config.TokenSource(ctx, toOAuthToken(token))
	svc, err := NewRawServiceFromTokenSource(ctx, ts)
	if err != nil {
		return "", "", err
	}
	return CreateSpreadsheet(ctx, svc, title)
}

func toDomainToken(t *oauth2.Token) *gateway.GoogleToken {
	if t == nil {
		return nil
	}
	return &gateway.GoogleToken{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
	}
}

func toOAuthToken(t *gateway.GoogleToken) *oauth2.Token {
	if t == nil {
		return nil
	}
	return &oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
		TokenType:    "Bearer",
	}
}
