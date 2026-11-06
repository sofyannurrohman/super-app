package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

func TestBankAccountDeleteError(t *testing.T) {
	other := errors.New("database unavailable")
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"pgx foreign key", &pgconn.PgError{Code: "23503"}, domain.ErrBankAccountInUse},
		{"pq foreign key", &pq.Error{Code: "23503"}, domain.ErrBankAccountInUse},
		{"gorm foreign key", gorm.ErrForeignKeyViolated, domain.ErrBankAccountInUse},
		{"other database error", other, other},
		{"no error", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bankAccountDeleteError(tt.err); !errors.Is(got, tt.want) {
				t.Fatalf("bankAccountDeleteError() = %v, want %v", got, tt.want)
			}
		})
	}
}
