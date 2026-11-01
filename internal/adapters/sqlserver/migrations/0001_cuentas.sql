CREATE TABLE accounts (
    id BIGINT IDENTITY(1,1) NOT NULL,
    number VARCHAR(20) NOT NULL,
    owner_id VARCHAR(64) NOT NULL,
    currency CHAR(3) NOT NULL,
    balance BIGINT NOT NULL CONSTRAINT df_accounts_balance DEFAULT 0,
    blocked BIT NOT NULL CONSTRAINT df_accounts_blocked DEFAULT 0,
    created_at DATETIME2 NOT NULL CONSTRAINT df_accounts_created DEFAULT SYSUTCDATETIME(),
    updated_at DATETIME2 NOT NULL CONSTRAINT df_accounts_updated DEFAULT SYSUTCDATETIME(),
    CONSTRAINT pk_accounts PRIMARY KEY (id),
    CONSTRAINT uq_accounts_number UNIQUE (number),
    CONSTRAINT ck_accounts_currency CHECK (currency IN ('GTQ', 'USD')),
    CONSTRAINT ck_accounts_balance CHECK (balance >= 0)
);
GO

CREATE INDEX ix_accounts_owner ON accounts (owner_id);
