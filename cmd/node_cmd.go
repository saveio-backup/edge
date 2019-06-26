package cmd

import (
	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	"github.com/urfave/cli"
)

var NodeCommand = cli.Command{
	Name:  "node",
	Usage: "Display information about the dsp",
	Subcommands: []cli.Command{
		{
			Action:    registerNode,
			Name:      "register",
			Usage:     "Register node to themis",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspNodeAddrFlag,
				flags.DspVolumeFlag,
				flags.DspServiceTimeFlag,
			},
		},
		{
			Action:      unregisterNode,
			Name:        "unregister",
			Usage:       "Unregister node from themis",
			Description: "Unregister node",
		},
		{
			Action:    queryNode,
			Name:      "query",
			Usage:     "Query node info",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspWalletAddrFlag,
			},
			Description: "Query node info",
		},
		{
			Action:    updateNode,
			Name:      "update",
			Usage:     "Update node info",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspNodeAddrFlag,
				flags.DspVolumeFlag,
				flags.DspServiceTimeFlag,
			},
			Description: "Update node info",
		},
		{
			Action:      withDraw,
			Name:        "withdraw",
			Usage:       "Withdraw node profit",
			Description: "Withdraw node profit",
		},
	},
	Description: `./dsp node --help command to view help information.`,
}

//node command
func registerNode(ctx *cli.Context) error {
	if ctx.NumFlags() < 3 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	nodeAddr := ctx.String(flags.GetFlagName(flags.DspNodeAddrFlag))
	volume := ctx.String(flags.GetFlagName(flags.DspVolumeFlag))
	serviceTime := ctx.String(flags.GetFlagName(flags.DspServiceTimeFlag))
	ret, err := utils.RegisterNode(nodeAddr, volume, serviceTime)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func unregisterNode(ctx *cli.Context) error {
	ret, err := utils.UnregisterNode()
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func queryNode(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	walletAddr := ctx.String(flags.GetFlagName(flags.DspWalletAddrFlag))
	ret, err := utils.QueryNode(walletAddr)
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}
	PrintJsonData(ret)
	return nil
}

func updateNode(ctx *cli.Context) error {
	nodeAddr := ctx.String(flags.GetFlagName(flags.DspNodeAddrFlag))
	volume := ctx.String(flags.GetFlagName(flags.DspVolumeFlag))
	serviceTime := ctx.String(flags.GetFlagName(flags.DspServiceTimeFlag))
	ret, err := utils.NodeUpdate(nodeAddr, volume, serviceTime)
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}
	PrintJsonData(ret)
	return nil
}

func withDraw(ctx *cli.Context) error {
	ret, err := utils.NodeWithdrawProfit()
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}
	PrintJsonData(ret)
	return nil
}
