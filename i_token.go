package oauth2

import "time"

// Token describes a token's characteristic, it can be either access token or refresh token.
type Token interface {

	// Return client's ID.
	ClientID() string

	// Return user's ID.
	UserID() string

	// Return token.
	Token() string

	// Check if token is expired or not.
	IsExpired() bool

	// Return token's created time.
	CreatedTime() time.Time

	// Return token's expired time.
	ExpiredTime() time.Time
}
