package auth

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
	duration     time.Duration
}

func NewPasetoMaker(symmetricKey string, duration time.Duration) (TokenMaker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be exactly %d characters", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
		duration:     duration,
	}
	return maker, nil
}

func (p *PasetoMaker) CreateToken(email string, userID uuid.UUID) (string, *TokenPayload, error) {
	payload, err := NewTokenPayload(email, userID, p.duration)
	if err != nil {
		return "", payload, err
	}

	token, err := p.paseto.Encrypt(p.symmetricKey, payload, nil)
	return token, payload, err
}

func (p *PasetoMaker) VerifyToken(token string) (*TokenPayload, error) {
	payload := &TokenPayload{}

	if err := p.paseto.Decrypt(token, p.symmetricKey, payload, nil); err != nil {
		return nil, ErrInvalidToken
	}

	if err := payload.Validate(); err != nil {
		return nil, err
	}

	return payload, nil
}
