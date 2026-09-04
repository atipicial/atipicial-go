package atipicialrpc_test

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/atipicial/atipicial-go/pkg/atipicialrpc"
	"github.com/atipicial/atipicial-go/pkg/vm"
	"github.com/stretchr/testify/require"
)

func TestFaultException_ErrorsAs(t *testing.T) {
	err := atipicialrpc.GASLimitExceededException
	wrapped := fmt.Errorf("%w executing RET", err)

	// Check that FaultException can be used as a target for errors.As:
	var actual *atipicialrpc.FaultException
	require.True(t, errors.As(wrapped, &actual))
	require.Equal(t, "GAS limit exceeded", actual.Error())

	var bad *fs.PathError
	require.False(t, errors.As(wrapped, &bad))
}

func TestFaultException_ErrorsIs(t *testing.T) {
	err := &atipicialrpc.FaultException{Message: "GAS limit exceeded executing System.Contract.Call"}

	// Check that a specific FaultException can be recognized via errors.Is:
	ref := atipicialrpc.GASLimitExceededException
	require.True(t, errors.Is(err, ref))

	// Target exception message mismatch.
	require.False(t, errors.Is(err, &atipicialrpc.FaultException{Message: "some error"}))
}

func TestGASLimitExceededException_MessageCompat(t *testing.T) {
	require.Equal(t, vm.ErrGASLimitExceeded.Error(), atipicialrpc.GASLimitExceededException.Error())
}
