// Package pgutil provides helpers for converting pgx/pgtype values to
// protobuf well-known types.
package pgutil

import (
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PgTimestampToProto converts a pgtype.Timestamptz to a *timestamppb.Timestamp.
// Returns nil when ts.Valid is false (NULL column value).
func PgTimestampToProto(ts pgtype.Timestamptz) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}
	return timestamppb.New(ts.Time)
}
