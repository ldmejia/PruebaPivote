CREATE TABLE transfers (
    id UNIQUEIDENTIFIER NOT NULL,
    from_account VARCHAR(20) NOT NULL,
    to_account VARCHAR(20) NOT NULL,
    amount BIGINT NOT NULL,
    currency CHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL,
    reason VARCHAR(50) NULL,
    created_at DATETIME2 NOT NULL CONSTRAINT df_transfers_created DEFAULT SYSUTCDATETIME(),
    CONSTRAINT pk_transfers PRIMARY KEY (id),
    CONSTRAINT ck_transfers_amount CHECK (amount > 0),
    CONSTRAINT ck_transfers_status CHECK (status IN ('completed', 'rejected'))
);
GO

CREATE INDEX ix_transfers_from ON transfers (from_account, created_at);
GO

CREATE TABLE movements (
    id BIGINT IDENTITY(1,1) NOT NULL,
    account_number VARCHAR(20) NOT NULL,
    transfer_id UNIQUEIDENTIFIER NULL,
    type VARCHAR(10) NOT NULL,
    amount BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    created_at DATETIME2 NOT NULL CONSTRAINT df_movements_created DEFAULT SYSUTCDATETIME(),
    CONSTRAINT pk_movements PRIMARY KEY (id),
    CONSTRAINT fk_movements_account FOREIGN KEY (account_number) REFERENCES accounts (number),
    CONSTRAINT fk_movements_transfer FOREIGN KEY (transfer_id) REFERENCES transfers (id),
    CONSTRAINT ck_movements_type CHECK (type IN ('APERTURA', 'DEBITO', 'CREDITO')),
    CONSTRAINT ck_movements_amount CHECK (amount > 0)
);
GO

CREATE INDEX ix_movements_account_date ON movements (account_number, created_at, id);
GO

CREATE TRIGGER tr_movements_inmutable ON movements
INSTEAD OF UPDATE, DELETE
AS
BEGIN
    THROW 50001, 'movements es de solo insercion, no se puede modificar ni borrar', 1;
END
