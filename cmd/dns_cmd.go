package cmd

import (
	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"

	"github.com/urfave/cli"
)

var DnsCommand = cli.Command{
	Name:  "dns",
	Usage: "Manage url, link",
	Subcommands: []cli.Command{
		{
			Action:    registerUrl,
			Name:      "register",
			Usage:     "Register url and link to DNS",
			ArgsUsage: "<hash>",
			Flags: []cli.Flag{
				flags.DnsURLFlag,
				flags.DnsLinkFlag,
			},
			Description: "Register url and link to DNS",
		},
		{
			Action:    bindUrl,
			Name:      "bind",
			Usage:     "Bind url with link",
			ArgsUsage: "<hash>",
			Flags: []cli.Flag{
				flags.DnsURLFlag,
				flags.DnsLinkFlag,
			},
			Description: "Bind url with link",
		},
		{
			Action:    queryLink,
			Name:      "querylink",
			Usage:     "Query link binded with url",
			ArgsUsage: "<link>",
			Flags: []cli.Flag{
				flags.DnsURLFlag,
			},
			Description: "Query link binded with url",
		},
		// dns candiate cmd
		{
			Action:    registerDns,
			Name:      "registerdns",
			Usage:     "Register dns candidate",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DnsIpFlag,
				flags.DnsPortFlag,
				flags.InitDepositFlag,
			},
			Description: "Request register as dns candidate",
		},
		{
			Action:      unregisterDns,
			Name:        "unregisterdns",
			Usage:       "Cancel previous register request",
			ArgsUsage:   " ",
			Description: "Cancel previous register request",
		},
		{
			Action:      quitDns,
			Name:        "quitdns",
			Usage:       "Quit working as dns",
			ArgsUsage:   " ",
			Description: "Quit working as dns",
		},
		{
			Action:    addPos,
			Name:      "addInitPos",
			Usage:     "Increase init deposit",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DeltaDepositFlag,
			},
			Description: "Increase init deposit",
		},
		{
			Action:    reducePos,
			Name:      "reduceInitPos",
			Usage:     "Reduce init deposit",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DeltaDepositFlag,
			},
			Description: "Reduce init deposit",
		},
		{
			Action:    getRegisterInfo,
			Name:      "getRegInfo",
			Usage:     "Display all or specified Dns register info",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DnsAllFlag,
				flags.PeerPubkeyFlag,
			},
			Description: "Display all or specified Dns register info",
		},
		{
			Action:    getHostInfo,
			Name:      "getHostInfo",
			Usage:     "Display all or specified Dns host info including ip, port",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DnsAllFlag,
				flags.DnsWalletFlag,
			},
			Description: "Display all or specified Dns host info including ip, port",
		},
		{
			Action:    getPublicIP,
			Name:      "getPublicIP",
			Usage:     "Display all or specified Dns host info including ip, port",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DnsWalletFlag,
			},
			Description: "Display all or specified Dns host info including ip, port",
		},
	},
}

//dns command
func registerUrl(ctx *cli.Context) error {
	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	url := ctx.String(flags.GetFlagName(flags.DnsURLFlag))
	link := ctx.String(flags.GetFlagName(flags.DnsLinkFlag))
	ret, err := utils.RegisterUrl(url, link)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func bindUrl(ctx *cli.Context) error {
	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	url := ctx.String(flags.GetFlagName(flags.DnsURLFlag))
	link := ctx.String(flags.GetFlagName(flags.DnsLinkFlag))
	ret, err := utils.BindUrl(url, link)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func queryLink(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	url := ctx.String(flags.GetFlagName(flags.DnsURLFlag))
	ret, err := utils.QueryLink(url)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func registerDns(ctx *cli.Context) error {
	if ctx.NumFlags() < 3 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	ip := ctx.String(flags.GetFlagName(flags.DnsIpFlag))
	port := ctx.String(flags.GetFlagName(flags.DnsPortFlag))
	initDeposit := ctx.String(flags.GetFlagName(flags.InitDepositFlag))
	ret, err := utils.RegisterDns(ip, port, initDeposit)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func unregisterDns(ctx *cli.Context) error {
	// TODO: check password
	ret, err := utils.UnregisterDns()
	if err != nil {
		return err
	}
	PrintJsonData(ret)

	return nil
}

func quitDns(ctx *cli.Context) error {
	// TODO: check password
	ret, err := utils.QuitDns()
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func addPos(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	deltaDeposit := ctx.String(flags.GetFlagName(flags.DeltaDepositFlag))
	ret, err := utils.AddPos(deltaDeposit)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func reducePos(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	deltaDeposit := ctx.String(flags.GetFlagName(flags.DeltaDepositFlag))
	ret, err := utils.ReducePos(deltaDeposit)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func getRegisterInfo(ctx *cli.Context) error {
	ret, err := utils.QueryRegInfos()
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func getHostInfo(ctx *cli.Context) error {
	DnsAllFlag := ctx.Bool(flags.GetFlagName(flags.DnsAllFlag))
	if DnsAllFlag {
		ret, err := utils.QueryHostInfos()
		if err != nil {
			return err
		}
		PrintJsonData(ret)
	} else {
		walletAddr := ctx.String(flags.GetFlagName(flags.DnsWalletFlag))
		ret, err := utils.QueryHostInfo(walletAddr)
		if err != nil {
			return err
		}
		PrintJsonData(ret)
	}
	return nil
}

func getPublicIP(ctx *cli.Context) error {
	walletAddr := ctx.String(flags.GetFlagName(flags.DnsWalletFlag))
	ret, err := utils.QueryPublicIP(walletAddr)
	if err != nil {
		return err
	}
	PrintJsonData(ret)

	return nil
}
