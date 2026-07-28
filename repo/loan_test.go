package repo

import (
	"database/sql"
	"database/sql/driver"
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

func TestLoanRepo_InsertLoan(t *testing.T) {
	loan := entity.LoanDB{
		LoanID:           1,
		BorrowerID:       2,
		PrincipalAmount:  decimal.NewFromInt(5000000),
		Rate:             decimal.NewFromFloat(0.1),
		TotalAmount:      decimal.NewFromInt(5500000),
		InstallmentCount: 50,
		InstallmentType:  entity.InstallmentTypeWeekly,
		Currency:         entity.DefaultCurrency,
		StartDate:        time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
		DisbursedAt:      sql.NullTime{},
		Metadata:         nil,
		Status:           entity.LoanStatusActive,
		UpdatedAt:        sql.NullTime{},
	}

	queryPattern := regexp.QuoteMeta("insert into loan")

	args := []driver.Value{
		loan.LoanID,
		loan.BorrowerID,
		loan.PrincipalAmount,
		loan.Rate,
		loan.TotalAmount,
		loan.InstallmentCount,
		loan.InstallmentType,
		loan.Currency,
		loan.StartDate,
		loan.DisbursedAt,
		loan.Metadata,
		loan.Status,
		loan.UpdatedAt,
	}

	t.Run("success returns no error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectExec(queryPattern).
			WithArgs(args...).
			WillReturnResult(sqlmock.NewResult(0, 1))

		r := &loanRepo{db: db}
		err = r.InsertLoan(loan)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectExec(queryPattern).
			WithArgs(args...).
			WillReturnError(errors.New("connection refused"))

		r := &loanRepo{db: db}
		err = r.InsertLoan(loan)

		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_InsertLoanWithStatements(t *testing.T) {
	loan := entity.LoanDB{
		LoanID:           1,
		BorrowerID:       2,
		PrincipalAmount:  decimal.NewFromInt(5000000),
		Rate:             decimal.NewFromFloat(0.1),
		TotalAmount:      decimal.NewFromInt(5500000),
		InstallmentCount: 2,
		InstallmentType:  entity.InstallmentTypeWeekly,
		Currency:         entity.DefaultCurrency,
		StartDate:        time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
		DisbursedAt:      sql.NullTime{},
		Metadata:         nil,
		Status:           entity.LoanStatusActive,
		UpdatedAt:        sql.NullTime{},
	}

	loanArgs := []driver.Value{
		loan.LoanID, loan.BorrowerID, loan.PrincipalAmount, loan.Rate, loan.TotalAmount,
		loan.InstallmentCount, loan.InstallmentType, loan.Currency, loan.StartDate,
		loan.DisbursedAt, loan.Metadata, loan.Status, loan.UpdatedAt,
	}

	statements := []entity.StatementDB{
		{
			LoanID: loan.LoanID, InstallmentAmount: decimal.NewFromInt(2750000),
			StatementDate: time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC),
			Deadline:      time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
			Status:        entity.StatementStatusUnpaid,
		},
		{
			LoanID: loan.LoanID, InstallmentAmount: decimal.NewFromInt(2750000),
			StatementDate: time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC),
			Deadline:      time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC),
			Status:        entity.StatementStatusUnpaid,
		},
	}

	statementArgs := func(s entity.StatementDB) []driver.Value {
		return []driver.Value{
			s.LoanID, s.InstallmentAmount, s.CarryOverAmount, s.PaidAmount,
			s.StatementDate, s.Deadline, s.Status, s.PaidAt, s.UpdatedAt,
		}
	}

	loanQueryPattern := regexp.QuoteMeta("insert into loan")
	statementQueryPattern := regexp.QuoteMeta("insert into statement")

	t.Run("success commits loan and every statement in one transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec(loanQueryPattern).WithArgs(loanArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
		for _, s := range statements {
			mock.ExpectQuery(statementQueryPattern).
				WithArgs(statementArgs(s)...).
				WillReturnRows(sqlmock.NewRows([]string{"statement_id"}).AddRow(int64(1)))
		}
		mock.ExpectCommit()

		r := &loanRepo{db: db}
		err = r.InsertLoanWithStatements(loan, statements)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("loan insert failure rolls back before any statement is inserted", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec(loanQueryPattern).WithArgs(loanArgs...).WillReturnError(errors.New("connection refused"))
		mock.ExpectRollback()

		r := &loanRepo{db: db}
		err = r.InsertLoanWithStatements(loan, statements)

		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("statement insert failure rolls back the whole transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectBegin()
		mock.ExpectExec(loanQueryPattern).WithArgs(loanArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(statementQueryPattern).
			WithArgs(statementArgs(statements[0])...).
			WillReturnError(errors.New("connection refused"))
		mock.ExpectRollback()

		r := &loanRepo{db: db}
		err = r.InsertLoanWithStatements(loan, statements)

		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestLoanRepo_GetOutstandingAmount(t *testing.T) {
	loanID := int64(1)
	queryPattern := regexp.QuoteMeta("from loan l")

	t.Run("success returns outstanding amount and borrower id", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"borrower_id", "outstanding"}).AddRow(int64(2), "4500000")

		mock.ExpectQuery(queryPattern).
			WithArgs(loanID, entity.StatementStatusPaid, entity.StatementStatusPaidLate).
			WillReturnRows(rows)

		r := &loanRepo{db: db}
		got, borrowerID, err := r.GetOutstandingAmount(loanID)

		require.NoError(t, err)
		assert.True(t, decimal.NewFromInt(4500000).Equal(got))
		assert.Equal(t, int64(2), borrowerID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("loan not found returns ErrLoanNotFound", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(loanID, entity.StatementStatusPaid, entity.StatementStatusPaidLate).
			WillReturnRows(sqlmock.NewRows([]string{"borrower_id", "outstanding"}))

		r := &loanRepo{db: db}
		_, _, err = r.GetOutstandingAmount(loanID)

		assert.ErrorIs(t, err, entity.ErrLoanNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error is returned", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		mock.ExpectQuery(queryPattern).
			WithArgs(loanID, entity.StatementStatusPaid, entity.StatementStatusPaidLate).
			WillReturnError(errors.New("connection refused"))

		r := &loanRepo{db: db}
		_, _, err = r.GetOutstandingAmount(loanID)

		assert.ErrorContains(t, err, "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
