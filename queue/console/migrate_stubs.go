package console

type MysqlStubs struct{}

// FailedJobsUp Create jobs table migration content.
func (receiver MysqlStubs) FailedJobsUp() string {
	return `CREATE TABLE failed_jobs
(
    id         BIGINT AUTO_INCREMENT PRIMARY KEY,
    queue      VARCHAR(255) NOT NULL,
    job        VARCHAR(255) NOT NULL,
    payload    JSON         NOT NULL,
    exception  TEXT         NOT NULL,
    failed_at  TIMESTAMP    NOT NULL
);

CREATE INDEX idx_failed_jobs_failed_at ON failed_jobs (queue, job, failed_at);
`
}

// FailedJobsDown Drop jobs table migration content.
func (receiver MysqlStubs) FailedJobsDown() string {
	return `DROP TABLE IF EXISTS failed_jobs;
`
}

type PostgresqlStubs struct{}

// FailedJobsUp Create jobs table migration content.
func (receiver PostgresqlStubs) FailedJobsUp() string {
	return `CREATE TABLE failed_jobs
(
    id        BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    queue     TEXT      NOT NULL,
    job       TEXT      NOT NULL,
    payload   JSONB     NOT NULL,
    exception TEXT      NOT NULL,
    failed_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_failed_jobs_failed_at ON failed_jobs (queue, job, failed_at);
`
}

// FailedJobsDown Drop jobs table migration content.
func (receiver PostgresqlStubs) FailedJobsDown() string {
	return `DROP TABLE IF EXISTS failed_jobs;
`
}

type SqliteStubs struct{}

// FailedJobsUp Create jobs table migration content.
func (receiver SqliteStubs) FailedJobsUp() string {
	return `CREATE TABLE failed_jobs
(
    id        BIGINT AUTO_INCREMENT PRIMARY KEY,
    queue     TEXT      NOT NULL,
    job       TEXT      NOT NULL,
    payload   JSON      NOT NULL,
    exception TEXT      NOT NULL,
    failed_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_failed_jobs_failed_at ON failed_jobs (queue, job, failed_at);
`
}

// FailedJobsDown Drop jobs table migration content.
func (receiver SqliteStubs) FailedJobsDown() string {
	return `DROP TABLE IF EXISTS failed_jobs;
`
}

type SqlserverStubs struct{}

// FailedJobsUp Create jobs table migration content.
func (receiver SqlserverStubs) FailedJobsUp() string {
	return `CREATE TABLE failed_jobs
(
    id        BIGINT IDENTITY (1,1) PRIMARY KEY,
    queue     VARCHAR(MAX) NOT NULL,
    job       VARCHAR(MAX) NOT NULL,
    payload   VARCHAR(MAX) NOT NULL,
    exception VARCHAR(MAX) NOT NULL,
    failed_at TIMESTAMP    NOT NULL
);

CREATE INDEX idx_failed_jobs_failed_at ON failed_jobs (queue, job, failed_at);
`
}

// FailedJobsDown Drop jobs table migration content.
func (receiver SqlserverStubs) FailedJobsDown() string {
	return `DROP TABLE IF EXISTS failed_jobs;
`
}
