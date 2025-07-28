CREATE TABLE IF NOT EXISTS jetton_transfers (
    "tx_lt" BIGINT NOT NULL,
    "tx_hash" CHAR(64) NOT NULL,
    "sender_address" CHAR(48) NOT NULL,
    "receiver_address" CHAR(48) NOT NULL,
    
    "jetton_name" VARCHAR(30) NOT NULL,
    "jetton_amount" DECIMAL NOT NULL,
    "text_comment" TEXT NOT NULL,
    "is_applied" BOOLEAN NOT NULL,

    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,

    PRIMARY KEY ("tx_lt", "tx_hash")
);
