package publicapi

import (
	"github.com/orbs-network/orbs-network-go/gossip"
	"github.com/orbs-network/orbs-network-go/transactionpool"
	"github.com/orbs-network/orbs-network-go/ledger"
	"github.com/orbs-network/orbs-network-go/instrumentation"
	"github.com/orbs-network/orbs-spec/types/go/services"
	"github.com/orbs-network/orbs-spec/types/go/services/handlers"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
)

type publicApi struct {
	gossip          gossip.Gossip
	transactionPool transactionpool.TransactionPool
	ledger          ledger.Ledger
	events          instrumentation.Reporting
	isLeader        bool
}

func NewPublicApi(gossip gossip.Gossip,
	transactionPool transactionpool.TransactionPool,
	ledger ledger.Ledger,
	events instrumentation.Reporting,
	isLeader bool) services.PublicApi {
	return &publicApi{
		gossip: gossip,
		transactionPool: transactionPool,
		ledger: ledger,
		events: events,
		isLeader: isLeader,
	}
}


func (p *publicApi) SendTransaction(input *services.SendTransactionInput) (*services.SendTransactionOutput, error) {
	p.events.Info("enter_send_transaction")
	defer p.events.Info("exit_send_transaction")
	//TODO leader should also propagate transactions to other nodes
	if p.isLeader {
		p.transactionPool.Add(input.SignedTransaction())
	} else {
		p.gossip.ForwardTransaction(input.SignedTransaction())
	}

	output := services.SendTransactionOutputBuilder{}

	return output.Build(), nil
}

func (p *publicApi) CallMethod(input *services.CallMethodInput) (*services.CallMethodOutput, error) {
	p.events.Info("enter_call_method")
	defer p.events.Info("exit_call_method")

	output := services.CallMethodOutputBuilder{
		OutputArgument: []*protocol.MethodArgumentBuilder{
			{Name: "balance", Type: protocol.MethodArgumentTypeUint64, Uint64: uint64(p.ledger.GetState())},
		},
	}

	return output.Build(), nil
}

func (p *publicApi) GetTransactionStatus(input *services.GetTransactionStatusInput) (*services.GetTransactionStatusOutput, error) {
	panic("Not implemented")
}

func (p *publicApi) HandleTransactionResults(input *handlers.HandleTransactionResultsInput) (*handlers.HandleTransactionResultsOutput, error) {
	panic("Not implemented")
}
