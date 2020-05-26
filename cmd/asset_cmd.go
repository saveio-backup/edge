package cmd

import (
	"github.com/saveio/edge/cmd/flags"
	cmdutil "github.com/saveio/edge/cmd/utils"
	"github.com/saveio/themis/cmd/utils"
	"github.com/urfave/cli"
)

var AssetCommand = cli.Command{
	Name:        "asset",
	Usage:       "Handle assets",
	Description: "Asset management commands can check account balance.",
	Subcommands: []cli.Command{

		{
			Action:    getBalance,
			Name:      "balance",
			Usage:     "Show balance of specified account",
			ArgsUsage: "<address>",
			Flags: []cli.Flag{
				flags.DspWalletAddrFlag,
			},
		},
	},
}

func getBalance(ctx *cli.Context) error {
	addr := ctx.String(utils.GetFlagName(flags.DspWalletAddrFlag))
	balance, err := cmdutil.GetBalance(addr)
	if err != nil {
		return err
	}
	PrintInfoMsg("BalanceOf:%s", addr)
	PrintJsonData(balance)
	return nil
}
