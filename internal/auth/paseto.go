package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"github.com/vantutran2k1/rwe/internal/common/token"
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

func (p *PasetoMaker) CreateToken(email string, userID uuid.UUID) (string, *token.Payload, error) {
	payload, err := token.NewPayload(email, userID, p.duration)
	if err != nil {
		return "", payload, err
	}

	t, err := p.paseto.Encrypt(p.symmetricKey, payload, nil)
	return t, payload, err
}

func (p *PasetoMaker) VerifyToken(t string) (*token.Payload, error) {
	payload := &token.Payload{}

	if err := p.paseto.Decrypt(t, p.symmetricKey, payload, nil); err != nil {
		return nil, errors.New("token is invalid")
	}

	if err := payload.Validate(); err != nil {
		return nil, err
	}

	return payload, nil
}
