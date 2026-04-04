// Package pgutil provides helpers for converting pgx/pgtype values to
// protobuf well-known types.
package pgutil

import (
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PgTimestampToProto converts a pgtype.Timestamptz to a *timestamppb.Timestamp.
//
// wip: check ts.Valid; return nil when false (NULL column value).
// Otherwise return timestamppb.New(ts.Time).
// Pure function — no side effects, safe for concurrent use.
func PgTimestampToProto(ts pgtype.Timestamptz) *timestamppb.Timestamp {
	panic("not implemented")
}
