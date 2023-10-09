CREATE TABLE jobs
(
    id             BIGINT AUTO_INCREMENT PRIMARY KEY,
    queue          TEXT      NOT NULL,
    job            TEXT      NOT NULL,
    arg            JSON      NOT NULL,
    attempts       BIGINT    NOT NULL DEFAULT 0,
    max_tries      BIGINT             DEFAULT NULL,
    max_exceptions BIGINT             DEFAULT NULL,
    exception      TEXT               DEFAULT NULL,
    backoff        BIGINT    NOT NULL DEFAULT 0,
    timeout        BIGINT             DEFAULT NULL,
    timeout_at     TIMESTAMP          DEFAULT NULL,
    reserved_at    TIMESTAMP          DEFAULT NULL,
    available_at   TIMESTAMP NOT NULL,
    created_at     TIMESTAMP NOT NULL,
    failed_at      TIMESTAMP          DEFAULT NULL
);
