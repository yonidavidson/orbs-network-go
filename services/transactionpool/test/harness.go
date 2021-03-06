package test

import (
	"context"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/orbs-network-go/config"
	"github.com/orbs-network/orbs-network-go/crypto/digest"
	"github.com/orbs-network/orbs-network-go/crypto/signature"
	"github.com/orbs-network/orbs-network-go/instrumentation/log"
	"github.com/orbs-network/orbs-network-go/services/transactionpool"
	"github.com/orbs-network/orbs-network-go/test/crypto/keys"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/gossipmessages"
	"github.com/orbs-network/orbs-spec/types/go/services"
	"github.com/orbs-network/orbs-spec/types/go/services/gossiptopics"
	"github.com/orbs-network/orbs-spec/types/go/services/handlers"
	"time"
)

type harness struct {
	txpool             services.TransactionPool
	gossip             *gossiptopics.MockTransactionRelay
	vm                 *services.MockVirtualMachine
	trh                *handlers.MockTransactionResultsHandler
	lastBlockHeight    primitives.BlockHeight
	lastBlockTimestamp primitives.TimestampNano
}

var thisNodeKeyPair = keys.Ed25519KeyPairForTests(8)
var otherNodeKeyPair = keys.Ed25519KeyPairForTests(9)
var transactionExpirationWindow = 30 * time.Minute

func (h *harness) expectTransactionToBeForwarded(tx *protocol.SignedTransaction, sig primitives.Ed25519Sig) {

	h.gossip.When("BroadcastForwardedTransactions", &gossiptopics.ForwardedTransactionsInput{
		Message: &gossipmessages.ForwardedTransactionsMessage{
			Sender: (&gossipmessages.SenderSignatureBuilder{
				SenderPublicKey: thisNodeKeyPair.PublicKey(),
				Signature:       sig,
			}).Build(),
			SignedTransactions: transactionpool.Transactions{tx},
		},
	}).Return(&gossiptopics.EmptyOutput{}, nil).Times(1)
}

func (h *harness) expectNoTransactionsToBeForwarded() {
	h.gossip.Never("BroadcastForwardedTransactions", mock.Any)
}

func (h *harness) ignoringForwardMessages() {
	h.gossip.When("BroadcastForwardedTransactions", mock.Any).Return(&gossiptopics.EmptyOutput{}, nil).AtLeast(0)
}

func (h *harness) addNewTransaction(tx *protocol.SignedTransaction) (*services.AddNewTransactionOutput, error) {
	out, err := h.txpool.AddNewTransaction(&services.AddNewTransactionInput{
		SignedTransaction: tx,
	})

	return out, err
}

func (h *harness) addTransactions(txs ...*protocol.SignedTransaction) {
	for _, tx := range txs {
		h.addNewTransaction(tx)
	}
}

func (h *harness) reportTransactionsAsCommitted(transactions ...*protocol.SignedTransaction) (*services.CommitTransactionReceiptsOutput, error) {
	return h.txpool.CommitTransactionReceipts(&services.CommitTransactionReceiptsInput{
		LastCommittedBlockHeight: h.lastBlockHeight,
		ResultsBlockHeader:       (&protocol.ResultsBlockHeaderBuilder{Timestamp: h.lastBlockTimestamp, BlockHeight: h.lastBlockHeight}).Build(), //TODO ResultsBlockHeader is too much info here, awaiting change in proto, see issue #121
		TransactionReceipts:      asReceipts(transactions),
	})

}

func (h *harness) verifyMocks() error {
	if _, err := h.gossip.Verify(); err != nil {
		return err
	}

	if _, err := h.trh.Verify(); err != nil {
		return err
	}

	if _, err := h.vm.Verify(); err != nil {
		return err
	}

	return nil
}

func (h *harness) handleForwardFrom(sender *keys.Ed25519KeyPair, transactions ...*protocol.SignedTransaction) {

	//TODO this is copying and needs to go away pending issue #119
	var allTransactions []byte
	for _, tx := range transactions {
		allTransactions = append(allTransactions, tx.Raw()...)
	}

	sig, err := signature.SignEd25519(sender.PrivateKey(), allTransactions)
	if err != nil {
		panic(err)
	}

	h.txpool.HandleForwardedTransactions(&gossiptopics.ForwardedTransactionsInput{
		Message: &gossipmessages.ForwardedTransactionsMessage{
			Sender: (&gossipmessages.SenderSignatureBuilder{
				SenderPublicKey: sender.PublicKey(),
				Signature:       sig,
			}).Build(),
			SignedTransactions: transactions,
		},
	})
}
func (h *harness) expectTransactionResultsCallbackFor(transactions ...*protocol.SignedTransaction) {
	h.trh.When("HandleTransactionResults", &handlers.HandleTransactionResultsInput{
		BlockHeight:         h.lastBlockHeight,
		Timestamp:           h.lastBlockTimestamp,
		TransactionReceipts: asReceipts(transactions),
	}).Times(1).Return(&handlers.HandleTransactionResultsOutput{}, nil)
}

func (h *harness) ignoringTransactionResults() {
	h.trh.When("HandleTransactionResults", mock.Any)
}

func (h *harness) assumeBlockStorageAtHeight(height primitives.BlockHeight) {
	h.lastBlockHeight = height
	h.lastBlockTimestamp = primitives.TimestampNano(time.Now().UnixNano())
}

func (h *harness) getTransactionsForOrdering(maxNumOfTransactions uint32) (*services.GetTransactionsForOrderingOutput, error) {
	return h.txpool.GetTransactionsForOrdering(&services.GetTransactionsForOrderingInput{
		MaxNumberOfTransactions: maxNumOfTransactions,
	})
}

func (h *harness) failPreOrderCheckFor(failOn func(tx *protocol.SignedTransaction) bool) {
	h.vm.Reset().When("TransactionSetPreOrder", mock.Any).Call(func(input *services.TransactionSetPreOrderInput) (*services.TransactionSetPreOrderOutput, error) {
		if input.BlockHeight != h.lastBlockHeight {
			log.GetLogger().Error("Invalid block height", log.Uint64("expected-block-height", h.lastBlockHeight.KeyForMap()), log.Uint64("actual-block-height", input.BlockHeight.KeyForMap()))
			panic("Invalid block height")
		}
		statuses := make([]protocol.TransactionStatus, len(input.SignedTransactions))
		for i, tx := range input.SignedTransactions {
			if failOn(tx) {
				statuses[i] = protocol.TRANSACTION_STATUS_REJECTED_SMART_CONTRACT_PRE_ORDER
			} else {
				statuses[i] = protocol.TRANSACTION_STATUS_PRE_ORDER_VALID
			}
		}
		return &services.TransactionSetPreOrderOutput{
			PreOrderResults: statuses,
		}, nil
	})
}

func (h *harness) passAllPreOrderChecks() {
	h.failPreOrderCheckFor(func(tx *protocol.SignedTransaction) bool {
		return false
	})
}
func (h *harness) goToBlock(height primitives.BlockHeight, timestamp primitives.TimestampNano) {
	h.ignoringTransactionResults()
	currentBlock := primitives.BlockHeight(0)
	for currentBlock <= height {
		out, _ := h.txpool.CommitTransactionReceipts(&services.CommitTransactionReceiptsInput{
			LastCommittedBlockHeight: currentBlock,
			ResultsBlockHeader:       (&protocol.ResultsBlockHeaderBuilder{BlockHeight: currentBlock, Timestamp: timestamp}).Build(),
		})
		currentBlock = out.NextDesiredBlockHeight
	}
	h.lastBlockHeight = height
}

func (h *harness) validateTransactionsForOrdering(blockHeight primitives.BlockHeight, txs ...*protocol.SignedTransaction) error {
	_, err := h.txpool.ValidateTransactionsForOrdering(&services.ValidateTransactionsForOrderingInput{
		BlockHeight:        blockHeight,
		SignedTransactions: txs,
	})

	return err
}

func newHarness() *harness {
	return newHarnessWithSizeLimit(20 * 1024 * 1024)
}

func getConfig(sizeLimit uint32, transactionExpirationInSeconds time.Duration, keyPair *keys.Ed25519KeyPair) transactionpool.Config {
	cfg := config.EmptyConfig()

	cfg.SetNodePublicKey(keyPair.PublicKey())
	cfg.SetNodePrivateKey(keyPair.PrivateKey())

	cfg.SetUint32(config.VIRTUAL_CHAIN_ID, 42)
	cfg.SetDuration(config.BLOCK_TRACKER_GRACE_TIMEOUT, 100*time.Millisecond)
	cfg.SetUint32(config.BLOCK_TRACKER_GRACE_DISTANCE, 5)

	cfg.SetUint32(config.TRANSACTION_POOL_PENDING_POOL_SIZE_IN_BYTES, sizeLimit)
	cfg.SetDuration(config.TRANSACTION_POOL_TRANSACTION_EXPIRATION_WINDOW, transactionExpirationInSeconds)
	cfg.SetDuration(config.TRANSACTION_POOL_FUTURE_TIMESTAMP_GRACE_TIMEOUT, 180*time.Second)
	cfg.SetDuration(config.TRANSACTION_POOL_PENDING_POOL_CLEAR_EXPIRED_INTERVAL, 10*time.Millisecond)
	cfg.SetDuration(config.TRANSACTION_POOL_COMMITTED_POOL_CLEAR_EXPIRED_INTERVAL, 30*time.Millisecond)

	return cfg
}

func newHarnessWithSizeLimit(sizeLimit uint32) *harness {
	ctx := context.Background()

	ts := primitives.TimestampNano(time.Now().UnixNano())

	gossip := &gossiptopics.MockTransactionRelay{}
	gossip.When("RegisterTransactionRelayHandler", mock.Any).Return()

	virtualMachine := &services.MockVirtualMachine{}

	config := getConfig(sizeLimit, transactionExpirationWindow, thisNodeKeyPair)
	service := transactionpool.NewTransactionPool(ctx, gossip, virtualMachine, config, log.GetLogger(), ts)

	transactionResultHandler := &handlers.MockTransactionResultsHandler{}
	service.RegisterTransactionResultsHandler(transactionResultHandler)

	h := &harness{
		txpool:             service,
		gossip:             gossip,
		vm:                 virtualMachine,
		trh:                transactionResultHandler,
		lastBlockTimestamp: ts,
	}

	h.passAllPreOrderChecks()

	return h
}

func asReceipts(transactions transactionpool.Transactions) []*protocol.TransactionReceipt {
	var receipts []*protocol.TransactionReceipt
	for _, tx := range transactions {
		receipts = append(receipts, (&protocol.TransactionReceiptBuilder{
			Txhash: digest.CalcTxHash(tx.Transaction()),
		}).Build())
	}
	return receipts
}
