package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type TokenMaker interface {
	CreateToken(email string, userID uuid.UUID) (string, *TokenPayload, error)
	VerifyToken(token string) (*TokenPayload, error)
}

type TokenPayload struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewTokenPayload(email string, userID uuid.UUID, duration time.Duration) (*TokenPayload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &TokenPayload{
		ID:        tokenID,
		UserID:    userID,
		Email:     email,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}

func (p *TokenPayload) Validate() error {
	if time.Now().After(p.ExpiredAt) {
		return ErrExpiredToken
	}

	return nil
}
