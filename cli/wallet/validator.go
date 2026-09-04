package wallet

import (
	"fmt"
	"math/big"

	"github.com/atipicial/atipicial-go/cli/cmdargs"
	"github.com/atipicial/atipicial-go/cli/flags"
	"github.com/atipicial/atipicial-go/cli/options"
	"github.com/atipicial/atipicial-go/cli/txctx"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativehashes"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/crypto/keys"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/gas"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/atipicial"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/aep17"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/wallet"
	"github.com/urfave/cli/v2"
)

func newValidatorCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:      "register",
			Usage:     "Register as a new candidate",
			UsageText: "register -w <path> -r <rpc> [-s timeout] -a <addr> [-g gas] [-e sysgas] [--out file] [--force] [--await]",
			Action:    handleRegister,
			Flags: append([]cli.Flag{
				walletPathFlag,
				walletConfigFlag,
				txctx.GasFlag,
				txctx.SysGasFlag,
				txctx.OutFlag,
				txctx.ForceFlag,
				txctx.AwaitFlag,
				&flags.AddressFlag{
					Name:     "address",
					Aliases:  []string{"a"},
					Required: true,
					Usage:    "Address to register",
				},
				&cli.BoolFlag{
					Name:  "useRegisterCall",
					Usage: "Use a call to registerCandidate method instead of AEP-17 GAS transfer",
				},
			}, options.RPC...),
		},
		{
			Name:      "unregister",
			Usage:     "Unregister self as a candidate",
			UsageText: "unregister -w <path> -r <rpc> [-s timeout] -a <addr> [-g gas] [-e sysgas] [--out file] [--force] [--await]",
			Action:    handleUnregister,
			Flags: append([]cli.Flag{
				walletPathFlag,
				walletConfigFlag,
				txctx.GasFlag,
				txctx.SysGasFlag,
				txctx.OutFlag,
				txctx.ForceFlag,
				txctx.AwaitFlag,
				&flags.AddressFlag{
					Name:     "address",
					Required: true,
					Aliases:  []string{"a"},
					Usage:    "Address to unregister",
				},
			}, options.RPC...),
		},
		{
			Name:      "vote",
			Usage:     "Vote for a validator",
			UsageText: "vote -w <path> -r <rpc> [-s <timeout>] [-g gas] [-e sysgas] -a <addr> [-c <public key>] [--out file] [--force] [--await]",
			Description: `Votes for a validator by calling "vote" method of a ATC native
   contract. Do not provide candidate argument to perform unvoting. If --await flag is 
   included, the command waits for the transaction to be included in a block before exiting.
`,
			Action: handleVote,
			Flags: append([]cli.Flag{
				walletPathFlag,
				walletConfigFlag,
				txctx.GasFlag,
				txctx.SysGasFlag,
				txctx.OutFlag,
				txctx.ForceFlag,
				txctx.AwaitFlag,
				&flags.AddressFlag{
					Name:     "address",
					Required: true,
					Aliases:  []string{"a"},
					Usage:    "Address to vote from",
				},
				&cli.StringFlag{
					Name:    "candidate",
					Aliases: []string{"c"},
					Usage:   "Public key of candidate to vote for",
				},
			}, options.RPC...),
		},
	}
}

func handleRegister(ctx *cli.Context) error {
	if ctx.Bool("useRegisterCall") {
		return handleAtipicialAction(ctx, func(contract *atipicial.Contract, _ util.Uint160, acc *wallet.Account) (*transaction.Transaction, error) {
			return contract.RegisterCandidateUnsigned(acc.PublicKey())
		})
	}
	return handleGasAction(ctx, func(nc *atipicial.Contract, gasT *aep17.Token, _ util.Uint160, acc *wallet.Account) (*transaction.Transaction, error) {
		regPrice, err := nc.GetRegisterPrice()
		if err != nil {
			return nil, err
		}
		return gasT.TransferUnsigned(
			acc.ScriptHash(),
			nativehashes.AtipicialCoin,
			big.NewInt(regPrice),
			acc.PublicKey().Bytes(),
		)
	})
}

func handleUnregister(ctx *cli.Context) error {
	return handleAtipicialAction(ctx, func(contract *atipicial.Contract, _ util.Uint160, acc *wallet.Account) (*transaction.Transaction, error) {
		return contract.UnregisterCandidateUnsigned(acc.PublicKey())
	})
}

func handleAtipicialAction(ctx *cli.Context, mkTx func(*atipicial.Contract, util.Uint160, *wallet.Account) (*transaction.Transaction, error)) error {
	return handleTokenAction(ctx, func(nc *atipicial.Contract, _ *aep17.Token, addr util.Uint160, acc *wallet.Account) (*transaction.Transaction, error) {
		return mkTx(nc, addr, acc)
	}, transaction.CalledByEntry)
}

func handleGasAction(ctx *cli.Context, mkTx func(*atipicial.Contract, *aep17.Token, util.Uint160, *wallet.Account) (*transaction.Transaction, error)) error {
	return handleTokenAction(ctx, mkTx, transaction.Global)
}

func handleTokenAction(
	ctx *cli.Context,
	mkTx func(nc *atipicial.Contract, gasT *aep17.Token, addr util.Uint160, acc *wallet.Account) (*transaction.Transaction, error),
	scope transaction.WitnessScope,
) error {
	if err := cmdargs.EnsureNone(ctx); err != nil {
		return err
	}
	wall, pass, err := readWallet(ctx)
	if err != nil {
		return cli.Exit(err, 1)
	}
	defer wall.Close()

	addrFlag := ctx.Generic("address").(*flags.Address)
	addr := addrFlag.Uint160()
	acc, err := options.GetUnlockedAccount(wall, addr, pass)
	if err != nil {
		return cli.Exit(err, 1)
	}

	gctx, cancel := options.GetTimeoutContext(ctx)
	defer cancel()

	signers, err := cmdargs.GetSignersAccounts(acc, wall, nil, scope)
	if err != nil {
		return cli.Exit(fmt.Errorf("invalid signers: %w", err), 1)
	}
	_, act, exitErr := options.GetRPCWithActor(gctx, ctx, signers)
	if exitErr != nil {
		return exitErr
	}

	atipicialT := atipicial.New(act)
	gasT := gas.New(act)
	tx, err := mkTx(atipicialT, gasT, addr, acc)
	if err != nil {
		return cli.Exit(err, 1)
	}
	return txctx.SignAndSend(ctx, act, acc, tx)
}

func handleVote(ctx *cli.Context) error {
	return handleAtipicialAction(ctx, func(contract *atipicial.Contract, addr util.Uint160, acc *wallet.Account) (*transaction.Transaction, error) {
		var (
			err error
			pub *keys.PublicKey
		)
		pubStr := ctx.String("candidate")
		if pubStr != "" {
			pub, err = keys.NewPublicKeyFromString(pubStr)
			if err != nil {
				return nil, fmt.Errorf("invalid public key: '%s'", pubStr)
			}
		}

		return contract.VoteUnsigned(addr, pub)
	})
}
