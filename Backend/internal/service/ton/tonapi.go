package ton

import (
	"academy/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"go.uber.org/zap"
)

type tonapiJettonTransferPayload struct {
	SumType string `json:"SumType"`
	OpCode  int    `json:"OpCode"`
	Value   struct {
		Text string `json:"Text"`
	} `json:"Value"`
}

func (s *Service) updateJettonTransfersWithTONAPI(
	ctx context.Context,
	destAddr *address.Address,
	fromTxLT *uint64,
) (int, error) {

	transfers := make([]*model.JettonTransfer, 0)

	// Pagination.
	var beforeLt tonapi.OptInt64
	limit := 100

	// Rate limit.
	const tonapiRPS = 1
	lastReq := time.Now()

	// Count number of iterations to warn about any edge cases.
	loopCounter := 0

	for {
		waitTime := lastReq.Add(time.Second / tonapiRPS).Sub(lastReq)

		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			return 0, nil
		}

		jettonOps, err := s.tonapiClient.GetAccountJettonsHistory(ctx, tonapi.GetAccountJettonsHistoryParams{
			AccountID: destAddr.String(),
			BeforeLt:  beforeLt,
			Limit:     limit,
		})

		if err != nil {
			return 0, fmt.Errorf("s.tonapiClient.GetAccountJettonsHistory: %w", err)
		}

		for _, op := range jettonOps.GetOperations() {
			loopCounter++
			if 0 < loopCounter && loopCounter%1000 == 0 {
				s.logger.Warn("loop counter",
					zap.Int("iteration", loopCounter),
					zap.String("destAddr", destAddr.String()),
					zap.Int64("tx.LT", op.GetLt()),
				)
			}

			beforeLt = tonapi.NewOptInt64(op.GetLt())

			if fromTxLT != nil && beforeLt.Set && uint64(beforeLt.Value) <= *fromTxLT {
				break
			}

			if op.Operation != tonapi.JettonOperationOperationTransfer {
				continue
			}

			if dest, ok := op.GetDestination().Get(); !ok || dest.GetAddress() != destAddr.StringRaw() {
				s.logger.Warn("invalid destination",
					zap.String("dest_address", dest.GetAddress()),
					zap.String("expected dest_address", destAddr.StringRaw()),
					zap.String("tx_hash", op.GetTransactionHash()),
				)
				continue
			}

			transfer := s.processJettonTransferOperation(op)
			if transfer == nil {
				continue
			}

			transfers = append(transfers, transfer)
		}

		if len(jettonOps.GetOperations()) < limit {
			break
		}

		if fromTxLT != nil && beforeLt.Set && uint64(beforeLt.Value) <= *fromTxLT {
			break
		}
	}

	return s.applyJettonTransfers(ctx, transfers)
}

func (s *Service) processJettonTransferOperation(op tonapi.JettonOperation) *model.JettonTransfer {
	jettonAddress := op.GetJetton().Address
	txHash := op.GetTransactionHash()
	txLT := op.GetLt()

	jetton, ok := s.acceptedJettonsByHex[jettonAddress]
	if !ok {
		s.logger.Warn("invalid jetton",
			zap.String("tx_hash", txHash),
			zap.String("jetton_address", jettonAddress),
		)
		return nil
	}

	var sourceAddr *address.Address
	if addr, ok := op.Source.Get(); !ok {
		s.logger.Warn("invalid source_address",
			zap.String("tx_hash", txHash),
		)

		return nil
	} else {
		var err error
		sourceAddr, err = address.ParseRawAddr(addr.GetAddress())
		if err != nil {
			s.logger.Warn("invalid source_address",
				zap.String("tx_hash", txHash),
				zap.String("source_address", addr.GetAddress()),
				zap.Error(err),
			)
			return nil
		}
	}
	sourceAddr = s.verifyAddress(sourceAddr)

	var destAddr *address.Address
	if addr, ok := op.GetDestination().Get(); !ok {
		s.logger.Warn("invalid dest_address",
			zap.String("tx_hash", txHash),
		)

		return nil
	} else {
		var err error
		destAddr, err = address.ParseRawAddr(addr.GetAddress())
		if err != nil {
			s.logger.Warn("invalid dest_address",
				zap.String("tx_hash", txHash),
				zap.String("dest_address", addr.GetAddress()),
				zap.Error(err),
			)
			return nil
		}
	}
	destAddr = s.verifyAddress(destAddr)

	var payload tonapiJettonTransferPayload
	if err := json.Unmarshal(op.GetPayload(), &payload); err != nil {
		s.logger.Warn("invalid payload",
			zap.String("tx_hash", txHash),
		)
		return nil
	}

	if payload.SumType != "TextComment" || payload.OpCode != 0 {
		s.logger.Warn("invalid payload",
			zap.String("tx_hash", txHash),
			zap.String("payload.SumType", payload.SumType),
			zap.Int("payload.OpCode", payload.OpCode),
		)
		return nil
	}

	amount, err := decimal.NewFromString(op.Amount)
	if err != nil {
		s.logger.Warn("invalid amount",
			zap.String("tx_hash", txHash),
			zap.String("amount", op.Amount),
		)
		return nil
	}

	return &model.JettonTransfer{
		TxHash: txHash,
		TxLT:   uint64(txLT),

		SenderAddress:   sourceAddr.String(),
		ReceiverAddress: destAddr.String(),

		JettonName:   jetton.Name,
		JettonAmount: amount.Shift(-int32(jetton.Decimal)),
		TextComment:  payload.Value.Text,
		CreatedAt:    time.Unix(op.Utime, 0),
	}
}
