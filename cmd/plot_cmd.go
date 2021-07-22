package cmd

import (
	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
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
				flags.PlotPathFlag,
			},
			Description: "Generate plotfile",
		},
	},
	Description: `./dsp plot --help command to view help information.`,
}

// plot command
func generatePlotFile(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.PlotSystemFlag)) ||
		!ctx.IsSet(flags.GetFlagName(flags.PlotNumericIDFlag)) ||
		!ctx.IsSet(flags.GetFlagName(flags.PlotStartNonceFlag)) ||
		!ctx.IsSet(flags.GetFlagName(flags.PlotNoncesFlag)) {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	system := ctx.String(flags.GetFlagName(flags.PlotSystemFlag))
	numericId := ctx.String(flags.GetFlagName(flags.PlotNumericIDFlag))
	path := ctx.String(flags.GetFlagName(flags.PlotPathFlag))
	start := ctx.Uint64(flags.GetFlagName(flags.PlotStartNonceFlag))
	nonces := ctx.Uint64(flags.GetFlagName(flags.PlotNoncesFlag))

	ret, err := utils.GeneratePlotFile(system, numericId, path, start, nonces)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}
