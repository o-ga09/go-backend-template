-- +migrate Up
CREATE TABLE carts (
    id         CHAR(36)    NOT NULL,
    version    INT         NOT NULL DEFAULT 1,
    user_id    CHAR(36)    NOT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_carts_user_id (user_id),
    CONSTRAINT fk_carts_user FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE cart_items (
    id         CHAR(36)    NOT NULL,
    version    INT         NOT NULL DEFAULT 1,
    cart_id    CHAR(36)    NOT NULL,
    product_id CHAR(36)    NOT NULL,
    quantity   INT         NOT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_cart_items_cart_product (cart_id, product_id),
    CONSTRAINT fk_cart_items_cart FOREIGN KEY (cart_id) REFERENCES carts (id),
    CONSTRAINT fk_cart_items_product FOREIGN KEY (product_id) REFERENCES products (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE cart_items;
DROP TABLE carts;
