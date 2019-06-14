package cmd

import (
	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp"
	ccom "github.com/saveio/themis/common"

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
	endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	DnsAllFlag := ctx.Bool(flags.GetFlagName(flags.DnsAllFlag))

	if DnsAllFlag {
		m, err := endpoint.Dsp.Chain.Native.Dns.GetPeerPoolMap()

		if err != nil {
			PrintErrorMsg("Get all dns register info err:%s\n", err)
			return nil
		}

		if _, ok := m.PeerPoolMap[""]; ok {
			delete(m.PeerPoolMap, "")
		}

		for _, item := range m.PeerPoolMap {
			PrintInfoMsg("PeerPubkey: %s\n", item.PeerPubkey)
			PrintInfoMsg("WalletAddress: %s\n", item.WalletAddress.ToBase58())
			PrintInfoMsg("Status: %d\n", item.Status)
			PrintInfoMsg("InitPos: %d\n", item.TotalInitPos)
			PrintInfoMsg("\n")
		}
	} else {
		peerPubkey := ctx.String(flags.GetFlagName(flags.PeerPubkeyFlag))
		item, err := endpoint.Dsp.Chain.Native.Dns.GetPeerPoolItem(peerPubkey)
		if err != nil {
			PrintErrorMsg("Get dns register info err:%s\n", err)
			return nil
		}

		PrintInfoMsg("PeerPubkey: %s\n", item.PeerPubkey)
		PrintInfoMsg("WalletAddress: %s\n", item.WalletAddress.ToBase58())
		PrintInfoMsg("Status: %d\n", item.Status)
		PrintInfoMsg("TotalInitPos: %d\n", item.TotalInitPos)
	}

	return nil
}

func getHostInfo(ctx *cli.Context) error {
	endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	DnsAllFlag := ctx.Bool(flags.GetFlagName(flags.DnsAllFlag))

	if DnsAllFlag {
		infos, err := endpoint.Dsp.Chain.Native.Dns.GetAllDnsNodes()
		if err != nil {
			PrintErrorMsg("Get all dns host info err:%s\n", err)
			return nil
		}

		for k, v := range infos {
			PrintInfoMsg("Pubkey:%s\n", k)
			PrintInfoMsg("wallet:%s\n", v.WalletAddr.ToBase58())
			PrintInfoMsg("ip:%s\n", v.IP)
			PrintInfoMsg("port:%s\n", v.Port)
			PrintInfoMsg("\n")
		}
	} else {
		var addr ccom.Address
		var err error

		walletAddr := ctx.String(flags.GetFlagName(flags.DnsWalletFlag))
		if walletAddr != "" {
			addr, err = ccom.AddressFromBase58(walletAddr)
			if err != nil {
				PrintErrorMsg("Get dns host info err:%s\n", err)
				return nil
			}
		}

		info, err := endpoint.Dsp.Chain.Native.Dns.GetDnsNodeByAddr(addr)
		if err != nil {
			PrintErrorMsg("Get dns host info err:%s\n", err)
			return nil
		}

		PrintInfoMsg("Pubkey:%s\n", info.PeerPubKey)
		PrintInfoMsg("Wallet:%s\n", info.WalletAddr.ToBase58())
		PrintInfoMsg("Ip:%v\n", info.IP)
		PrintInfoMsg("Port:%v\n", info.Port)
	}

	return nil
}
