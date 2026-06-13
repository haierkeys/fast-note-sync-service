package domain

import (
	"context"
	"time"
)

// OIDCIdentity links an external OIDC subject to a local user.
type OIDCIdentity struct {
	ID        int64
	UID       int64
	Issuer    string
	Subject   string
	Email     string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OIDCIdentityRepository stores OIDC-to-local-user bindings.
type OIDCIdentityRepository interface {
	GetByIssuerSubject(ctx context.Context, issuer, subject string) (*OIDCIdentity, error)
	Create(ctx context.Context, identity *OIDCIdentity) (*OIDCIdentity, error)
}
