CREATE TYPE type_transaction AS ENUM ('transfer', 'top_up');

CREATE TABLE public.transactions (
    id                  INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type                type_transaction NOT NULL,
    amount              INTEGER NOT NULL,
    total               INTEGER NOT NULL,
    note                TEXT,
    id_sender           INTEGER,
    id_receiver         INTEGER,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_sender      TIMESTAMP,
    deleted_receiver    TIMESTAMP,
    -- Relasi ke wallets atau internal_accounts
    CONSTRAINT fk_tx_sender_wallet FOREIGN KEY (id_sender) REFERENCES wallets(id) ,
    CONSTRAINT fk_tx_receiver_wallet FOREIGN KEY (id_receiver) REFERENCES wallets(id) ,
    CONSTRAINT fk_tx_sender_internal FOREIGN KEY (id_sender) REFERENCES internal_accounts(id) ,
    CONSTRAINT fk_tx_receiver_internal FOREIGN KEY (id_receiver) REFERENCES internal_accounts(id) 
);
