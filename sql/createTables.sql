-- Enum untuk jenis transaksi
CREATE TYPE type_transaction AS ENUM ('transfer', 'top_up');

-- Table accounts
CREATE TABLE public.accounts (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email       VARCHAR(255) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL,
    pin         VARCHAR(6) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table profiles
CREATE TABLE public.profiles (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    fullname    VARCHAR(255),
    phone       VARCHAR(255),
    img         VARCHAR(255),
    verified    BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_profiles_account FOREIGN KEY (id) REFERENCES accounts(id) 
);

-- Table wallets
CREATE TABLE public.wallets (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    balance     INTEGER DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_wallets_account FOREIGN KEY (id) REFERENCES accounts(id) 
);

-- Table internal_accounts
CREATE TABLE public.internal_accounts (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(255),
    img         VARCHAR(255),
    tax         INTEGER DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table transactions
CREATE TABLE public.transactions (
    id              INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type            type_transaction NOT NULL,
    amount          INTEGER NOT NULL,
    total           INTEGER NOT NULL,
    note            TEXT,
    id_sender       INTEGER,
    id_receiver     INTEGER,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP,
    -- Relasi ke wallets atau internal_accounts
    CONSTRAINT fk_tx_sender_wallet FOREIGN KEY (id_sender) REFERENCES wallets(id) ,
    CONSTRAINT fk_tx_receiver_wallet FOREIGN KEY (id_receiver) REFERENCES wallets(id) ,
    CONSTRAINT fk_tx_sender_internal FOREIGN KEY (id_sender) REFERENCES internal_accounts(id) ,
    CONSTRAINT fk_tx_receiver_internal FOREIGN KEY (id_receiver) REFERENCES internal_accounts(id) 
);
