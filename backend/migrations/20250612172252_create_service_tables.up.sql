DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS popular_products;

CREATE TABLE sessions
(
    session_id  UUID PRIMARY KEY,
    context     TEXT,
    last_active TIMESTAMP DEFAULT NOW()
);

CREATE TABLE popular_products
(
    product_id  SERIAL PRIMARY KEY,
    name        VARCHAR(255),
    description TEXT,
    price       DECIMAL(10, 2),
    rating      DECIMAL(3, 2),
    category    VARCHAR(255),
    product_url VARCHAR(255),
    image_url   VARCHAR(255)
);