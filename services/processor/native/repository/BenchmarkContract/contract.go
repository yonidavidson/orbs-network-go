package benchmarkcontract

import (
	"errors"
	"github.com/orbs-network/orbs-network-go/services/processor/native/types"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
)

var CONTRACT = types.ContractInfo{
	Name:       "BenchmarkContract",
	Permission: protocol.PERMISSION_SCOPE_SERVICE,
	Methods: map[primitives.MethodName]types.MethodInfo{
		METHOD_INIT.Name:               METHOD_INIT,
		METHOD_NOP.Name:                METHOD_NOP,
		METHOD_ADD.Name:                METHOD_ADD,
		METHOD_SET.Name:                METHOD_SET,
		METHOD_GET.Name:                METHOD_GET,
		METHOD_ARG_TYPES.Name:          METHOD_ARG_TYPES,
		METHOD_THROW.Name:              METHOD_THROW,
		METHOD_PANIC.Name:              METHOD_PANIC,
		METHOD_INVALID_NO_ERROR.Name:   METHOD_INVALID_NO_ERROR,
		METHOD_INVALID_NO_CONTEXT.Name: METHOD_INVALID_NO_CONTEXT,
	},
	InitSingleton: newContract,
}

func newContract(base *types.BaseContract) types.Contract {
	return &contract{base}
}

type contract struct{ *types.BaseContract }

///////////////////////////////////////////////////////////////////////////

var METHOD_INIT = types.MethodInfo{
	Name:           "_init",
	External:       false,
	Access:         protocol.ACCESS_SCOPE_READ_WRITE,
	Implementation: (*contract)._init,
}

func (c *contract) _init(ctx types.Context) error {
	return nil
}

///////////////////////////////////////////////////////////////////////////

var METHOD_NOP = types.MethodInfo{
	Name:           "nop",
	External:       false,
	Access:         protocol.ACCESS_SCOPE_READ_ONLY,
	Implementation: (*contract).nop,
}

func (c *contract) nop(ctx types.Context) error {
	return nil
}

///////////////////////////////////////////////////////////////////////////

var METHOD_ADD = types.MethodInfo{
	Name:           "add",
	External:       true,
	Access:         protocol.ACCESS_SCOPE_READ_ONLY,
	Implementation: (*contract).add,
}

func (c *contract) add(ctx types.Context, a uint64, b uint64) (uint64, error) {
	return a + b, nil
}

///////////////////////////////////////////////////////////////////////////

var METHOD_SET = types.MethodInfo{
	Name:           "set",
	External:       true,
	Access:         protocol.ACCESS_SCOPE_READ_WRITE,
	Implementation: (*contract).set,
}

func (c *contract) set(ctx types.Context, a uint64) error {
	return c.State.WriteUint64ByKey(ctx, "example-key", a)
}

///////////////////////////////////////////////////////////////////////////

var METHOD_GET = types.MethodInfo{
	Name:           "get",
	External:       true,
	Access:         protocol.ACCESS_SCOPE_READ_ONLY,
	Implementation: (*contract).get,
}

func (c *contract) get(ctx types.Context) (uint64, error) {
	return c.State.ReadUint64ByKey(ctx, "example-key")
}

///////////////////////////////////////////////////////////////////////////

var METHOD_ARG_TYPES = types.MethodInfo{
	Name:           "argTypes",
	External:       true,
	Access:         protocol.ACCESS_SCOPE_READ_ONLY,
	Implementation: (*contract).argTypes,
}

func (c *contract) argTypes(ctx types.Context, a1 uint32, a2 uint64, a3 string, a4 []byte) (uint32, uint64, string, []byte, error) {
	return a1 + 1, a2 + 1, a3 + "1", append(a4, 0x01), nil
}

///////////////////////////////////////////////////////////////////////////

var METHOD_THROW = types.MethodInfo{
	Name:           "throw",
	External:       true,
	Access:         protocol.ACCESS_SCOPE_READ_ONLY,
	Implementation: (*contract).throw,
}

func (c *contract) throw(ctx types.Context) error {
	return errors.New("contract returns error")
}

///////////////////////////////////////////////////////////////////////////

var METHOD_PANIC = types.MethodInfo{
	Name:           "panic",
	External:       true,
	Access:         protocol.ACCESS_SCOPE_READ_ONLY,
	Implementation: (*contract).panic,
}

func (c *contract) panic(ctx types.Context) error {
	panic("contract panicked")
}

///////////////////////////////////////////////////////////////////////////

var METHOD_INVALID_NO_ERROR = types.MethodInfo{
	Name:           "invalidNoError",
	External:       true,
	Access:         protocol.ACCESS_SCOPE_READ_ONLY,
	Implementation: (*contract).invalidNoError,
}

func (c *contract) invalidNoError(ctx types.Context) {
	return
}

///////////////////////////////////////////////////////////////////////////

var METHOD_INVALID_NO_CONTEXT = types.MethodInfo{
	Name:           "invalidNoContext",
	External:       true,
	Access:         protocol.ACCESS_SCOPE_READ_ONLY,
	Implementation: (*contract).invalidNoContext,
}

func (c *contract) invalidNoContext() error {
	return nil
}
