CREATE TABLE if not exists "users"
(
    id       UUID PRIMARY KEY,
    email    VARCHAR(256) NOT NULL UNIQUE,
    password VARCHAR(512) NOT NULL,
    role     VARCHAR(10)  NOT NULL
);

ALTER TABLE "users"
    ADD CONSTRAINT city_check
        CHECK (role IN ('employee', 'moderator'));

CREATE TABLE pickup_point
(
    id                UUID PRIMARY KEY,
    registration_date TIMESTAMP    NOT NULL,
    city              VARCHAR(255) NOT NULL
);

ALTER TABLE pickup_point
    ADD CONSTRAINT city_check
        CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань'));

CREATE TABLE receiving
(
    id                 UUID PRIMARY KEY,
    receiving_datetime TIMESTAMP   NOT NULL,
    pickup_point_id    UUID        NOT NULL,
    status             VARCHAR(50) NOT NULL,
    CONSTRAINT fk_pickup_point
        FOREIGN KEY (pickup_point_id) REFERENCES pickup_point (id)
);

ALTER TABLE receiving
    ADD CONSTRAINT receiving_status_check
        CHECK (status IN ('in_progress', 'close'));

CREATE TABLE goods
(
    id                UUID PRIMARY KEY,
    receiving_id      UUID        NOT NULL,
    accepted_datetime TIMESTAMP   NOT NULL,
    product_type      VARCHAR(50) NOT NULL,
    CONSTRAINT fk_receiving
        FOREIGN KEY (receiving_id) REFERENCES receiving (id)
);

ALTER TABLE goods
    ADD CONSTRAINT goods_type_check
        CHECK (product_type IN ('электроника', 'одежда', 'обувь'));