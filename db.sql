CREATE TABLE loan (
    loan_id           bigint        PRIMARY KEY,
    borrower_id       bigint        NOT NULL,
    principal_amount  numeric(18,2) NOT NULL CHECK (principal_amount > 0),
    rate              numeric(9,6)  NOT NULL CHECK (rate >= 0),
    total_amount      numeric(18,2) NOT NULL CHECK (total_amount >= principal_amount),
    installment_count smallint      NOT NULL CHECK (installment_count > 0),
    installment_type  smallint      NOT NULL CHECK (installment_type > 0),
    currency          char(3)       NOT NULL DEFAULT 'IDR',
    start_date        date          NOT NULL,
    disbursed_at      timestamptz,
    metadata          jsonb,
    status            smallint      NOT NULL DEFAULT 1,
    created_at        timestamptz   NOT NULL DEFAULT now(),
    updated_at        timestamptz   
);

COMMENT ON COLUMN loan.installment_type	IS '1: daily, 2: weekly, 3: monthly';
COMMENT ON COLUMN loan.status      		IS '1 active, 2 closed';

CREATE INDEX idx_loan_borrower ON loan (borrower_id, status);

-----------------------------------------------------------------------------------------------------------------------------


CREATE TABLE statement (
    statement_id 		bigint        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    loan_id       		bigint        NOT NULL CHECK (loan_id > 0),
    installment_amount  numeric(18,2) NOT NULL CHECK (installment_amount > 0),
    carry_over_amount	numeric(18,2) NOT NULL DEFAULT 0 CHECK (carry_over_amount >= 0),
    paid_amount 		numeric(18,2) NOT NULL DEFAULT 0 CHECK (paid_amount >= 0),
    statement_date      date          NOT NULL,
    deadline        	date          NOT null CHECK (deadline >= statement_date),
    status 				smallint      NOT NULL DEFAULT 0,
    paid_at      		timestamptz,
    created_at        	timestamptz   NOT NULL DEFAULT now(),
    updated_at        	timestamptz,
    CONSTRAINT uk_statement_date UNIQUE (loan_id, statement_date)
);

COMMENT ON COLUMN statement.status IS '-1: overdue, 0: unpaid, 1: paid, 2: paid_late';

CREATE INDEX idx_statement_deadline_status ON statement (deadline, status);

-----------------------------------------------------------------------------------------------------------------------------

CREATE TABLE delinquency_hist					 (
    dh_id 			bigint        	GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    borrower_id     bigint        	NOT NULL CHECK (borrower_id > 0),
    loan_id     	bigint        	NOT NULL CHECK (loan_id > 0),
    statement_id	bigint 			NOT NULL CHECK (statement_id > 0),
    marked_at		timestamptz		NOT NULL DEFAULT now(),
    cleared_at 		timestamptz		CHECK (cleared_at IS NULL OR cleared_at >= marked_at),
    updated_at      timestamptz,
   CONSTRAINT uk_delinquency_statement UNIQUE (statement_id)
);

CREATE INDEX idx_delinquency_open ON delinquency_hist (borrower_id)
WHERE cleared_at IS NULL;

CREATE INDEX idx_delinquency_borrower ON delinquency_hist (borrower_id);

CREATE INDEX idx_delinquency_loan_open ON delinquency_hist (loan_id) WHERE cleared_at IS NULL;