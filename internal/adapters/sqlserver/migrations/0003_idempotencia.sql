CREATE TABLE idempotency_keys (
    owner_id VARCHAR(64) NOT NULL,
    idem_key VARCHAR(100) NOT NULL,
    request_hash CHAR(64) NOT NULL,
    transfer_id UNIQUEIDENTIFIER NOT NULL,
    created_at DATETIME2 NOT NULL CONSTRAINT df_idempotency_created DEFAULT SYSUTCDATETIME(),
    CONSTRAINT pk_idempotency_keys PRIMARY KEY (owner_id, idem_key),
    CONSTRAINT fk_idempotency_transfer FOREIGN KEY (transfer_id) REFERENCES transfers (id)
);
