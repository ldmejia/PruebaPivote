CREATE TABLE daily_usage (
    account_number VARCHAR(20) NOT NULL,
    usage_date DATE NOT NULL,
    total BIGINT NOT NULL,
    CONSTRAINT pk_daily_usage PRIMARY KEY (account_number, usage_date),
    CONSTRAINT fk_daily_usage_account FOREIGN KEY (account_number) REFERENCES accounts (number),
    CONSTRAINT ck_daily_usage_total CHECK (total >= 0)
);
