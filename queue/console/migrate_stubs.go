package console

type MysqlStubs struct{}

// Jobs Create jobs table migration content.
func (receiver MysqlStubs) Jobs() string {
	return `CREATE TABLE jobs
(
    id             BIGINT AUTO_INCREMENT PRIMARY KEY,
    queue          VARCHAR(255) NOT NULL,
    job            VARCHAR(255) NOT NULL,
    payload        JSON         NOT NULL,
    attempts       BIGINT       NOT NULL DEFAULT 0,
    max_tries      BIGINT                DEFAULT NULL,
    max_exceptions BIGINT                DEFAULT NULL,
    backoff        BIGINT       NOT NULL DEFAULT 0,
    timeout        BIGINT                DEFAULT NULL,
    timeout_at     TIMESTAMP             DEFAULT NULL,
    reserved_at    TIMESTAMP             DEFAULT NULL,
    available_at   TIMESTAMP    NOT NULL,
    created_at     TIMESTAMP    NOT NULL
);

CREATE TABLE failed_jobs
(
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    queue      VARCHAR(255) NOT NULL,
    job        VARCHAR(255) NOT NULL,
    payload    JSON         NOT NULL,
    exception  TEXT         NOT NULL,
    failed_at  TIMESTAMP    NOT NULL
);

CREATE INDEX idx_jobs_index ON jobs (queue, job, reserved_at, available_at);
CREATE INDEX idx_failed_jobs_failed_at ON failed_jobs (queue, job, failed_at);
`
}

type PostgresqlStubs struct{}

// Jobs Create jobs table migration content.
func (receiver PostgresqlStubs) Jobs() string {
	return `CREATE TABLE jobs
(
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    queue          TEXT      NOT NULL,
    job            TEXT      NOT NULL,
    payload        JSONB     NOT NULL,
    attempts       BIGINT    NOT NULL DEFAULT 0,
    max_tries      BIGINT             DEFAULT NULL,
    max_exceptions BIGINT             DEFAULT NULL,
    backoff        BIGINT    NOT NULL DEFAULT 0,
    timeout        BIGINT             DEFAULT NULL,
    timeout_at     TIMESTAMP          DEFAULT NULL,
    reserved_at    TIMESTAMP          DEFAULT NULL,
    available_at   TIMESTAMP NOT NULL,
    created_at     TIMESTAMP NOT NULL
);

CREATE TABLE failed_jobs
(
    id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    queue     TEXT      NOT NULL,
    job       TEXT      NOT NULL,
    payload   JSONB     NOT NULL,
    exception TEXT      NOT NULL,
    failed_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_jobs_index ON jobs (queue, job, reserved_at, available_at);
CREATE INDEX idx_failed_jobs_failed_at ON failed_jobs (queue, job, failed_at);
`
}

type SqliteStubs struct{}

// Jobs Create jobs table migration content.
func (receiver SqliteStubs) Jobs() string {
	return `CREATE TABLE jobs
(
    id             BIGINT AUTO_INCREMENT PRIMARY KEY,
    queue          TEXT      NOT NULL,
    job            TEXT      NOT NULL,
    payload        JSON      NOT NULL,
    attempts       BIGINT    NOT NULL DEFAULT 0,
    max_tries      BIGINT             DEFAULT NULL,
    max_exceptions BIGINT             DEFAULT NULL,
    backoff        BIGINT    NOT NULL DEFAULT 0,
    timeout        BIGINT             DEFAULT NULL,
    timeout_at     TIMESTAMP          DEFAULT NULL,
    reserved_at    TIMESTAMP          DEFAULT NULL,
    available_at   TIMESTAMP NOT NULL,
    created_at     TIMESTAMP NOT NULL
);

CREATE TABLE failed_jobs
(
    id        BIGINT AUTO_INCREMENT PRIMARY KEY,
    queue     TEXT      NOT NULL,
    job       TEXT      NOT NULL,
    payload   JSON      NOT NULL,
    exception TEXT      NOT NULL,
    failed_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_jobs_index ON jobs (queue, job, reserved_at, available_at);
CREATE INDEX idx_failed_jobs_failed_at ON failed_jobs (queue, job, failed_at);
`
}

type SqlserverStubs struct{}

// Jobs Create jobs table migration content.
func (receiver SqlserverStubs) Jobs() string {
	return `CREATE TABLE jobs
(
    id             BIGINT IDENTITY (1,1) PRIMARY KEY,
    queue          VARCHAR(MAX) NOT NULL,
    job            VARCHAR(MAX) NOT NULL,
    payload        VARCHAR(MAX) NOT NULL,
    attempts       BIGINT       NOT NULL DEFAULT 0,
    max_tries      BIGINT                DEFAULT NULL,
    max_exceptions BIGINT                DEFAULT NULL,
    backoff        BIGINT       NOT NULL DEFAULT 0,
    timeout        BIGINT                DEFAULT NULL,
    timeout_at     TIMESTAMP             DEFAULT NULL,
    reserved_at    TIMESTAMP             DEFAULT NULL,
    available_at   TIMESTAMP    NOT NULL,
    created_at     TIMESTAMP    NOT NULL
);

CREATE TABLE failed_jobs
(
    id        BIGINT IDENTITY (1,1) PRIMARY KEY,
    queue     VARCHAR(MAX) NOT NULL,
    job       VARCHAR(MAX) NOT NULL,
    payload   VARCHAR(MAX) NOT NULL,
    exception VARCHAR(MAX) NOT NULL,
    failed_at TIMESTAMP    NOT NULL
);

CREATE INDEX idx_jobs_index ON jobs (queue, job, reserved_at, available_at);
CREATE INDEX idx_failed_jobs_failed_at ON failed_jobs (queue, job, failed_at);
`
}
