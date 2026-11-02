CREATE TABLE outbox (
    id BIGINT IDENTITY(1,1) NOT NULL,
    event_id UNIQUEIDENTIFIER NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    payload NVARCHAR(MAX) NOT NULL,
    created_at DATETIME2 NOT NULL CONSTRAINT df_outbox_created DEFAULT SYSUTCDATETIME(),
    published_at DATETIME2 NULL,
    CONSTRAINT pk_outbox PRIMARY KEY (id),
    CONSTRAINT uq_outbox_event UNIQUE (event_id)
);
GO

CREATE INDEX ix_outbox_pending ON outbox (id) WHERE published_at IS NULL;
