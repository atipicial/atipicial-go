package wallet

import (
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/atipicial/atipicial-go/cli/cmdargs"
	"github.com/atipicial/atipicial-go/cli/flags"
	"github.com/atipicial/atipicial-go/cli/options"
	"github.com/atipicial/atipicial-go/cli/txctx"
	"github.com/atipicial/atipicial-go/pkg/config"
	"github.com/atipicial/atipicial-go/pkg/encoding/address"
	"github.com/atipicial/atipicial-go/pkg/atipicialrpc/result"
	"github.com/atipicial/atipicial-go/pkg/rpcclient"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/aep11"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/vm/stackitem"
	"github.com/atipicial/atipicial-go/pkg/wallet"
	"github.com/urfave/cli/v2"
)

func newAEP11Commands() []*cli.Command {
	maxIters := strconv.Itoa(config.DefaultMaxIteratorResultItems)
	tokenAddressFlag := &flags.AddressFlag{
		Name:     "token",
		Usage:    "Token contract address or hash in LE",
		Required: true,
	}
	ownerAddressFlag := &flags.AddressFlag{
		Name:     "address",
		Usage:    "NFT owner address or hash in LE",
		Required: true,
	}
	tokenID := &cli.StringFlag{
		Name:  "id",
		Usage: "Hex-encoded token ID",
	}

	balanceFlags := slices.Clone(baseBalanceFlags)
	balanceFlags = append(balanceFlags, tokenID)
	balanceFlags = append(balanceFlags, options.RPC...)

	transferFlags := slices.Clone(baseTransferFlags)
	transferFlags = append(transferFlags, tokenID)
	transferFlags = append(transferFlags, options.RPC...)
	return []*cli.Command{
		{
			Name:      "balance",
			Usage:     "Get address balance",
			UsageText: "balance [-w wallet] [--wallet-config path] --rpc-endpoint <node> [--timeout <time>] [--address <address>] [--token <hash-or-name>] [--id <token-id>]",
			Description: `Prints AEP-11 balances for address and assets/IDs specified. One of wallet
   or address must be specified, passing both is valid too. If a wallet is
   given without an address all tokens (NFT contracts) for all accounts in
   the specified wallet are listed with all tokens (actual NFTs) insied. A
   single account can be chosen with the address option and/or a single NFT
   contract can be selected with the token option. Further, you can specify a
   particular NFT ID (hex-encoded) to display (which is mostly useful for
   divisible NFTs). Tokens can be specified by hash, address, name or symbol.
   Hashes and addresses always work (as long as they belong to a correct AEP-11
   contract), while names or symbols are matched against the token data
   stored in the wallet (see import command) or balance data returned from the
   server. If the token is not specified directly (with hash/address) and is
   not found in the wallet then depending on the balances data from the server
   this command can print no data at all or print multiple tokens for one
   account (if they use the same names/symbols).
`,
			Action: getAEP11Balance,
			Flags:  balanceFlags,
		},
		{
			Name:      "import",
			Usage:     "Import AEP-11 token to a wallet",
			UsageText: "import -w wallet [--wallet-config path] --rpc-endpoint <node> [--timeout <time>] --token <hash>",
			Action:    importAEP11Token,
			Flags:     importFlags,
		},
		{
			Name:      "info",
			Usage:     "Print imported AEP-11 token info",
			UsageText: "print -w wallet [--wallet-config path] [--token <hash-or-name>]",
			Action:    printAEP11Info,
			Flags: []cli.Flag{
				walletPathFlag,
				walletConfigFlag,
				tokenFlag,
			},
		},
		{
			Name:      "remove",
			Usage:     "Remove AEP-11 token from the wallet",
			UsageText: "remove -w wallet [--wallet-config path] --token <hash-or-name>",
			Action:    removeAEP11Token,
			Flags: []cli.Flag{
				walletPathFlag,
				walletConfigFlag,
				tokenFlag,
				txctx.ForceFlag,
			},
		},
		{
			Name:      "transfer",
			Usage:     "Transfer AEP-11 tokens",
			UsageText: "transfer -w wallet [--wallet-config path] --rpc-endpoint <node> [--timeout <time>] --from <addr> --to <addr> --token <hash-or-name> --id <token-id> [--amount string] [--await] [data] [-- <cosigner1:Scope> [<cosigner2> [...]]]",
			Action:    transferAEP11,
			Flags:     transferFlags,
			Description: `Transfers specified AEP-11 token with optional cosigners list attached to
   the transfer. Amount should be specified for divisible AEP-11
   tokens and omitted for non-divisible AEP-11 tokens. See
   'contract testinvokefunction' documentation for the details
   about cosigners syntax. If no cosigners are given then the
   sender with CalledByEntry scope will be used as the only
   signer. If --await flag is set then the command will wait
   for the transaction to be included in a block.
`,
		},
		{
			Name:      "properties",
			Usage:     "Print properties of AEP-11 token",
			UsageText: "properties --rpc-endpoint <node> [--timeout <time>] --token <hash> --id <token-id> [--historic <block/hash>]",
			Action:    printAEP11Properties,
			Flags: append([]cli.Flag{
				tokenAddressFlag,
				tokenID,
				options.Historic,
			}, options.RPC...),
		},
		{
			Name:      "ownerOf",
			Usage:     "Print owner of non-divisible AEP-11 token with the specified ID",
			UsageText: "ownerOf --rpc-endpoint <node> [--timeout <time>] --token <hash> --id <token-id> [--historic <block/hash>]",
			Action:    printAEP11NDOwner,
			Flags: append([]cli.Flag{
				tokenAddressFlag,
				tokenID,
				options.Historic,
			}, options.RPC...),
		},
		{
			Name:      "ownerOfD",
			Usage:     "Print set of owners of divisible AEP-11 token with the specified ID (" + maxIters + " will be printed at max)",
			UsageText: "ownerOfD --rpc-endpoint <node> [--timeout <time>] --token <hash> --id <token-id> [--historic <block/hash>]",
			Action:    printAEP11DOwner,
			Flags: append([]cli.Flag{
				tokenAddressFlag,
				tokenID,
				options.Historic,
			}, options.RPC...),
		},
		{
			Name:      "tokensOf",
			Usage:     "Print list of tokens IDs for the specified NFT owner (" + maxIters + " will be printed at max)",
			UsageText: "tokensOf --rpc-endpoint <node> [--timeout <time>] --token <hash> --address <addr> [--historic <block/hash>]",
			Action:    printAEP11TokensOf,
			Flags: append([]cli.Flag{
				tokenAddressFlag,
				ownerAddressFlag,
				options.Historic,
			}, options.RPC...),
		},
		{
			Name:      "tokens",
			Usage:     "Print list of tokens IDs minted by the specified NFT (optional method; " + maxIters + " will be printed at max)",
			UsageText: "tokens --rpc-endpoint <node> [--timeout <time>] --token <hash> [--historic <block/hash>]",
			Action:    printAEP11Tokens,
			Flags: append([]cli.Flag{
				tokenAddressFlag,
				options.Historic,
			}, options.RPC...),
		},
	}
}

func importAEP11Token(ctx *cli.Context) error {
	return importAEPToken(ctx, manifest.AEP11StandardName)
}

func printAEP11Info(ctx *cli.Context) error {
	return printAEPInfo(ctx, manifest.AEP11StandardName)
}

func removeAEP11Token(ctx *cli.Context) error {
	return removeAEPToken(ctx, manifest.AEP11StandardName)
}

func getAEP11Balance(ctx *cli.Context) error {
	return getAEPBalance(ctx, manifest.AEP11StandardName, func(ctx *cli.Context, c *rpcclient.Client, addrHash util.Uint160, name string, token *wallet.Token, nftID string) error {
		balances, err := c.GetAEP11Balances(addrHash)
		if err != nil {
			return err
		}
		var tokenFound bool
		for i := range balances.Balances {
			curToken := tokenFromAEP11Balance(&balances.Balances[i])
			if tokenMatch(curToken, token, name) {
				printNFTBalance(ctx, balances.Balances[i], nftID)
				tokenFound = true
			}
		}
		if name == "" || tokenFound {
			return nil
		}
		if token != nil {
			// We have an exact token, but there is no balance data for it -> print without NFTs.
			printNFTBalance(ctx, result.AEP11AssetBalance{
				Asset:    token.Hash,
				Decimals: int(token.Decimals),
				Name:     token.Name,
				Symbol:   token.Symbol,
			}, "")
		} else {
			// We have no data for this token at all, maybe it's not even correct -> complain.
			fmt.Fprintf(ctx.App.Writer, "Can't find data for %q token\n", name)
		}
		return nil
	})
}

func printNFTBalance(ctx *cli.Context, balance result.AEP11AssetBalance, nftID string) {
	fmt.Fprintf(ctx.App.Writer, "%s: %s (%s)\n", balance.Symbol, balance.Name, balance.Asset.StringLE())
	for _, tok := range balance.Tokens {
		if len(nftID) > 0 && nftID != tok.ID {
			continue
		}
		fmt.Fprintf(ctx.App.Writer, "\tToken: %s\n", tok.ID)
		fmt.Fprintf(ctx.App.Writer, "\t\tAmount: %s\n", decimalAmount(tok.Amount, balance.Decimals))
		fmt.Fprintf(ctx.App.Writer, "\t\tUpdated: %d\n", tok.LastUpdated)
	}
}

func transferAEP11(ctx *cli.Context) error {
	return transferAEP(ctx, manifest.AEP11StandardName)
}

func printAEP11NDOwner(ctx *cli.Context) error {
	return printAEP11Owner(ctx, false)
}

func printAEP11DOwner(ctx *cli.Context) error {
	return printAEP11Owner(ctx, true)
}

func printAEP11Owner(ctx *cli.Context, divisible bool) error {
	var err error
	if err := cmdargs.EnsureNone(ctx); err != nil {
		return err
	}
	tokenHash := ctx.Generic("token").(*flags.Address)
	tokenID := ctx.String("id")
	if tokenID == "" {
		return cli.Exit(errors.New("token ID should be specified"), 1)
	}
	tokenIDBytes, err := hex.DecodeString(tokenID)
	if err != nil {
		return cli.Exit(fmt.Errorf("invalid tokenID bytes: %w", err), 1)
	}

	gctx, cancel := options.GetTimeoutContext(ctx)
	defer cancel()

	_, inv, err := options.GetRPCWithInvoker(gctx, ctx, nil)
	if err != nil {
		return err
	}

	if divisible {
		n11 := aep11.NewDivisibleReader(inv, tokenHash.Uint160())
		result, err := n11.OwnerOfExpanded(tokenIDBytes, config.DefaultMaxIteratorResultItems)
		if err != nil {
			return cli.Exit(fmt.Sprintf("failed to call AEP-11 divisible `ownerOf` method: %s", err.Error()), 1)
		}
		for _, h := range result {
			fmt.Fprintln(ctx.App.Writer, address.Uint160ToString(h))
		}
	} else {
		n11 := aep11.NewNonDivisibleReader(inv, tokenHash.Uint160())
		result, err := n11.OwnerOf(tokenIDBytes)
		if err != nil {
			return cli.Exit(fmt.Sprintf("failed to call AEP-11 non-divisible `ownerOf` method: %s", err.Error()), 1)
		}
		fmt.Fprintln(ctx.App.Writer, address.Uint160ToString(result))
	}

	return nil
}

func printAEP11TokensOf(ctx *cli.Context) error {
	var err error
	tokenHash := ctx.Generic("token").(*flags.Address)
	acc := ctx.Generic("address").(*flags.Address)
	gctx, cancel := options.GetTimeoutContext(ctx)
	defer cancel()

	_, inv, err := options.GetRPCWithInvoker(gctx, ctx, nil)
	if err != nil {
		return err
	}

	n11 := aep11.NewBaseReader(inv, tokenHash.Uint160())
	result, err := n11.TokensOfExpanded(acc.Uint160(), config.DefaultMaxIteratorResultItems)
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to call AEP-11 `tokensOf` method: %s", err.Error()), 1)
	}

	for i := range result {
		fmt.Fprintln(ctx.App.Writer, hex.EncodeToString(result[i]))
	}
	return nil
}

func printAEP11Tokens(ctx *cli.Context) error {
	var err error
	if err := cmdargs.EnsureNone(ctx); err != nil {
		return err
	}
	tokenHash := ctx.Generic("token").(*flags.Address)
	gctx, cancel := options.GetTimeoutContext(ctx)
	defer cancel()

	_, inv, err := options.GetRPCWithInvoker(gctx, ctx, nil)
	if err != nil {
		return err
	}

	n11 := aep11.NewBaseReader(inv, tokenHash.Uint160())
	result, err := n11.TokensExpanded(config.DefaultMaxIteratorResultItems)
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to call optional AEP-11 `tokens` method: %s", err.Error()), 1)
	}

	for i := range result {
		fmt.Fprintln(ctx.App.Writer, hex.EncodeToString(result[i]))
	}
	return nil
}

func printAEP11Properties(ctx *cli.Context) error {
	var err error
	if err := cmdargs.EnsureNone(ctx); err != nil {
		return err
	}
	tokenHash := ctx.Generic("token").(*flags.Address)
	tokenID := ctx.String("id")
	if tokenID == "" {
		return cli.Exit(errors.New("token ID should be specified"), 1)
	}
	tokenIDBytes, err := hex.DecodeString(tokenID)
	if err != nil {
		return cli.Exit(fmt.Errorf("invalid tokenID bytes: %w", err), 1)
	}

	gctx, cancel := options.GetTimeoutContext(ctx)
	defer cancel()

	_, inv, err := options.GetRPCWithInvoker(gctx, ctx, nil)
	if err != nil {
		return err
	}

	n11 := aep11.NewBaseReader(inv, tokenHash.Uint160())
	result, err := n11.Properties(tokenIDBytes)
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to call AEP-11 `properties` method: %s", err.Error()), 1)
	}

	bytes, err := stackitem.ToJSON(result)
	if err != nil {
		return cli.Exit(fmt.Sprintf("failed to convert result to JSON: %s", err), 1)
	}
	fmt.Fprintln(ctx.App.Writer, string(bytes))
	return nil
}
