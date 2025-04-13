CREATE TABLE cities(
    id SERIAL,
    name VARCHAR(16) UNIQUE NOT NULL,

    PRIMARY KEY (id)
);

CREATE TABLE points(
    id UUID DEFAULT gen_random_uuid() NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    city_id INTEGER NOT NULL REFERENCES cities(id),

    PRIMARY KEY (id)
);

CREATE TYPE reception_status AS ENUM(
    'in_progress',
    'close'
);

CREATE TABLE receptions(
    id UUID DEFAULT gen_random_uuid() NOT NULL,
    point_id UUID NOT NULL REFERENCES points(id),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    status reception_status DEFAULT 'in_progress' NOT NULL,

    PRIMARY KEY (id)
);

CREATE TYPE product_type AS ENUM(
    'электроника',
    'одежда',
    'обувь'
);

CREATE TABLE products(
    id UUID DEFAULT gen_random_uuid() NOT NULL,
    reception_id UUID NOT NULL REFERENCES receptions(id),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    type product_type NOT NULL,

    PRIMARY KEY (id)
);

CREATE TYPE user_role AS ENUM(
    'employee',
    'moderator'
);

CREATE TABLE users(
    id UUID DEFAULT gen_random_uuid() NOT NULL,
    email VARCHAR(256) UNIQUE NOT NULL,
    password CHAR(60) NOT NULL,
    role user_role NOT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_receptions_created_at ON receptions(created_at);
CREATE UNIQUE INDEX idx_receptions_point_id_in_progress ON receptions(point_id) WHERE status = 'in_progress';
CREATE INDEX idx_products_reception_id_created_at_desc ON products(reception_id, created_at DESC);

INSERT INTO cities(name) VALUES
                       ('Москва'),
                       ('Санкт-Петербург'),
                       ('Казань');