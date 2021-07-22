package cmd

import (
	"fmt"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	eUtils "github.com/saveio/edge/utils"
	"github.com/saveio/edge/utils/plot"
	"github.com/saveio/themis/account"
	"github.com/urfave/cli"
)

var PlotCommand = cli.Command{
	Name:  "plot",
	Usage: "Plot file",
	Subcommands: []cli.Command{
		{
			Action:    generatePlotFile,
			Name:      "generate",
			Usage:     "Generate plotfile",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.PlotSystemFlag,
				flags.PlotNumericIDFlag,
				flags.PlotStartNonceFlag,
				flags.PlotNoncesFlag,
				flags.PlotNumFlag,
				flags.PlotSizeFlag,
				flags.PlotPathFlag,
			},
			Description: "Generate plotfile",
		},
	},
	Description: `./dsp plot --help command to view help information.`,
}

// plot command
func generatePlotFile(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.PlotSystemFlag)) {
		PrintErrorMsg("Missing argument --system.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if !ctx.IsSet(flags.GetFlagName(flags.PlotStartNonceFlag)) &&
		!ctx.IsSet(flags.GetFlagName(flags.PlotNoncesFlag)) &&
		!ctx.IsSet(flags.GetFlagName(flags.PlotSizeFlag)) {
		PrintErrorMsg("Missing argument. --size or --nonce --startNonce")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	system := ctx.String(flags.GetFlagName(flags.PlotSystemFlag))
	numericId := ctx.String(flags.GetFlagName(flags.PlotNumericIDFlag))
	if len(numericId) == 0 {
		numericId, _ = numericdFromDefaultWallet(ctx)
	}
	path := ctx.String(flags.GetFlagName(flags.PlotPathFlag))
	start := ctx.Uint64(flags.GetFlagName(flags.PlotStartNonceFlag))
	nonces := ctx.Uint64(flags.GetFlagName(flags.PlotNoncesFlag))
	size := ctx.Uint64(flags.GetFlagName(flags.PlotSizeFlag))
	if nonces == 0 {
		nonces = size / 256
		var err error
		start, err = plot.GetMinStartNonce(numericId, path)
		if err != nil {
			return err
		}
	}

	num := ctx.Uint64(flags.GetFlagName(flags.PlotNumFlag))
	for i := uint64(0); i < num; i++ {
		ret, err := utils.GeneratePlotFile(system, numericId, path, start, nonces)
		if err != nil {
			return err
		}
		PrintJsonData(ret)
	}

	return nil
}

func numericdFromDefaultWallet(ctx *cli.Context) (string, error) {
	optionFile := checkWalletFileName(ctx)
	wallet, err := account.Open(optionFile)
	if err != nil {
		return "", fmt.Errorf("open wallet:%s error:%s", optionFile, err)
	}
	accNum := wallet.GetAccountNum()
	if accNum == 0 {
		return "", fmt.Errorf("no account for wallet")
	}
	accMeta := wallet.GetAccountMetadataByIndex(1)
	return fmt.Sprintf("%v", eUtils.WalletAddressToId([]byte(accMeta.Address))), nil
}
