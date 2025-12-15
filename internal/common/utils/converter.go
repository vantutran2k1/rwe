package utils

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func PgUUIDToString(input pgtype.UUID) string {
	uid, _ := uuid.FromBytes(input.Bytes[:])
	return uid.String()
}

func UUIDToPgUUID(input uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: input, Valid: true}
}

func TimeToPgTimestamptz(input time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: input, InfinityModifier: pgtype.Finite, Valid: true}
}
