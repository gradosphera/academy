package ton

import (
	"academy/internal/config"
	repo "academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/storage/repository"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tonkeeper/tonapi-go"
	"github.com/uptrace/bun"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"go.uber.org/zap"
)

type tonNetwork string

const (
	tonNetworkMainnet tonNetwork = "mainnet"
	tonNetworkTestnet tonNetwork = "testnet"
)

type Service struct {
	logger *zap.Logger

	jettonTransferRepository *repository.JettonTransferRepository
	paymentRepository        *repository.PaymentRepository
	transactionManager       *repo.TransactionManager

	conn                 *liteclient.ConnectionPool
	api                  ton.APIClientWrapped
	acceptedJettons      []*jettonInfo
	acceptedJettonsByHex map[string]*jettonInfo
	network              tonNetwork

	tonapiClient *tonapi.Client
}

type jettonInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Decimal int    `json:"decimal"`
}

func NewService(
	logger *zap.Logger,
	cfg *config.Config,

	jettonTransferRepository *repository.JettonTransferRepository,
	paymentRepository *repository.PaymentRepository,
	transactionManager *repo.TransactionManager,
) (*Service, error) {

	conn := liteclient.NewConnectionPool()

	err := conn.AddConnectionsFromConfigUrl(context.Background(), cfg.TON.ConfigURL)
	if err != nil {
		return nil, fmt.Errorf("conn.AddConnectionsFromConfigUrl: %w", err)
	}
	api := ton.NewAPIClient(conn).WithRetry()

	acceptedJettons := make([]*jettonInfo, 0)
	acceptedJettonsByHex := make(map[string]*jettonInfo, 0)
	if len(cfg.TON.AcceptedJettons) != 0 {
		err := json.Unmarshal([]byte(cfg.TON.AcceptedJettons), &acceptedJettons)
		if err != nil {
			return nil, fmt.Errorf("json.Unmarshal: %w", err)
		}
		for _, acceptedJetton := range acceptedJettons {
			addr, err := address.ParseAddr(acceptedJetton.Address)
			if err != nil {
				return nil, fmt.Errorf("failed to parse jetton address: %w", err)
			}

			acceptedJettonsByHex[addr.StringRaw()] = acceptedJetton
		}
	}

	var network tonNetwork
	tonapiURL := tonapi.TonApiURL
	switch tonNetwork(cfg.TON.Network) {
	case tonNetworkMainnet:
		network = tonNetworkMainnet
	case tonNetworkTestnet:
		network = tonNetworkTestnet
		tonapiURL = tonapi.TestnetTonApiURL
	default:
		return nil, fmt.Errorf("unsupported %q TON network", cfg.TON.Network)
	}

	tonapiClient, err := tonapi.NewClient(tonapiURL, &tonapi.Security{
		Token: cfg.TON.TONAPIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("tonapi.NewClient: %w", err)
	}

	return &Service{
		logger:                   logger,
		jettonTransferRepository: jettonTransferRepository,
		paymentRepository:        paymentRepository,
		transactionManager:       transactionManager,

		conn:                 conn,
		api:                  api,
		acceptedJettons:      acceptedJettons,
		acceptedJettonsByHex: acceptedJettonsByHex,
		network:              network,

		tonapiClient: tonapiClient,
	}, nil
}

func (s *Service) UpdateJettonTransfers(ctx context.Context, destAddress string) (int, error) {
	ctx = s.conn.StickyContext(ctx)

	destAddr, err := address.ParseAddr(destAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to load Addr: %w", err)
	}
	destAddr = s.verifyAddress(destAddr)

	lastTransfer, err := s.jettonTransferRepository.GetLatest(ctx, destAddr.String())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("s.jettonTransferRepository.GetLatest: %w", err)
	}

	var fromTxLT *uint64
	if !errors.Is(err, sql.ErrNoRows) {
		fromTxLT = &lastTransfer.TxLT
	}

	latestBlock, err := s.api.CurrentMasterchainInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("s.api.CurrentMasterchainInfo: %w", err)
	}

	account, err := s.api.GetAccount(ctx, latestBlock, destAddr)
	if err != nil {
		return 0, fmt.Errorf("s.api.GetAccount: %w", err)
	}

	if !account.IsActive {
		return s.updateJettonTransfersWithTONAPI(ctx, destAddr, fromTxLT)
	}

	lastTxLT := account.LastTxLT
	lastTxHash := account.LastTxHash

	if fromTxLT != nil && lastTxLT == *fromTxLT {
		return 0, nil
	}

	jettonWallets := make(map[string]*jettonInfo)
	for _, info := range s.acceptedJettons {
		masterContractAddr, err := address.ParseAddr(info.Address)
		if err != nil {
			s.logger.Warn("failed to parse accepted jetton", zap.Error(err), zap.String("name", info.Name))
			continue
		}
		masterContractAddr = s.verifyAddress(masterContractAddr)

		jettonMaster := jetton.NewJettonMasterClient(s.api, masterContractAddr)
		jettonWallet, err := jettonMaster.GetJettonWallet(ctx, destAddr)
		if err != nil {
			return 0, fmt.Errorf("%v jetton: jettonMaster.GetJettonWallet: %w", info.Name, err)
		}
		jettonWalletAddress := s.verifyAddress(jettonWallet.Address())

		jettonWallets[jettonWalletAddress.String()] = info
	}

	// Count number of iterations to warn about any edge cases.
	loopCounter := 0

	limit := uint32(10)

	// Because ListTransactions will include lastTxHash in the result we need
	// to skip it on the next iterations.
	includeFirstTx := true

	transfers := make([]*model.JettonTransfer, 0)

	for {
		txs, err := s.api.ListTransactions(ctx, destAddr, limit, lastTxLT, lastTxHash)
		if err != nil && !errors.Is(err, ton.ErrNoTransactionsWereFound) {
			return 0, fmt.Errorf("s.api.ListTransactions(%v,%v,%v,%v): %w",
				destAddr.String(), limit, lastTxLT, hex.EncodeToString(lastTxHash), err)
		}

		if errors.Is(err, ton.ErrNoTransactionsWereFound) || len(txs) == 0 {
			break
		}

		// Make newest first.
		slices.Reverse(txs)

		for txIdx, tx := range txs {
			loopCounter++
			if 0 < loopCounter && loopCounter%1000 == 0 {
				s.logger.Warn("loop counter",
					zap.Int("iteration", loopCounter),
					zap.String("destAddr", destAddr.String()),
					zap.Uint64("tx.LT", tx.LT),
				)
			}

			lastTxLT = tx.LT
			lastTxHash = tx.Hash

			if fromTxLT != nil && lastTxLT <= *fromTxLT {
				break
			}

			if txIdx == 0 {
				if includeFirstTx {
					includeFirstTx = false
				} else {
					continue
				}
			}

			transfer := s.processJettonTransferNotification(tx, destAddr, jettonWallets)
			if transfer == nil {
				continue
			}

			transfers = append(transfers, transfer)
		}

		if len(txs) < int(limit) {
			break
		}

		if fromTxLT != nil && lastTxLT <= *fromTxLT {
			break
		}
	}

	return s.applyJettonTransfers(ctx, transfers)
}

func (s *Service) applyJettonTransfers(ctx context.Context, transfers []*model.JettonTransfer) (int, error) {
	err := s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		for _, transfer := range transfers {
			paymentID, err := uuid.Parse(transfer.TextComment)
			if err != nil {
				s.logger.Warn("message comment don't include uuid",
					zap.Error(err),
					zap.String("comment", transfer.TextComment),
				)
				continue
			}
			payment, err := s.paymentRepository.WithTx(tx).GetByID(ctx, paymentID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("failed to get payment by id: %w", err)
			}

			if errors.Is(err, sql.ErrNoRows) {
				s.logger.Warn("message comment don't include real uuid",
					zap.Error(err),
					zap.String("comment", transfer.TextComment),
				)
				continue
			}

			if transfer.JettonAmount.LessThan(payment.AmountBLG) {
				s.logger.Warn("paid jetton amount is less than required payment amount",
					zap.Error(err),
					zap.String("required amount", payment.AmountBLG.String()),
					zap.String("actual amount", transfer.JettonAmount.String()),
				)
				continue
			}

			payment.Status = model.PaymentStatusCompleted
			payment.UpdatedAt = time.Now()

			transfer.IsApplied = true

			err = s.paymentRepository.WithTx(tx).Update(ctx, payment)
			if err != nil {
				return fmt.Errorf("failed to update payment: %w", err)
			}
		}

		if len(transfers) != 0 {
			err := s.jettonTransferRepository.WithTx(tx).Create(ctx, transfers...)
			if err != nil {
				return fmt.Errorf("failed to save new jetton transfers: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return len(transfers), nil
}

func (s *Service) processJettonTransferNotification(
	tx *tlb.Transaction,
	expectedDestAddr *address.Address,
	jettonWallets map[string]*jettonInfo,
) *model.JettonTransfer {

	txHash := hex.EncodeToString(tx.Hash)

	switch txDescription := tx.Description.(type) {
	default:
		s.logger.Warn("transaction description is not ordinary", zap.String("tx_hash", txHash))
		return nil
	case tlb.TransactionDescriptionOrdinary:
		if txDescription.BouncePhase != nil {
			s.logger.Warn("transaction include bounce phase", zap.String("tx_hash", txHash))
			return nil
		}
		switch computePhase := txDescription.ComputePhase.Phase.(type) {
		default:
			s.logger.Warn("compute phase is not VM", zap.String("tx_hash", txHash))
			return nil
		case tlb.ComputePhaseSkipped:
			// For transfer_notification it is expected to skip compute phase.
			if computePhase.Reason.Type != tlb.ComputeSkipReasonNoGas {
				s.logger.Warn("compute phase failed",
					zap.String("tx_hash", txHash),
					zap.String("reason", string(computePhase.Reason.Type)),
				)
				return nil
			}
		case tlb.ComputePhaseVM:
			if computePhase.Details.ExitCode != 0 {
				s.logger.Warn("compute phase exit code is not 0",
					zap.String("tx_hash", txHash),
					zap.Int32("exit code", computePhase.Details.ExitCode),
				)
				return nil
			}
		}

		if txDescription.ActionPhase != nil && txDescription.ActionPhase.ResultCode != 0 {
			s.logger.Warn("action phase exit code is not 0",
				zap.String("tx_hash", txHash),
				zap.Int32("exit code", txDescription.ActionPhase.ResultCode),
			)
			return nil
		}
	}

	if tx.IO.In == nil || tx.IO.In.MsgType != tlb.MsgTypeInternal {
		return nil
	}

	msg := tx.IO.In.Msg.(*tlb.InternalMessage)
	if msg == nil {
		s.logger.Warn("unexpectedly message is not InternalMessage", zap.String("tx_hash", txHash))
		return nil
	}

	if msg.Bounce {
		s.logger.Warn("the in-message bounced", zap.String("tx_hash", txHash))
		return nil
	}

	createdAt := time.Unix(int64(msg.CreatedAt), 0)

	destAddr := s.verifyAddress(msg.DestAddr())
	jettonWalletAddress := s.verifyAddress(msg.SenderAddr())

	if destAddr.String() != expectedDestAddr.String() {
		return nil
	}

	var jettonInfo *jettonInfo
	if info, ok := jettonWallets[jettonWalletAddress.String()]; ok {
		jettonInfo = info
	} else {
		return nil
	}

	if msg.Payload() == nil || msg.Payload() == cell.BeginCell().EndCell() {
		return nil
	}

	msgBodySlice := msg.Payload().BeginParse()

	opCode, err := msgBodySlice.LoadUInt(32)
	if err != nil {
		s.logger.Warn("failed to load UInt32", zap.Error(err), zap.String("tx_hash", txHash))
		return nil
	}

	const transferNotificationOpCode = 0x7362d09c

	if opCode != transferNotificationOpCode {
		return nil
	}

	_, err = msgBodySlice.LoadUInt(64)
	if err != nil {
		s.logger.Warn("failed to load UInt64", zap.Error(err), zap.String("tx_hash", txHash))
		return nil
	}

	amount, err := msgBodySlice.LoadBigCoins()
	if err != nil {
		s.logger.Warn("failed to load BigCoins", zap.Error(err), zap.String("tx_hash", txHash))
		return nil
	}

	senderAddr, err := msgBodySlice.LoadAddr()
	if err != nil {
		s.logger.Warn("failed to load Addr", zap.Error(err), zap.String("tx_hash", txHash))
		return nil
	}
	senderAddr = s.verifyAddress(senderAddr)

	payload, err := msgBodySlice.LoadMaybeRef()
	if err != nil {
		s.logger.Warn("failed to load MaybeRef", zap.Error(err), zap.String("tx_hash", txHash))
		return nil
	}

	payloadOpCode, err := payload.LoadUInt(32)
	if err != nil {
		s.logger.Warn("failed to load UInt32", zap.Error(err), zap.String("tx_hash", txHash))
		return nil
	}

	if payloadOpCode != 0 {
		s.logger.Warn("no text comment in transfer_notification", zap.String("tx_hash", txHash))
		return nil
	}

	comment, err := payload.LoadStringSnake()
	if err != nil {
		s.logger.Warn("failed to load StringSnake", zap.Error(err), zap.String("tx_hash", txHash))
		return nil
	}

	return &model.JettonTransfer{
		TxHash:          txHash,
		TxLT:            tx.LT,
		SenderAddress:   senderAddr.String(),
		ReceiverAddress: destAddr.String(),

		JettonName:   jettonInfo.Name,
		JettonAmount: decimal.NewFromBigInt(amount, -int32(jettonInfo.Decimal)),
		TextComment:  comment,

		CreatedAt: createdAt,
	}
}

func (s *Service) verifyAddress(addr *address.Address) *address.Address {
	addr = addr.Bounce(true)

	if s.network == tonNetworkTestnet {
		addr.SetTestnetOnly(true)
	} else {
		addr.SetTestnetOnly(false)
	}

	return addr
}
