package test

import (
	"context"
	"github.com/orbs-network/orbs-network-go/test"
	"github.com/orbs-network/orbs-network-go/test/builders"
	"testing"
)

func TestHandlerOfLeaderForValidBlockConsensus(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newLeaderHarnessAndInit(t, ctx)
		aBlockFromLeader := builders.BlockPair().WithBenchmarkConsensusBlockProof(leaderPublicKey, leaderPrivateKey)

		b1 := aBlockFromLeader.WithHeight(1).Build()
		b2 := aBlockFromLeader.WithHeight(2).WithPrevBlockHash(b1).Build()
		err := h.handleBlockConsensus(b2, b1)
		if err != nil {
			t.Fatal("handle did not validate valid block:", err)
		}
	})
}

func TestHandlerOfNonLeaderForValidBlockConsensus(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newNonLeaderHarnessAndInit(t, ctx)
		aBlockFromLeader := builders.BlockPair().WithBenchmarkConsensusBlockProof(leaderPublicKey, leaderPrivateKey)

		b1 := aBlockFromLeader.WithHeight(1).Build()
		b2 := aBlockFromLeader.WithHeight(2).WithPrevBlockHash(b1).Build()
		err := h.handleBlockConsensus(b2, b1)
		if err != nil {
			t.Fatal("handle did not validate valid block:", err)
		}
	})
}

func TestHandlerForBlockConsensusWithBadPrevBlockHashPointer(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newNonLeaderHarnessAndInit(t, ctx)
		aBlockFromLeader := builders.BlockPair().WithBenchmarkConsensusBlockProof(leaderPublicKey, leaderPrivateKey)

		b1 := aBlockFromLeader.WithHeight(1).Build()
		b2 := aBlockFromLeader.WithHeight(2).Build()
		err := h.handleBlockConsensus(b2, b1)
		if err == nil {
			t.Fatal("handle did not discover blocks with bad hash pointers:", err)
		}
	})
}

func TestHandlerForBlockConsensusWithBadSignature(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newNonLeaderHarnessAndInit(t, ctx)
		aBlockFromLeader := builders.BlockPair().WithBenchmarkConsensusBlockProof(leaderPublicKey, leaderPrivateKey)

		b1 := aBlockFromLeader.WithHeight(1).Build()
		b2 := builders.BlockPair().
			WithHeight(2).
			WithPrevBlockHash(b1).
			WithInvalidBenchmarkConsensusBlockProof(leaderPublicKey, leaderPrivateKey).
			Build()
		err := h.handleBlockConsensus(b2, b1)
		if err == nil {
			t.Fatal("handle did not discover blocks with bad signature:", err)
		}
	})
}

func TestHandlerForBlockConsensusFromNonLeader(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newNonLeaderHarnessAndInit(t, ctx)
		otherNonLeaderPublicKey, otherNonLeaderPrivateKey := otherNonLeaderKeyPair()
		aBlockFromNonLeader := builders.BlockPair().WithBenchmarkConsensusBlockProof(otherNonLeaderPublicKey, otherNonLeaderPrivateKey)

		b1 := aBlockFromNonLeader.WithHeight(1).Build()
		b2 := aBlockFromNonLeader.WithHeight(2).WithPrevBlockHash(b1).Build()
		err := h.handleBlockConsensus(b2, b1)
		if err == nil {
			t.Fatal("handle did not discover blocks not from the leader:", err)
		}
	})
}

// TODO: rely on future block to set lastCommitted
