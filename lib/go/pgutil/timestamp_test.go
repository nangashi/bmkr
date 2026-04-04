package pgutil_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nangashi/bmkr/lib/go/pgutil"
)

func TestPgTimestampToProto(t *testing.T) {
	fixedTime := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		input   pgtype.Timestamptz
		wantNil bool
		wantTs  *timestamppb.Timestamp
	}{
		{
			name:    "Valid=true returns non-nil timestamppb.Timestamp with correct time",
			input:   pgtype.Timestamptz{Time: fixedTime, Valid: true},
			wantNil: false,
			wantTs:  timestamppb.New(fixedTime),
		},
		{
			name:    "Valid=false (NULL) returns nil",
			input:   pgtype.Timestamptz{Valid: false},
			wantNil: true,
			wantTs:  nil,
		},
		{
			name:    "zero value (Valid=false) returns nil",
			input:   pgtype.Timestamptz{},
			wantNil: true,
			wantTs:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pgutil.PgTimestampToProto(tt.input)

			if tt.wantNil {
				if got != nil {
					t.Errorf("PgTimestampToProto(%v) = %v, want nil", tt.input, got)
				}
				return
			}

			if got == nil {
				t.Fatalf("PgTimestampToProto(%v) = nil, want non-nil", tt.input)
			}

			if !got.AsTime().Equal(tt.wantTs.AsTime()) {
				t.Errorf("PgTimestampToProto(%v).AsTime() = %v, want %v", tt.input, got.AsTime(), tt.wantTs.AsTime())
			}
		})
	}
}
