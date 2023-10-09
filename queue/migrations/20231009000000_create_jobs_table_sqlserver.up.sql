CREATE TABLE jobs
(
    id             BIGINT IDENTITY (1,1) PRIMARY KEY,
    queue          VARCHAR(MAX) NOT NULL,
    job            VARCHAR(MAX) NOT NULL,
    arg            VARCHAR(MAX) NOT NULL,
    attempts       BIGINT       NOT NULL DEFAULT 0,
    max_tries      BIGINT                DEFAULT NULL,
    max_exceptions BIGINT                DEFAULT NULL,
    exception      VARCHAR(MAX)          DEFAULT NULL,
    backoff        BIGINT       NOT NULL DEFAULT 0,
    timeout        BIGINT                DEFAULT NULL,
    timeout_at     TIMESTAMP             DEFAULT NULL,
    reserved_at    TIMESTAMP             DEFAULT NULL,
    available_at   TIMESTAMP    NOT NULL,
    created_at     TIMESTAMP    NOT NULL,
    failed_at      TIMESTAMP             DEFAULT NULL
);
