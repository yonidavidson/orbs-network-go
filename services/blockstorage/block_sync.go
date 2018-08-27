package blockstorage

import (
	"context"
	"github.com/orbs-network/orbs-network-go/instrumentation/log"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/gossipmessages"
	"github.com/orbs-network/orbs-spec/types/go/services/gossiptopics"
)

type BlockSyncStorage interface {
	GetBlocks(first primitives.BlockHeight, last primitives.BlockHeight) (blocks []*protocol.BlockPairContainer, firstAvailableBlockHeight primitives.BlockHeight, lastAvailableBlockHeight primitives.BlockHeight)
	LastCommittedBlockHeight() primitives.BlockHeight
}

type BlockSync struct {
	reporting log.BasicLogger

	config  Config
	storage BlockSyncStorage
	gossip  gossiptopics.BlockSync
	Events  chan interface{}
}

func NewBlockSync(ctx context.Context, storage BlockSyncStorage, gossip gossiptopics.BlockSync, config Config, reporting log.BasicLogger) *BlockSync {
	blockSync := &BlockSync{
		reporting: reporting.For(log.String("flow", "block-sync")),
		storage:   storage,
		gossip:    gossip,
		config:    config,
		Events:    make(chan interface{}),
	}

	go blockSync.mainLoop(ctx)

	return blockSync
}

func (b *BlockSync) mainLoop(ctx context.Context) {
	for {
		//isActive := false
		//var syncSource primitives.Ed25519PublicKey

		select {
		case <-ctx.Done():
			return
		case event := <-b.Events:
			switch event.(type) {
			case *gossiptopics.BlockAvailabilityResponseInput:
				input := event.(*gossiptopics.BlockAvailabilityResponseInput)

				b.reporting.Info("Received block availability response", log.Stringable("sender", input.Message.Sender))

				senderPublicKey := input.Message.Sender.SenderPublicKey()

				//if isActive && !syncSource.Equal(senderPublicKey) {
				//	continue
				//}

				lastCommittedBlockHeight := b.storage.LastCommittedBlockHeight()

				if lastCommittedBlockHeight >= input.Message.SignedRange.LastCommittedBlockHeight() {
					continue
				}

				//syncSource = senderPublicKey
				//isActive = true

				blockType := input.Message.SignedRange.BlockType()

				lastAvailableBlockHeight := lastCommittedBlockHeight + primitives.BlockHeight(b.config.BlockSyncBatchSize())
				firstAvailableBlockHeight := lastCommittedBlockHeight + 1

				request := &gossiptopics.BlockSyncRequestInput{
					RecipientPublicKey: senderPublicKey,
					Message: &gossipmessages.BlockSyncRequestMessage{
						Sender: (&gossipmessages.SenderSignatureBuilder{
							SenderPublicKey: b.config.NodePublicKey(),
						}).Build(),
						SignedRange: (&gossipmessages.BlockSyncRangeBuilder{
							BlockType:                 blockType,
							LastAvailableBlockHeight:  lastAvailableBlockHeight,
							FirstAvailableBlockHeight: firstAvailableBlockHeight,
							LastCommittedBlockHeight:  lastCommittedBlockHeight,
						}).Build(),
					},
				}

				b.gossip.SendBlockSyncRequest(request)
			case *gossiptopics.BlockSyncRequestInput:
				input := event.(*gossiptopics.BlockSyncRequestInput)
				b.reporting.Info("Received block sync request", log.Stringable("sender", input.Message.Sender))

				senderPublicKey := input.Message.Sender.SenderPublicKey()
				blockType := input.Message.SignedRange.BlockType()

				lastCommittedBlockHeight := b.storage.LastCommittedBlockHeight()
				firstRequestedBlockHeight := input.Message.SignedRange.FirstAvailableBlockHeight()
				lastRequestedBlockHeight := input.Message.SignedRange.LastAvailableBlockHeight()

				if firstRequestedBlockHeight-lastCommittedBlockHeight > primitives.BlockHeight(b.config.BlockSyncBatchSize()-1) {
					lastRequestedBlockHeight = firstRequestedBlockHeight + primitives.BlockHeight(b.config.BlockSyncBatchSize()-1)
				}

				blocks, firstAvailableBlockHeight, lastAvailableBlockHeight := b.storage.GetBlocks(firstRequestedBlockHeight, lastRequestedBlockHeight)

				b.reporting.Info("Sending blocks to another node via block sync",
					log.Stringable("recipient", senderPublicKey),
					log.Stringable("first-available-block-height", firstAvailableBlockHeight),
					log.Stringable("last-available-block-height", lastAvailableBlockHeight))

				response := &gossiptopics.BlockSyncResponseInput{
					RecipientPublicKey: senderPublicKey,
					Message: &gossipmessages.BlockSyncResponseMessage{
						Sender: (&gossipmessages.SenderSignatureBuilder{
							SenderPublicKey: b.config.NodePublicKey(),
						}).Build(),
						SignedRange: (&gossipmessages.BlockSyncRangeBuilder{
							BlockType:                 blockType,
							FirstAvailableBlockHeight: firstAvailableBlockHeight,
							LastAvailableBlockHeight:  lastAvailableBlockHeight,
							LastCommittedBlockHeight:  lastCommittedBlockHeight,
						}).Build(),
						BlockPairs: blocks,
					},
				}

				b.gossip.SendBlockSyncResponse(response)
			}
		}
	}
}
