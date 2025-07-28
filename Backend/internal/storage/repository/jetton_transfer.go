package repository

import (
	"academy/internal/database/repository"
	"academy/internal/model"
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type JettonTransferRepository struct {
	repository.Generic[model.JettonTransfer, uuid.UUID]
}

func (r *JettonTransferRepository) WithTx(tx bun.Tx) *JettonTransferRepository {
	return &JettonTransferRepository{Generic: r.Generic.WithTx(tx)}
}

func NewJettonTransferRepository(
	genericRepository repository.Generic[model.JettonTransfer, uuid.UUID],
) *JettonTransferRepository {
	return &JettonTransferRepository{
		Generic: genericRepository,
	}
}

func (r *JettonTransferRepository) Create(
	ctx context.Context, transfer ...*model.JettonTransfer,
) error {

	_, err := r.DB.NewInsert().
		Model(&transfer).
		On(`CONFLICT ("tx_lt", "tx_hash") DO UPDATE`).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *JettonTransferRepository) GetLatest(
	ctx context.Context, receiverAddress string,
) (*model.JettonTransfer, error) {

	transfer := new(model.JettonTransfer)

	query := r.DB.NewSelect().
		Model(transfer).
		Where(`receiver_address = ?`, receiverAddress).
		OrderExpr(`tx_lt DESC`).
		Limit(1)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return transfer, nil
}
