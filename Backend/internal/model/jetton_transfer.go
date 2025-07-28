package model

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type JettonTransfer struct {
	bun.BaseModel `bun:"table:jetton_transfers"`

	TxHash          string `bun:"tx_hash,pk,type:char(64),notnull" json:"tx_hash"`
	TxLT            uint64 `bun:"tx_lt,pk,type:bigint,notnull" json:"tx_lt"`
	SenderAddress   string `bun:"sender_address,type:char(48),notnull" json:"sender_address"`
	ReceiverAddress string `bun:"receiver_address,type:char(48),notnull" json:"receiver_address"`

	JettonName   string          `bun:"jetton_name,type:varchar(30),notnull" json:"jetton_name"`
	JettonAmount decimal.Decimal `bun:"jetton_amount,type:decimal,notnull" json:"jetton_amount"`
	TextComment  string          `bun:"text_comment,type:text,notnull" json:"text_comment"`
	IsApplied    bool            `bun:"is_applied,type:boolean,notnull" json:"is_applied"`

	CreatedAt time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`
}
