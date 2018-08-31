package test

import (
	"github.com/orbs-network/orbs-network-go/services/processor/native/repository/_Deployments"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProcessTransactionSetSuccess(t *testing.T) {
	h := newHarness()
	h.expectSystemContractCalled(deployments.CONTRACT.Name, deployments.METHOD_GET_INFO.Name, nil, uint32(protocol.PROCESSOR_TYPE_NATIVE)) // assume all contracts are deployed

	h.expectNativeContractMethodCalled("Contract1", "method1", func(contextId primitives.ExecutionContextId) (protocol.ExecutionResult, error) {
		t.Log("Transaction 1: successful")
		return protocol.EXECUTION_RESULT_SUCCESS, nil
	})
	h.expectNativeContractMethodCalled("Contract1", "method2", func(contextId primitives.ExecutionContextId) (protocol.ExecutionResult, error) {
		t.Log("Transaction 2: successful")
		return protocol.EXECUTION_RESULT_SUCCESS, nil
	})

	results, _ := h.processTransactionSet([]*contractAndMethod{
		{"Contract1", "method1"},
		{"Contract1", "method2"},
	})
	require.Equal(t, results, []protocol.ExecutionResult{
		protocol.EXECUTION_RESULT_SUCCESS,
		protocol.EXECUTION_RESULT_SUCCESS,
	}, "processTransactionSet returned receipts should match")

	h.verifySystemContractCalled(t)
	h.verifyNativeContractMethodCalled(t)
}

func TestProcessTransactionSetWithErrors(t *testing.T) {
	h := newHarness()
	h.expectSystemContractCalled(deployments.CONTRACT.Name, deployments.METHOD_GET_INFO.Name, nil, uint32(protocol.PROCESSOR_TYPE_NATIVE)) // assume all contracts are deployed

	h.expectNativeContractMethodCalled("Contract1", "method1", func(contextId primitives.ExecutionContextId) (protocol.ExecutionResult, error) {
		t.Log("Transaction 1: failed (contract error)")
		return protocol.EXECUTION_RESULT_ERROR_SMART_CONTRACT, errors.New("contract error")
	})
	h.expectNativeContractMethodCalled("Contract1", "method2", func(contextId primitives.ExecutionContextId) (protocol.ExecutionResult, error) {
		t.Log("Transaction 2: failed (unexpected error)")
		return protocol.EXECUTION_RESULT_ERROR_UNEXPECTED, errors.New("unexpected error")
	})

	results, _ := h.processTransactionSet([]*contractAndMethod{
		{"Contract1", "method1"},
		{"Contract1", "method2"},
	})
	require.Equal(t, results, []protocol.ExecutionResult{
		protocol.EXECUTION_RESULT_ERROR_SMART_CONTRACT,
		protocol.EXECUTION_RESULT_ERROR_UNEXPECTED,
	}, "processTransactionSet returned receipts should match")

	h.verifySystemContractCalled(t)
	h.verifyNativeContractMethodCalled(t)
}
