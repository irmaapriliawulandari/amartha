package repo

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"amartha-test/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoanRepo_GetStatements(t *testing.T) {
	loanID := int64(1)
	until := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	limit, offset := 10, 0

	queryPattern := regexp.QuoteMeta("from statement")
	columns := []string{"statement_id", "installment_amount", "carry_over_amount", "paid_amount", "statement_date", "status"}

	t.Run("success scans rows into StatementDB", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows(columns).
			AddRow(int64(1), "100000", "0", "0", time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC), 0).
			AddRow(int64(2), "100000", "0", "100000", time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC), 1)

		mock.ExpectQuery(queryPattern).
			WithArgs(loanID, until, limit, offset).
			WillReturnRows(rows)

		r := &loanRepo{db: db}
		got, err := r.GetStatements(loanID, until, limit, offset)

		require.NoError(t, err)
		assert.Equal(t, []entity.StatementDB{
			{
				StatementID: 1, LoanID: loanID, StatementDate: time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
				InstallmentAmount: decimal.NewFromInt(100000), CarryOverAmount: decimal.NewFromInt(0), PaidAmount: decimal.NewFromInt(0),
				Status: entity.StatementStatusUnpaid,
			},
			{
				StatementID: 2, LoanID: loanID, StatementDate: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
				InstallmentAmount: decimal.NewFromInt(100000), CarryOverAmount: decimal.NewFromInt(0), PaidAmount: decimal.NewFromInt(100000),
				Status: entity.StatementStatusPaid,
			},
		}, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(loanID, until, limit, offset).
			WillReturnError(errors.New("connection refused"))

		r := &loanRepo{db: db}
		got, err := r.GetStatements(loanID, until, limit, offset)

		assert.Nil(t, got)
		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("row iteration error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows(columns).
			AddRow(int64(1), "100000", "0", "0", time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC), 0).
			RowError(0, errors.New("row read failure"))

		mock.ExpectQuery(queryPattern).
			WithArgs(loanID, until, limit, offset).
			WillReturnRows(rows)

		r := &loanRepo{db: db}
		got, err := r.GetStatements(loanID, until, limit, offset)

		assert.Nil(t, got)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no rows returns empty slice and no error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows(columns)

		mock.ExpectQuery(queryPattern).
			WithArgs(loanID, until, limit, offset).
			WillReturnRows(rows)

		r := &loanRepo{db: db}
		got, err := r.GetStatements(loanID, until, limit, offset)

		assert.NoError(t, err)
		assert.Empty(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_InsertStatement(t *testing.T) {
	statement := entity.StatementDB{
		LoanID:            1,
		InstallmentAmount: decimal.NewFromInt(110000),
		CarryOverAmount:   decimal.NewFromInt(0),
		PaidAmount:        decimal.NewFromInt(0),
		StatementDate:     time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Deadline:          time.Date(2026, 8, 8, 0, 0, 0, 0, time.UTC),
		Status:            entity.StatementStatusUnpaid,
		PaidAt:            sql.NullTime{},
		UpdatedAt:         sql.NullTime{},
	}

	queryPattern := regexp.QuoteMeta("insert into statement")

	t.Run("success returns generated statement_id", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"statement_id"}).AddRow(int64(10))

		mock.ExpectQuery(queryPattern).
			WithArgs(
				statement.LoanID,
				statement.InstallmentAmount,
				statement.CarryOverAmount,
				statement.PaidAmount,
				statement.StatementDate,
				statement.Deadline,
				statement.Status,
				statement.PaidAt,
				statement.UpdatedAt,
			).
			WillReturnRows(rows)

		r := &loanRepo{db: db}
		got, err := r.InsertStatement(statement)

		require.NoError(t, err)
		assert.Equal(t, int64(10), got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(
				statement.LoanID,
				statement.InstallmentAmount,
				statement.CarryOverAmount,
				statement.PaidAmount,
				statement.StatementDate,
				statement.Deadline,
				statement.Status,
				statement.PaidAt,
				statement.UpdatedAt,
			).
			WillReturnError(errors.New("connection refused"))

		r := &loanRepo{db: db}
		got, err := r.InsertStatement(statement)

		assert.Zero(t, got)
		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("conflict on loan_id+statement_date is not an error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(
				statement.LoanID,
				statement.InstallmentAmount,
				statement.CarryOverAmount,
				statement.PaidAmount,
				statement.StatementDate,
				statement.Deadline,
				statement.Status,
				statement.PaidAt,
				statement.UpdatedAt,
			).
			WillReturnRows(sqlmock.NewRows([]string{"statement_id"}))

		r := &loanRepo{db: db}
		got, err := r.InsertStatement(statement)

		require.NoError(t, err)
		assert.Zero(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
