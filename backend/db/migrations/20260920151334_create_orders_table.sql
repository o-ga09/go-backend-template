-- +migrate Up
CREATE TABLE orders (
    id              CHAR(36)     NOT NULL,
    version         INT          NOT NULL DEFAULT 1,
    user_id         CHAR(36)     NOT NULL,
    status          VARCHAR(32)  NOT NULL DEFAULT 'pending',
    total_price_yen INT          NOT NULL,
    created_at      DATETIME(3)  NOT NULL,
    updated_at      DATETIME(3)  NOT NULL,
    PRIMARY KEY (id),
    KEY idx_orders_user_id (user_id),
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE order_items (
    id             CHAR(36)    NOT NULL,
    version        INT         NOT NULL DEFAULT 1,
    order_id       CHAR(36)    NOT NULL,
    product_id     CHAR(36)    NOT NULL,
    quantity       INT         NOT NULL,
    unit_price_yen INT         NOT NULL,
    created_at     DATETIME(3) NOT NULL,
    updated_at     DATETIME(3) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_order_items_order_id (order_id),
    CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders (id),
    CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES products (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE order_items;
DROP TABLE orders;
