package cmd

import (
	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	"github.com/urfave/cli"
)

var SectorCommand = cli.Command{
	Name:  "sector",
	Usage: "Sector Management",
	Subcommands: []cli.Command{
		{
			Action:    createSector,
			Name:      "create",
			Usage:     "Create sector",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspSectorIdFlag,
				flags.DspSectorProveLevelFlag,
				flags.DspSectorSizeFlag,
			},
			Description: "Create sector",
		},

		{
			Action:    deleteSector,
			Name:      "delete",
			Usage:     "Delete sector",
			ArgsUsage: "<sectorid>",
			Flags: []cli.Flag{
				flags.DspSectorIdFlag,
			},
			Description: "Delete sector",
		},
		{
			Action:    getSectorInfo,
			Name:      "getsector",
			Usage:     "Get sector info",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspSectorIdFlag,
			},
			Description: "Get sector info",
		},
		{
			Action:    getSectorInfosForNode,
			Name:      "getsectorsfornode",
			Usage:     "Get all sector info for a node",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspWalletAddrFlag,
			},
			Description: "Get all sector info for a node",
		},
	},
	Description: `./dsp sector --help command to view help information.`,
}

//node command
func createSector(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspSectorIdFlag)) ||
		!ctx.IsSet(flags.GetFlagName(flags.DspSectorProveLevelFlag)) ||
		!ctx.IsSet(flags.GetFlagName(flags.DspSectorSizeFlag)) {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	sectorId := ctx.String(flags.GetFlagName(flags.DspSectorIdFlag))
	proveLevel := ctx.Uint64(flags.GetFlagName(flags.DspSectorProveLevelFlag))
	size := ctx.Uint64(flags.GetFlagName(flags.DspSectorSizeFlag))
	ret, err := utils.CreateSector(sectorId, proveLevel, size)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func deleteSector(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspSectorIdFlag)) {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	sectorId := ctx.String(flags.GetFlagName(flags.DspSectorIdFlag))
	ret, err := utils.DeleteSector(sectorId)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func getSectorInfo(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspSectorIdFlag)) {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	sectorId := ctx.String(flags.GetFlagName(flags.DspSectorIdFlag))
	ret, err := utils.GetSectorInfo(sectorId)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func getSectorInfosForNode(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspWalletAddrFlag)) {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	addr := ctx.String(flags.GetFlagName(flags.DspWalletAddrFlag))
	ret, err := utils.GetSectorInfosForNode(addr)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}
