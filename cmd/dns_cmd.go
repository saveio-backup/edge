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
				flags.DspFileHashFlag,
				flags.DspFileNameFlag,
				flags.DspFileBlocksRoot,
				flags.DspFileOwner,
				flags.DspFileSize,
				flags.DspBlockCountSize,
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
			Action:    registerHeader,
			Name:      "registerheader",
			Usage:     "Register dns with header",
			ArgsUsage: "<header>",
			Flags: []cli.Flag{
				flags.DnsHeaderFlag,
				flags.DnsDescFlag,
				flags.DnsTTLFlag,
			},
			Description: "Register dns with header",
		},
		{
			Action:    queryLink,
			Name:      "query",
			Usage:     "Query link with url",
			ArgsUsage: "<link>",
			Flags: []cli.Flag{
				flags.DnsURLFlag,
			},
			Description: "Query link with url",
		},
		{
			Action:    getHostInfo,
			Name:      "hostinfo",
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
			Name:      "publicip",
			Usage:     "Display public ip for a node",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.DnsWalletFlag,
			},
			Description: "Display public ip for a node",
		},
	},
}

//dns command
func registerUrl(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DnsURLFlag)) || !ctx.IsSet(flags.GetFlagName(flags.DspFileHashFlag)) {
		PrintErrorMsg("Missing argument: url or fileHash")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	url := ctx.String(flags.GetFlagName(flags.DnsURLFlag))
	fileHash := ctx.String(flags.GetFlagName(flags.DspFileHashFlag))
	fileName := ctx.String(flags.GetFlagName(flags.DspFileNameFlag))
	blocksRoot := ctx.String(flags.GetFlagName(flags.DspFileBlocksRoot))
	fileOwner := ctx.String(flags.GetFlagName(flags.DspFileOwner))
	fileSize := ctx.Int64(flags.GetFlagName(flags.DspFileSize))
	totalCount := ctx.Int64(flags.GetFlagName(flags.DspBlockCountSize))
	ret, err := utils.RegisterUrl(url, fileHash, fileName, blocksRoot, fileOwner, fileSize, totalCount)
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

func registerHeader(ctx *cli.Context) error {
	if ctx.NumFlags() < 3 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	header := ctx.String(flags.GetFlagName(flags.DnsHeaderFlag))
	desc := ctx.String(flags.GetFlagName(flags.DnsDescFlag))
	ttl := ctx.Uint64(flags.GetFlagName(flags.DnsTTLFlag))
	ret, err := utils.RegisterDnsHeader(header, desc, ttl)
	if err != nil {
		return err
	}
	// result := make(map[string]interface{})
	// result["Tx"] = string(ret)
	PrintJsonData(ret)
	return nil
}

func queryLink(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DnsURLFlag)) {
		PrintErrorMsg("Missing argument: url")
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
