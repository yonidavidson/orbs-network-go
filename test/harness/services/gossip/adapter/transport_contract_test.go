package adapter

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/orbs-network-go/services/gossip/adapter"
	. "github.com/orbs-network/orbs-network-go/test"
	"github.com/orbs-network/orbs-spec/types/go/protocol/gossipmessages"
	"testing"
)

func TestContract(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gossip Transport Contract")
}

var _ = Describe("Tampering Transport", func() {
	assertContractOf(aTamperingTransport)
})

var _ = Describe("Memberlist Transport", func() {
	assertContractOf(aMemberlistTransport)
})

func assertContractOf(makeContext func() *transportContractContext) {

	/* // TODO: add me
	When("unicasting a message", func() {

		It("reaches only the intended recipient", func() {
			c := makeContext()
			header := (&gossipmessages.HeaderBuilder{
				RecipientMode: gossipmessages.RECIPIENT_LIST_MODE_BROADCAST,
				Topic: gossipmessages.HEADER_TOPIC_TRANSACTION_RELAY,
				TransactionRelay: gossipmessages.TRANSACTION_RELAY_FORWARDED_TRANSACTIONS,
				NumPayloads: 0,
			}).Build()
			payloads := [][]byte{}
			c.l2.expect(header, payloads)
			c.transport.Send(header, payloads)
			c.verify()
		})
	})
	*/

	When("broadcasting a message", func() {
		It("reaches all recipients", func() {
			c := makeContext()

			data := &adapter.TransportData{
				RecipientMode: gossipmessages.RECIPIENT_LIST_MODE_BROADCAST,
				Payloads:      [][]byte{{0x01, 0x02, 0x03}},
			}
			c.l1.expect(data.Payloads)
			c.l2.expect(data.Payloads)
			c.l3.expect(data.Payloads)

			c.transport.Send(data)
			c.verify()
		})
	})
}

type mockListener struct {
	mock.Mock
	name string
}

func (m *mockListener) OnTransportMessageReceived(payloads [][]byte) {
	m.Called(payloads)
}

func listenTo(transport adapter.Transport, name string) *mockListener {
	l := &mockListener{name: name}
	transport.RegisterListener(l, name)
	return l
}

func (m *mockListener) expect(payloads [][]byte) {
	m.When("OnTransportMessageReceived", payloads).Return().Times(1)
}

type transportContractContext struct {
	l1, l2, l3 *mockListener
	transport  adapter.Transport
}

func aTamperingTransport() *transportContractContext {
	transport := NewTamperingTransport()
	l1 := listenTo(transport, "l1")
	l2 := listenTo(transport, "l2")
	l3 := listenTo(transport, "l3")
	return &transportContractContext{l1, l2, l3, transport}
}

func aMemberlistTransport() *transportContractContext {
	config1 := adapter.MemberlistGossipConfig{"node1", 60001, []string{"127.0.0.1:60002", "127.0.0.1:60003", "127.0.0.1:60004"}}
	transport1 := adapter.NewMemberlistTransport(config1)

	config2 := adapter.MemberlistGossipConfig{"node2", 60002, []string{"127.0.0.1:60001", "127.0.0.1:60003", "127.0.0.1:60004"}}
	transport2 := adapter.NewMemberlistTransport(config2)

	config3 := adapter.MemberlistGossipConfig{"node3", 60003, []string{"127.0.0.1:60001", "127.0.0.1:60002", "127.0.0.1:60004"}}
	transport3 := adapter.NewMemberlistTransport(config3)

	config4 := adapter.MemberlistGossipConfig{"node4", 60004, []string{"127.0.0.1:60001", "127.0.0.1:60002", "127.0.0.1:60003"}}
	transport4 := adapter.NewMemberlistTransport(config4)

	l1 := listenTo(transport1, "l1")
	l2 := listenTo(transport2, "l2")
	l3 := listenTo(transport3, "l3")

	return &transportContractContext{l1, l2, l3, transport4}
}

func (c *transportContractContext) verify() {
	Eventually(c.l1).Should(ExecuteAsPlanned())
	Eventually(c.l2).Should(ExecuteAsPlanned())
	Eventually(c.l3).Should(ExecuteAsPlanned())
}
