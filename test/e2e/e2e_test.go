package e2e

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orbs-network/orbs-network-go/bootstrap"
	"github.com/orbs-network/orbs-network-go/config"
	"github.com/orbs-network/orbs-network-go/instrumentation"
	"github.com/orbs-network/orbs-network-go/test/builders"
	"github.com/orbs-network/orbs-network-go/test/crypto/keys"
	gossipAdapter "github.com/orbs-network/orbs-network-go/test/harness/services/gossip/adapter"
	"github.com/orbs-network/orbs-spec/types/go/protocol/consensus"
	"os"
	"testing"
	"time"
)

func TestE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

var _ = Describe("The Orbs Network", func() {
	It("accepts a transaction and reflects the state change after it is committed", func(done Done) {
		var nodes []bootstrap.Node

		// TODO: kill me - why do we need this override?
		if getConfig().Bootstrap {
			gossipTransport := gossipAdapter.NewTamperingTransport()

			federationNodes := make(map[string]config.FederationNode)
			leaderKeyPair := keys.Ed25519KeyPairForTests(0)
			for i := 0; i < 3; i++ {
				nodeKeyPair := keys.Ed25519KeyPairForTests(i)
				federationNodes[nodeKeyPair.PublicKey().KeyForMap()] = config.NewHardCodedFederationNode(nodeKeyPair.PublicKey())
			}

			logger := instrumentation.GetLogger().WithOutput(instrumentation.NewOutput(os.Stdout).WithFormatter(instrumentation.NewHumanReadableFormatter()))

			for i := 0; i < 3; i++ {
				nodeKeyPair := keys.Ed25519KeyPairForTests(i)
				node := bootstrap.NewNode(
					fmt.Sprintf(":%d", 8080+i),
					nodeKeyPair.PublicKey(),
					nodeKeyPair.PrivateKey(),
					federationNodes, 70,
					5,
					5,
					30*60,
					leaderKeyPair.PublicKey(),
					consensus.CONSENSUS_ALGO_TYPE_BENCHMARK_CONSENSUS,
					logger,
					2*1000,
					gossipTransport,
					5,
					3,
					300,
					300,
					0,
				)

				nodes = append(nodes, node)
			}

			// To let node start up properly, otherwise in Docker we get connection refused
			time.Sleep(100 * time.Millisecond)
		}

		tx := builders.TransferTransaction().WithAmount(17).Builder()
		SendTransaction(tx)

		Eventually(GetBalance).Should(BeEquivalentTo(17))

		if getConfig().Bootstrap {
			for _, node := range nodes {
				node.GracefulShutdown(1 * time.Second)
			}
		}

		close(done)
	}, 10)
})
