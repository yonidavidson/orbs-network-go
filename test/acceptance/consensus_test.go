package acceptance

import (
	"github.com/orbs-network/orbs-network-go/test/harness"
	"github.com/orbs-network/orbs-network-go/test/harness/services/gossip/adapter"
	"github.com/orbs-network/orbs-spec/types/go/protocol/consensus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLeanHelixLeaderGetsValidationsBeforeCommit(t *testing.T) {
	t.Skip("putting lean helix on hold until external library is integrated")
	harness.Network(t).WithConsensusAlgos(consensus.CONSENSUS_ALGO_TYPE_LEAN_HELIX).Start(func(network harness.AcceptanceTestNetwork) {

		network.DeployBenchmarkToken()

		prePrepareLatch := network.GossipTransport().LatchOn(adapter.LeanHelixMessage(consensus.LEAN_HELIX_PRE_PREPARE))
		prePrepareTamper := network.GossipTransport().Fail(adapter.LeanHelixMessage(consensus.LEAN_HELIX_PRE_PREPARE))
		<-network.SendTransfer(0, 17)

		prePrepareLatch.Wait()
		require.EqualValues(t, 0, <-network.CallGetBalance(0), "initial getBalance result on leader")
		require.EqualValues(t, 0, <-network.CallGetBalance(1), "initial getBalance result on non leader")

		prePrepareTamper.Release()
		prePrepareLatch.Remove()

		if err := network.BlockPersistence(0).GetBlockTracker().WaitForBlock(1); err != nil {
			t.Errorf("waiting for block on node 0 failed: %s", err)
		}
		require.EqualValues(t, 17, <-network.CallGetBalance(0), "eventual getBalance result on leader")

		if err := network.BlockPersistence(1).GetBlockTracker().WaitForBlock(1); err != nil {
			t.Errorf("waiting for block on node 1 failed: %s", err)
		}
		require.EqualValues(t, 17, <-network.CallGetBalance(1), "eventual getBalance result on non leader")

	})
}

func TestBenchmarkConsensusLeaderGetsVotesBeforeNextBlock(t *testing.T) {
	harness.Network(t).Start(func(network harness.AcceptanceTestNetwork) {

		network.DeployBenchmarkToken()

		committedTamper := network.GossipTransport().Fail(adapter.BenchmarkConsensusMessage(consensus.BENCHMARK_CONSENSUS_COMMITTED))
		committedLatch := network.GossipTransport().LatchOn(adapter.BenchmarkConsensusMessage(consensus.BENCHMARK_CONSENSUS_COMMITTED))

		<-network.SendTransfer(0, 0)

		committedLatch.Wait()
		committedLatch.Remove()

		txHash := network.SendTransferInBackground(0, 17)

		committedLatch = network.GossipTransport().LatchOn(adapter.BenchmarkConsensusMessage(consensus.BENCHMARK_CONSENSUS_COMMITTED))
		committedLatch.Wait()
		committedLatch.Wait()

		require.EqualValues(t, 0, <-network.CallGetBalance(0), "initial getBalance result on leader")
		require.EqualValues(t, 0, <-network.CallGetBalance(1), "initial getBalance result on non leader")

		committedLatch.Remove()
		committedTamper.Release()

		network.WaitForTransactionInState(0, txHash)
		require.EqualValues(t, 17, <-network.CallGetBalance(0), "eventual getBalance result on leader")

		network.WaitForTransactionInState(1, txHash)
		require.EqualValues(t, 17, <-network.CallGetBalance(1), "eventual getBalance result on non leader")

	})
}
