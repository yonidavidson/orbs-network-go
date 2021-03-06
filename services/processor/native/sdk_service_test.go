package native

import (
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/services/handlers"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCallMethod(t *testing.T) {
	s := createServiceSdk()

	err := s.CallMethod(EXAMPLE_CONTEXT, "AnotherContract", "someMethod")
	require.NoError(t, err, "callMethod should succeed")
}

func createServiceSdk() *serviceSdk {
	return &serviceSdk{
		handler:         &contractSdkServiceCallHandlerStub{},
		permissionScope: protocol.PERMISSION_SCOPE_SYSTEM,
	}
}

type contractSdkServiceCallHandlerStub struct {
}

func (c *contractSdkServiceCallHandlerStub) HandleSdkCall(input *handlers.HandleSdkCallInput) (*handlers.HandleSdkCallOutput, error) {
	if input.PermissionScope != protocol.PERMISSION_SCOPE_SYSTEM {
		panic("permissions passed to SDK are incorrect")
	}
	switch input.MethodName {
	case "callMethod":
		return nil, nil
	default:
		return nil, errors.New("unknown method")
	}
}
