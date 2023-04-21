package cmd

import (
	"encoding/json"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	chainCom "github.com/saveio/themis/common"
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
			Flags: []cli.Flag{
				flags.VerboseFlag,
			},
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
			Flags: []cli.Flag{
				flags.DspWalletAddrFlag,
			},
		},
	},
	Description: `./dsp node --help command to view help information.`,
}

// node command
func registerNode(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspVolumeFlag)) {
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
		PrintErrorMsg("query node err:%s\n", err)
		return err
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(ret, &m); err != nil {
		return err
	}
	info := m["Info"].(map[string]interface{})
	buf, err := json.Marshal(info["WalletAddr"])
	if err != nil {
		return err
	}
	bufs := make([]byte, 0)
	if err := json.Unmarshal(buf, &bufs); err != nil {
		return err
	}
	addr, err := chainCom.AddressParseFromBytes(bufs)
	if err != nil {
		return err
	}
	info["Pledge"] = utils.ParseAssets(info["Pledge"])
	// info["RestVol"] = utils.ParserByteSizeToKB(info["RestVol"])
	// info["Volume"] = utils.ParserByteSizeToKB(info["Volume"])
	info["WalletAddr"] = addr.ToBase58()
	PrintJsonObject(m)
	return nil
}

func updateNode(ctx *cli.Context) error {
	nodeAddr := ctx.String(flags.GetFlagName(flags.DspNodeAddrFlag))
	volume := ctx.String(flags.GetFlagName(flags.DspVolumeFlag))
	serviceTime := ctx.String(flags.GetFlagName(flags.DspServiceTimeFlag))
	ret, err := utils.NodeUpdate(nodeAddr, volume, serviceTime)
	if err != nil {
		PrintErrorMsg("update node err:%s\n", err)
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
