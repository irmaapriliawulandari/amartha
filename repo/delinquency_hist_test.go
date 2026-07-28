package repo

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"amartha-test/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoanRepo_InsertDelinquencyHist(t *testing.T) {
	dh := entity.DelinquencyHistDB{
		BorrowerID:  1,
		LoanID:      2,
		StatementID: 3,
		ClearedAt:   sql.NullTime{},
		UpdatedAt:   sql.NullTime{},
	}

	queryPattern := regexp.QuoteMeta("insert into delinquency_hist")

	t.Run("success returns generated dh_id", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"dh_id"}).AddRow(int64(7))

		mock.ExpectQuery(queryPattern).
			WithArgs(dh.BorrowerID, dh.LoanID, dh.StatementID, dh.ClearedAt, dh.UpdatedAt).
			WillReturnRows(rows)

		r := &loanRepo{db: db}
		got, err := r.InsertDelinquencyHist(dh)

		require.NoError(t, err)
		assert.Equal(t, int64(7), got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(dh.BorrowerID, dh.LoanID, dh.StatementID, dh.ClearedAt, dh.UpdatedAt).
			WillReturnError(errors.New("connection refused"))

		r := &loanRepo{db: db}
		got, err := r.InsertDelinquencyHist(dh)

		assert.Zero(t, got)
		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("conflict on statement_id is not an error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(dh.BorrowerID, dh.LoanID, dh.StatementID, dh.ClearedAt, dh.UpdatedAt).
			WillReturnRows(sqlmock.NewRows([]string{"dh_id"}))

		r := &loanRepo{db: db}
		got, err := r.InsertDelinquencyHist(dh)

		require.NoError(t, err)
		assert.Zero(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
