package rpcclient_test

import (
	"context"
	"testing"

	"github.com/atipicial/atipicial-go/internal/testcli"
	"github.com/atipicial/atipicial-go/pkg/config"
	"github.com/atipicial/atipicial-go/pkg/config/netmode"
	"github.com/atipicial/atipicial-go/pkg/atipicialrpc"
	"github.com/atipicial/atipicial-go/pkg/rpcclient"
	"github.com/stretchr/testify/require"
)

func TestInternalClientClose(t *testing.T) {
	icl, err := rpcclient.NewInternal(context.TODO(), func(ctx context.Context, ch chan<- atipicialrpc.Notification) func(*atipicialrpc.Request) (*atipicialrpc.Response, error) {
		return nil
	})
	require.NoError(t, err)
	icl.Close()
	require.NoError(t, icl.GetError())
}

func TestInternalClientCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	bc, rpcSrv, netSrv := testcli.NewTestChain(t, func(c *config.Config) {
		c.ApplicationConfiguration.Consensus.UnlockWallet.Path = "../../cli/testdata/wallet1_solo.json"
	}, true)
	t.Cleanup(func() {
		rpcSrv.Shutdown()
		netSrv.Shutdown()
		bc.Close()
	})
	icl, err := rpcclient.NewInternal(ctx, rpcSrv.RegisterLocal)
	require.NoError(t, err)

	// Check Network API.
	require.NoError(t, icl.Init())
	require.Equal(t, netmode.UnitTestNet, icl.Network())

	cancel()
	icl.Close()
	require.NoError(t, icl.GetError())
}
