package repo

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

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

func TestLoanRepo_IsDelinquent(t *testing.T) {
	borrowerID := int64(1)
	queryPattern := regexp.QuoteMeta("from delinquency_hist")

	t.Run("true when an uncleared record exists", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(borrowerID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		r := &loanRepo{db: db}
		got, err := r.IsDelinquent(borrowerID)

		require.NoError(t, err)
		assert.True(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("false when no uncleared record exists", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(borrowerID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		r := &loanRepo{db: db}
		got, err := r.IsDelinquent(borrowerID)

		require.NoError(t, err)
		assert.False(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(borrowerID).
			WillReturnError(errors.New("connection refused"))

		r := &loanRepo{db: db}
		_, err = r.IsDelinquent(borrowerID)

		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_IsEverDelinquent(t *testing.T) {
	borrowerID := int64(1)
	queryPattern := regexp.QuoteMeta("from delinquency_hist")

	t.Run("true when any record exists, cleared or not", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(borrowerID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		r := &loanRepo{db: db}
		got, err := r.IsEverDelinquent(borrowerID)

		require.NoError(t, err)
		assert.True(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("false when no record exists", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(borrowerID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		r := &loanRepo{db: db}
		got, err := r.IsEverDelinquent(borrowerID)

		require.NoError(t, err)
		assert.False(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(borrowerID).
			WillReturnError(errors.New("connection refused"))

		r := &loanRepo{db: db}
		_, err = r.IsEverDelinquent(borrowerID)

		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_IsLoanDelinquent(t *testing.T) {
	borrowerID, loanID := int64(1), int64(2)
	queryPattern := regexp.QuoteMeta("from delinquency_hist")

	t.Run("true when an uncleared record exists for that loan", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(borrowerID, loanID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		r := &loanRepo{db: db}
		got, err := r.IsLoanDelinquent(borrowerID, loanID)

		require.NoError(t, err)
		assert.True(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("false when no uncleared record exists for that loan", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(borrowerID, loanID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		r := &loanRepo{db: db}
		got, err := r.IsLoanDelinquent(borrowerID, loanID)

		require.NoError(t, err)
		assert.False(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(borrowerID, loanID).
			WillReturnError(errors.New("connection refused"))

		r := &loanRepo{db: db}
		_, err = r.IsLoanDelinquent(borrowerID, loanID)

		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_InsertDelinquencyHistTx(t *testing.T) {
	dh := entity.DelinquencyHistDB{BorrowerID: 1, LoanID: 2, StatementID: 3}
	pattern := regexp.QuoteMeta("insert into delinquency_hist")

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		require.NoError(t, err)

		mock.ExpectQuery(pattern).
			WithArgs(dh.BorrowerID, dh.LoanID, dh.StatementID, dh.ClearedAt, dh.UpdatedAt).
			WillReturnRows(sqlmock.NewRows([]string{"dh_id"}).AddRow(int64(7)))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		err = r.InsertDelinquencyHistTx(tx, dh)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_ClearDelinquency(t *testing.T) {
	loanID := int64(1)
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	pattern := regexp.QuoteMeta("update delinquency_hist")

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		require.NoError(t, err)

		mock.ExpectExec(pattern).
			WithArgs(now, loanID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		err = r.ClearDelinquency(tx, loanID, now)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		tx, err := db.Begin()
		require.NoError(t, err)

		mock.ExpectExec(pattern).
			WithArgs(now, loanID).
			WillReturnError(errors.New("connection refused"))
		mock.ExpectRollback()

		r := &loanRepo{db: db}
		err = r.ClearDelinquency(tx, loanID, now)
		require.NoError(t, tx.Rollback())

		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
