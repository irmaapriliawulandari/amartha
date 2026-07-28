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

func TestLoanRepo_GetLatestStatement(t *testing.T) {
	loanID := int64(1)
	before := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)

	queryPattern := regexp.QuoteMeta("from statement")
	columns := []string{"statement_id", "installment_amount", "carry_over_amount", "paid_amount", "statement_date", "deadline", "status"}

	t.Run("success returns the latest statement regardless of status", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows(columns).
			AddRow(int64(5), "110000", "10000", "120000", time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC), entity.StatementStatusPaid)

		mock.ExpectQuery(queryPattern).
			WithArgs(loanID, before).
			WillReturnRows(rows)

		r := &loanRepo{db: db}
		got, err := r.GetLatestStatement(loanID, before)

		require.NoError(t, err)
		assert.Equal(t, entity.StatementDB{
			StatementID: 5, LoanID: loanID,
			InstallmentAmount: decimal.NewFromInt(110000), CarryOverAmount: decimal.NewFromInt(10000), PaidAmount: decimal.NewFromInt(120000),
			StatementDate: time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC), Deadline: time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC),
			Status: entity.StatementStatusPaid,
		}, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no statement before the reference date returns ErrStatementNotFound", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(loanID, before).
			WillReturnRows(sqlmock.NewRows(columns))

		r := &loanRepo{db: db}
		_, err = r.GetLatestStatement(loanID, before)

		assert.ErrorIs(t, err, entity.ErrStatementNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(loanID, before).
			WillReturnError(errors.New("connection refused"))

		r := &loanRepo{db: db}
		_, err = r.GetLatestStatement(loanID, before)

		assert.ErrorContains(t, err, "connection refused")
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


func beginTx(t *testing.T, db *sql.DB, mock sqlmock.Sqlmock) *sql.Tx {
	t.Helper()
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	return tx
}

func TestLoanRepo_GetLatestStatementForUpdate(t *testing.T) {
	loanID := int64(1)
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	selectPattern := regexp.QuoteMeta("for update")
	columns := []string{"statement_id", "installment_amount", "carry_over_amount", "paid_amount", "statement_date", "deadline", "status"}

	t.Run("success returns and locks the latest statement", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectQuery(selectPattern).
			WithArgs(loanID, now).
			WillReturnRows(sqlmock.NewRows(columns).
				AddRow(int64(5), "110000", "10000", "0", time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC), entity.StatementStatusUnpaid))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		got, err := r.GetLatestStatementForUpdate(tx, loanID, now)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.Equal(t, int64(5), got.StatementID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no statement before the reference date returns ErrStatementNotFound", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectQuery(selectPattern).
			WithArgs(loanID, now).
			WillReturnRows(sqlmock.NewRows(columns))
		mock.ExpectRollback()

		r := &loanRepo{db: db}
		_, err = r.GetLatestStatementForUpdate(tx, loanID, now)
		require.NoError(t, tx.Rollback())

		assert.ErrorIs(t, err, entity.ErrStatementNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid transaction handle returns an error", func(t *testing.T) {
		r := &loanRepo{}
		_, err := r.GetLatestStatementForUpdate(nil, loanID, now)
		assert.Error(t, err)
	})
}

func TestLoanRepo_UpdateStatementPaid(t *testing.T) {
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	pattern := regexp.QuoteMeta("set paid_amount")

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectExec(pattern).
			WithArgs(decimal.NewFromInt(120000), entity.StatementStatusPaid, now, int64(5)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		err = r.UpdateStatementPaid(tx, 5, decimal.NewFromInt(120000), now)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectExec(pattern).
			WithArgs(decimal.NewFromInt(120000), entity.StatementStatusPaid, now, int64(5)).
			WillReturnError(errors.New("connection refused"))
		mock.ExpectRollback()

		r := &loanRepo{db: db}
		err = r.UpdateStatementPaid(tx, 5, decimal.NewFromInt(120000), now)
		require.NoError(t, tx.Rollback())

		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_MarkPriorOverdueAsPaidLate(t *testing.T) {
	loanID := int64(1)
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	pattern := regexp.QuoteMeta("where loan_id = $3 and status = $4")

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectExec(pattern).
			WithArgs(entity.StatementStatusPaidLate, now, loanID, entity.StatementStatusOverdue).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		err = r.MarkPriorOverdueAsPaidLate(tx, loanID, now)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_ListOverdueCandidates(t *testing.T) {
	deadline := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	pattern := regexp.QuoteMeta("join loan l")

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(pattern).
			WithArgs(entity.StatementStatusUnpaid, deadline, entity.LoanStatusActive).
			WillReturnRows(sqlmock.NewRows([]string{"statement_id", "loan_id", "borrower_id"}).
				AddRow(int64(10), int64(1), int64(2)))

		r := &loanRepo{db: db}
		got, err := r.ListOverdueCandidates(deadline)

		require.NoError(t, err)
		assert.Equal(t, []entity.OverdueCandidate{{StatementID: 10, LoanID: 1, BorrowerID: 2}}, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(pattern).
			WithArgs(entity.StatementStatusUnpaid, deadline, entity.LoanStatusActive).
			WillReturnError(errors.New("connection refused"))

		r := &loanRepo{db: db}
		_, err = r.ListOverdueCandidates(deadline)

		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_GetStatementForUpdate(t *testing.T) {
	pattern := regexp.QuoteMeta("where statement_id = $1")
	columns := []string{"loan_id", "installment_amount", "carry_over_amount", "paid_amount", "statement_date", "deadline", "status"}

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectQuery(pattern).
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows(columns).
				AddRow(int64(1), "110000", "0", "0", time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC), entity.StatementStatusUnpaid))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		got, err := r.GetStatementForUpdate(tx, 10)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.Equal(t, int64(10), got.StatementID)
		assert.Equal(t, int64(1), got.LoanID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found returns ErrStatementNotFound", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectQuery(pattern).
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows(columns))
		mock.ExpectRollback()

		r := &loanRepo{db: db}
		_, err = r.GetStatementForUpdate(tx, 10)
		require.NoError(t, tx.Rollback())

		assert.ErrorIs(t, err, entity.ErrStatementNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_MarkStatementOverdue(t *testing.T) {
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	pattern := regexp.QuoteMeta("set status = $1, updated_at = $2")

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectExec(pattern).
			WithArgs(entity.StatementStatusOverdue, now, int64(10)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		err = r.MarkStatementOverdue(tx, 10, now)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_GetNextStatementForUpdate(t *testing.T) {
	loanID := int64(1)
	statementDate := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	pattern := regexp.QuoteMeta("statement_date > $2")

	t.Run("found returns the next statement id", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectQuery(pattern).
			WithArgs(loanID, statementDate).
			WillReturnRows(sqlmock.NewRows([]string{"statement_id"}).AddRow(int64(11)))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		got, found, err := r.GetNextStatementForUpdate(tx, loanID, statementDate)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, int64(11), got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found (last installment) returns found=false, no error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectQuery(pattern).
			WithArgs(loanID, statementDate).
			WillReturnRows(sqlmock.NewRows([]string{"statement_id"}))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		_, found, err := r.GetNextStatementForUpdate(tx, loanID, statementDate)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.False(t, found)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_UpdateCarryOver(t *testing.T) {
	now := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
	pattern := regexp.QuoteMeta("set carry_over_amount = $1")

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectExec(pattern).
			WithArgs(decimal.NewFromInt(110000), now, int64(11)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		err = r.UpdateCarryOver(tx, 11, decimal.NewFromInt(110000), now)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_GetPreviousStatementStatus(t *testing.T) {
	loanID := int64(1)
	statementDate := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	pattern := regexp.QuoteMeta("statement_date < $2")

	t.Run("found returns its status", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectQuery(pattern).
			WithArgs(loanID, statementDate).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(entity.StatementStatusOverdue))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		status, found, err := r.GetPreviousStatementStatus(tx, loanID, statementDate)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, entity.StatementStatusOverdue, status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found (first installment) returns found=false, no error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		tx := beginTx(t, db, mock)
		mock.ExpectQuery(pattern).
			WithArgs(loanID, statementDate).
			WillReturnRows(sqlmock.NewRows([]string{"status"}))
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		_, found, err := r.GetPreviousStatementStatus(tx, loanID, statementDate)
		require.NoError(t, tx.Commit())

		require.NoError(t, err)
		assert.False(t, found)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
