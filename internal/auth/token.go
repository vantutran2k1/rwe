package auth

import (
	"github.com/google/uuid"
	"github.com/vantutran2k1/rwe/internal/common/token"
)

type TokenMaker interface {
	CreateToken(email string, userID uuid.UUID) (string, *token.Payload, error)
	VerifyToken(token string) (*token.Payload, error)
}
