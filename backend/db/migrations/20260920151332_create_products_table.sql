-- +migrate Up
CREATE TABLE products (
    id            CHAR(36)     NOT NULL,
    version       INT          NOT NULL DEFAULT 1,
    name          VARCHAR(255) NOT NULL,
    description   TEXT         NULL,
    price_yen     INT          NOT NULL,
    stock         INT          NOT NULL DEFAULT 0,
    created_at    DATETIME(3)  NOT NULL,
    updated_at    DATETIME(3)  NOT NULL,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE products;
