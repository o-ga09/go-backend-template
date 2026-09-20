-- +migrate Up
CREATE TABLE users (
    id         CHAR(36)     NOT NULL,
    version    INT          NOT NULL DEFAULT 1,
    google_sub VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL,
    name       VARCHAR(255) NOT NULL,
    created_at DATETIME(3)  NOT NULL,
    updated_at DATETIME(3)  NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_users_google_sub (google_sub),
    UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE users;
