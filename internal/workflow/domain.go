package workflow

import (
	"time"

	"github.com/google/uuid"
)

type Workflow struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Name       string
	Version    int
	Definition []byte
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Archived   bool
}
