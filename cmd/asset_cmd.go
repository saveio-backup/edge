package cmd

import (
	"github.com/saveio/themis/cmd/utils"

	"github.com/urfave/cli"
)

var AssetCommand = cli.Command{
	Name:        "asset",
	Usage:       "Handle assets",
	Description: "Asset management commands can check account balance, ONT/ONG transfers, extract ONGs, and view unbound ONGs, and so on.",
	Subcommands: []cli.Command{

		{
			Action:    getBalance,
			Name:      "balance",
			Usage:     "Show balance of ont and ong of specified account",
			ArgsUsage: "<address|label|index>",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
				utils.WalletFileFlag,
			},
		},
	},
}

func getBalance(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing account argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	// addrArg := ctx.Args().First()
	// accAddr, err := cmdcom.ParseAddress(addrArg, ctx)
	// if err != nil {
	// 	return err
	// }
	// balance, err := cutils.GetBalance(accAddr)
	// if err != nil {
	// 	return err
	// }

	// PrintInfoMsg("BalanceOf:%s", accAddr)
	// // PrintInfoMsg("  ONT:%s", balance.Usdt)
	// usdt, err := strconv.ParseUint(balance.Usdt, 10, 64)
	// if err != nil {
	// 	return err
	// }
	// PrintInfoMsg("  USDT:%s", utils.FormatUsdt(usdt))
	return nil
}
