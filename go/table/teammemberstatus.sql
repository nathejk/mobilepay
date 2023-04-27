CREATE TABLE IF NOT EXISTS teammemberstatus (
    memberId VARCHAR(99) NOT NULL,
    teamId VARCHAR(99) NOT NULL,
    teamType VARCHAR(99) NOT NULL,
    status VARCHAR(99) NOT NULL,
    PRIMARY KEY (memberId)
);
