package test

import (
	"github.com/orbs-network/orbs-network-go/services/processor/native/repository/_Deployments"
	"github.com/orbs-network/orbs-network-go/test/builders"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRunLocalMethodSuccess(t *testing.T) {
	h := newHarness()
	h.expectSystemContractCalled(deployments.CONTRACT.Name, deployments.METHOD_GET_INFO.Name, nil, uint32(protocol.PROCESSOR_TYPE_NATIVE)) // assume all contracts are deployed

	h.expectStateStorageBlockHeightRequested(12)
	h.expectNativeContractMethodCalled("Contract1", "method1", func(contextId primitives.ExecutionContextId) (protocol.ExecutionResult, *protocol.MethodArgumentArray, error) {
		return protocol.EXECUTION_RESULT_SUCCESS, builders.MethodArgumentsArray(uint32(17), "hello", []byte{0x01, 0x02}), nil
	})

	result, outputArgs, refHeight, err := h.runLocalMethod("Contract1", "method1")
	require.NoError(t, err, "run local method should not fail")
	require.Equal(t, protocol.EXECUTION_RESULT_SUCCESS, result, "run local method should return successful result")
	require.Equal(t, builders.MethodArgumentsOpaque(uint32(17), "hello", []byte{0x01, 0x02}), outputArgs, "run local method should return matching output args")
	require.EqualValues(t, 12, refHeight)

	h.verifySystemContractCalled(t)
	h.verifyStateStorageBlockHeightRequested(t)
	h.verifyNativeContractMethodCalled(t)
}

func TestRunLocalMethodContractError(t *testing.T) {
	h := newHarness()
	h.expectSystemContractCalled(deployments.CONTRACT.Name, deployments.METHOD_GET_INFO.Name, nil, uint32(protocol.PROCESSOR_TYPE_NATIVE)) // assume all contracts are deployed

	h.expectStateStorageBlockHeightRequested(12)
	h.expectNativeContractMethodCalled("Contract1", "method1", func(contextId primitives.ExecutionContextId) (protocol.ExecutionResult, *protocol.MethodArgumentArray, error) {
		return protocol.EXECUTION_RESULT_ERROR_SMART_CONTRACT, builders.MethodArgumentsArray(), errors.New("contract error")
	})

	result, outputArgs, refHeight, err := h.runLocalMethod("Contract1", "method1")
	require.Error(t, err, "run local method should fail")
	require.Equal(t, protocol.EXECUTION_RESULT_ERROR_SMART_CONTRACT, result, "run local method should return contract error")
	require.Equal(t, []byte{}, outputArgs, "run local method should return matching output args")
	require.EqualValues(t, 12, refHeight)

	h.verifySystemContractCalled(t)
	h.verifyStateStorageBlockHeightRequested(t)
	h.verifyNativeContractMethodCalled(t)
}

func TestRunLocalMethodUnexpectedError(t *testing.T) {
	h := newHarness()
	h.expectSystemContractCalled(deployments.CONTRACT.Name, deployments.METHOD_GET_INFO.Name, nil, uint32(protocol.PROCESSOR_TYPE_NATIVE)) // assume all contracts are deployed

	h.expectStateStorageBlockHeightRequested(12)
	h.expectNativeContractMethodCalled("Contract1", "method1", func(contextId primitives.ExecutionContextId) (protocol.ExecutionResult, *protocol.MethodArgumentArray, error) {
		return protocol.EXECUTION_RESULT_ERROR_UNEXPECTED, builders.MethodArgumentsArray(), errors.New("unexpected error")
	})

	result, outputArgs, refHeight, err := h.runLocalMethod("Contract1", "method1")
	require.Error(t, err, "run local method should fail")
	require.Equal(t, protocol.EXECUTION_RESULT_ERROR_UNEXPECTED, result, "run local method should return unexpected error")
	require.Equal(t, []byte{}, outputArgs, "run local method should return matching output args")
	require.EqualValues(t, 12, refHeight)

	h.verifySystemContractCalled(t)
	h.verifyStateStorageBlockHeightRequested(t)
	h.verifyNativeContractMethodCalled(t)
}
