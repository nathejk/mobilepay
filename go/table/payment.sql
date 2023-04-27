CREATE TABLE IF NOT EXISTS payment (
    ts VARCHAR(99) NOT NULL,
    amount DECIMAL(7, 2),
    currency VARCHAR(9),
    teamId VARCHAR(99),
    memberId VARCHAR(99),
    PRIMARY KEY (ts, teamId, memberId)
);
