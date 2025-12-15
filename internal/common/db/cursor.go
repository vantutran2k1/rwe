package db

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PaginationCursor struct {
	LastID        uuid.UUID `json:"last_id"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

func EncodeCursor(lastID uuid.UUID, lastUpdatedAt time.Time) string {
	cursor := PaginationCursor{
		LastID:        lastID,
		LastUpdatedAt: lastUpdatedAt,
	}

	data, _ := json.Marshal(cursor)
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeCursor(token string) (*PaginationCursor, error) {
	if token == "" {
		return getDefaultCursor(), nil
	}

	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}

	var cursor PaginationCursor
	err = json.Unmarshal(data, &cursor)
	return &cursor, err
}

func getDefaultCursor() *PaginationCursor {
	var maxBytes [16]byte
	for i := range maxBytes {
		maxBytes[i] = 0xff
	}

	maxTime := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

	return &PaginationCursor{
		LastID:        maxBytes,
		LastUpdatedAt: maxTime,
	}
}
