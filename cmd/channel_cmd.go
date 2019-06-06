package cmd

import (
	"fmt"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp"
	sdk "github.com/saveio/themis-go-sdk/utils"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/password"

	"github.com/urfave/cli"
)

const (
	OPEN_CHANNEL          = "/api/v1/channel/open/"
	DEPOSIT_CHANNEL       = "/api/v1/channel/deposit/"
	TRANSFER_BY_CHANNEL   = "/api/v1/channel/transfer/"
	QUERY_CHANNEL_DEPOSIT = "/api/v1/channel/query/deposit/"
	QUERY_CHANNEL         = "/api/v1/channel/query/detail/"
	QUERY_CHANNEL_BY_ID   = "/api/v1/channel/query/id/"
)

var ChannelCommand = cli.Command{
	Name:   "channel",
	Usage:  "Manage channel",
	Before: utils.BeforeFunc,
	Subcommands: []cli.Command{
		{
			Action:      channelInitProgress,
			Name:        "initprogress",
			Usage:       "Get channel init progress",
			ArgsUsage:   " ",
			Flags:       []cli.Flag{},
			Description: "Get channel init progress",
		},
		{
			Action:      listAllChannels,
			Name:        "list",
			Usage:       "Get all channels",
			ArgsUsage:   " ",
			Flags:       []cli.Flag{},
			Description: "Get all channels",
		},
		{
			Action:    openChannel,
			Name:      "open",
			Usage:     "Open a payment channel",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
			},
			Description: "Open a payment channel with partner",
		},
		{
			Action:    depositToChannel,
			Name:      "deposit",
			Usage:     "Deposit token to channel",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
				flags.AmountStrFlag,
			},
			Description: "Deposit token to channel with specified partner",
		},
		{
			Action:    transferToSomebody,
			Name:      "transfer",
			Usage:     "Make payment through channel",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.TargetAddressFlag,
				flags.AmountFlag,
				flags.PaymentIDFlag,
			},
			Description: "Transfer some token from owner to target with specified payment ID",
		},
		{
			Action:    queryChannelDeposit,
			Name:      "query",
			Usage:     "Query channel deposit",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
			},
			Description: "Query deposit of channel which belong to owner and partner",
		},
		{
			Action:    withdrawChannel,
			Name:      "withdraw",
			Usage:     "Withdraw channel deposit",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
				flags.AmountFlag,
			},
			Description: "Withdraw deposit of channel which belong to owner and partner",
		},
		{
			Action:    cooperativeSettle,
			Name:      "cooperativeSettle",
			Usage:     "Cooperative settle",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
			},
			Description: "settle cooperatively of channel which belong to owner and partner",
		},
	},
	Flags: []cli.Flag{
		flags.DspRestAddrFlag,
	},
	Description: ` You can use the ./dsp channel --help command to view help information.`,
}

func openChannel(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
	ret, err := utils.OpenPaymentChannel(partnerAddr)
	if err != nil {
		return err
	}
	PrintJsonObject(ret)
	return nil
}

//[TODO] All channel operation may need call rest server!
func openChannelExt(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	err = dsp.StartDspNode(endpoint, false, false, true)
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}

	// partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))

	// id, err := endpoint.OpenPaymentChannel(partnerAddr)
	// if err != nil {
	// 	PrintErrorMsg("open channel err: %s", err)
	// 	return err
	// }

	// PrintInfoMsg("Open channel finished. Channel Id:%d", id)
	return nil
}

func depositToChannel(ctx *cli.Context) error {
	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
	totalDeposit := ctx.String(flags.GetFlagName(flags.AmountStrFlag))
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	_, err = utils.DepositToChannel(partnerAddr, totalDeposit, string(pwd))
	if err != nil {
		PrintErrorMsg("%s", err)
		return err
	}
	PrintInfoMsg("deposit to channel success. use <channel list> to get infos")
	return nil
}

func transferToSomebody(ctx *cli.Context) error {
	if ctx.NumFlags() < 3 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	targetAddr := ctx.String(flags.GetFlagName(flags.TargetAddressFlag))
	amount := ctx.Uint64(flags.GetFlagName(flags.AmountFlag))
	paymentID := ctx.Uint(flags.GetFlagName(flags.PaymentIDFlag))

	if utils.CallDspRest() {
		reqPath := fmt.Sprintf("%s%s/%d/%d", TRANSFER_BY_CHANNEL, targetAddr, amount, paymentID)
		_, err := utils.SendRestGetRequest(reqPath)
		if err != nil {
			PrintErrorMsg("transfer to %s err: %s", targetAddr, err)
			return err
		}

		PrintInfoMsg("transfer to %s successed", targetAddr)
	} else {
		endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
		if err != nil {
			PrintErrorMsg("init dsp err:%s\n", err)
			return err
		}

		err = dsp.StartDspNode(endpoint, false, false, true)
		if err != nil {
			PrintErrorMsg("start dsp err:%s\n", err)
			return err
		}
		// err = endpoint.MediaTransfer(int32(paymentID), amount, targetAddr)
		// if err != nil {
		// 	PrintErrorMsg("Transfer to %s failed. err: %s", targetAddr, err)
		// 	return err
		// } else {
		// 	PrintInfoMsg("Transfer to %s successed.", targetAddr)
		// }
	}

	return nil
}

func queryChannelDeposit(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))

	_, err := common.AddressFromBase58(partnerAddr)
	if err != nil {
		PrintErrorMsg("partner address %s is not valid", partnerAddr)
		return err
	}

	if utils.CallDspRest() {
		reqPath := fmt.Sprintf("%s%s", QUERY_CHANNEL_DEPOSIT, partnerAddr)
		data, err := utils.SendRestGetRequest(reqPath)
		if err != nil {
			PrintErrorMsg("query channel deposit err: %s", err)
			return err
		}

		selfAmount, err := sdk.GetUint64(data)
		if err != nil {
			PrintErrorMsg("query channel deposit err: %s", err)
			return err
		}
		PrintInfoMsg("Our amount: %v", selfAmount)
	} else {
		endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
		if err != nil {
			PrintErrorMsg("init dsp err:%s\n", err)
			return err
		}

		err = dsp.StartDspNode(endpoint, false, false, true)
		if err != nil {
			PrintErrorMsg("start dsp err:%s\n", err)
			return err
		}

		// selfAmount, err := endpoint.QuerySpecialChannelDeposit(partnerAddr)
		// if err != nil {
		// 	PrintErrorMsg("query channel deposit err: %s", err)
		// 	return err
		// }
		// PrintInfoMsg("Our amount: %v", selfAmount)
	}
	return nil
}

func withdrawChannel(ctx *cli.Context) error {
	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	// partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
	// amount := ctx.Uint64(flags.GetFlagName(flags.AmountFlag))
	// endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	// if err != nil {
	// 	PrintErrorMsg("init dsp err:%s\n", err)
	// 	return err
	// }
	// err = dsp.StartDspNode(endpoint, false, false, true)
	// if err != nil {
	// 	PrintErrorMsg("start dsp err:%s\n", err)
	// 	return err
	// }

	// err = endpoint.ChannelWithdraw(partnerAddr, amount)
	// if err != nil {
	// 	PrintErrorMsg("withdraw channel err: %s", err)
	// 	return err
	// }
	// PrintInfoMsg("withdraw channel amount: %v success", amount)
	return nil
}

func cooperativeSettle(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	// partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
	// endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	// if err != nil {
	// 	PrintErrorMsg("init dsp err:%s\n", err)
	// 	return err
	// }
	// err = dsp.StartDspNode(endpoint, false, false, true)
	// if err != nil {
	// 	PrintErrorMsg("start dsp err:%s\n", err)
	// 	return err
	// }

	// err = endpoint.ChannelCooperativeSettle(partnerAddr)
	// if err != nil {
	// 	PrintErrorMsg("cooperative settle channel err: %s", err)
	// 	return err
	// }
	// PrintInfoMsg("Cooperative settle success")
	return nil
}

func channelInitProgress(ctx *cli.Context) error {
	progress, err := utils.GetFilterBlockProgress()
	if err != nil {
		return err
	}
	PrintJsonData(progress)
	return nil
}

func listAllChannels(ctx *cli.Context) error {
	lists, err := utils.GetAllChannels()
	if err != nil {
		return err
	}
	PrintJsonData(lists)
	return nil
}
