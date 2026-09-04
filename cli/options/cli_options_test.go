package options_test

import (
	"flag"
	"testing"

	"github.com/atipicial/atipicial-go/cli/app"
	"github.com/atipicial/atipicial-go/cli/options"
	"github.com/atipicial/atipicial-go/internal/testcli"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestGetRPCClient(t *testing.T) {
	e := testcli.NewExecutor(t, true)

	t.Run("no endpoint", func(t *testing.T) {
		set := flag.NewFlagSet("flagSet", flag.ExitOnError)
		ctx := cli.NewContext(app.New(), set, nil)
		gctx, _ := options.GetTimeoutContext(ctx)
		_, ec := options.GetRPCClient(gctx, ctx)
		require.Equal(t, 1, ec.ExitCode())
	})

	t.Run("success", func(t *testing.T) {
		set := flag.NewFlagSet("flagSet", flag.ExitOnError)
		set.String(options.RPCEndpointFlag, "http://"+e.RPC.Addresses()[0], "")
		ctx := cli.NewContext(app.New(), set, nil)
		gctx, _ := options.GetTimeoutContext(ctx)
		_, ec := options.GetRPCClient(gctx, ctx)
		require.Nil(t, ec)
	})
}
