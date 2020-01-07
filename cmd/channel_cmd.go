package cmd

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	eUtils "github.com/saveio/edge/utils"
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
				flags.AmountStrFlag,
			},
			Description: "Open a payment channel with partner",
		},
		{
			Action:    openAllDNShannel,
			Name:      "opentoalldns",
			Usage:     "Open payment channels to all dns",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.AmountStrFlag,
			},
			Description: "Open a payment channel with partner",
		},
		{
			Action:    closeChannel,
			Name:      "close",
			Usage:     "Close a payment channel",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PartnerAddressFlag,
			},
			Description: "Close a payment channel with partner",
		},
		{
			Action:      closeAllChannel,
			Name:        "closeall",
			Usage:       "Close all payment channels",
			ArgsUsage:   " ",
			Flags:       []cli.Flag{},
			Description: "Close a payment channel with partner",
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
			Name:      "mediatransfer",
			Usage:     "Make payment through channel",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.TargetAddressFlag,
				flags.AmountStrFlag,
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
	amount := ctx.String(flags.GetFlagName(flags.AmountStrFlag))
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	pwdHash := eUtils.Sha256HexStr(string(pwd))
	ret, err := utils.OpenPaymentChannel(partnerAddr, pwdHash, amount)
	if err != nil {
		return err
	}
	PrintJsonObject(ret)
	return nil
}

func openAllDNShannel(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	amount := ctx.String(flags.GetFlagName(flags.AmountStrFlag))
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	pwdHash := eUtils.Sha256HexStr(string(pwd))
	ret, err := utils.OpenAllDNSPaymentChannel(pwdHash, amount)
	if err != nil {
		return err
	}
	PrintJsonObject(ret)
	return nil
}

func closeChannel(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	pwdHash := eUtils.Sha256HexStr(string(pwd))
	err = utils.ClosePaymentChannel(partnerAddr, pwdHash)
	if err != nil {
		return err
	}
	PrintInfoMsg("close channel success")
	return nil
}

func closeAllChannel(ctx *cli.Context) error {
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	pwdHash := eUtils.Sha256HexStr(string(pwd))
	err = utils.CloseAllPaymentChannel(pwdHash)
	if err != nil {
		return err
	}
	PrintInfoMsg("close channel success")
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
	pwdHash := eUtils.Sha256HexStr(string(pwd))
	_, err = utils.DepositToChannel(partnerAddr, totalDeposit, pwdHash)
	if err != nil {
		PrintErrorMsg("%s", err)
		return err
	}
	PrintInfoMsg("deposit to channel success. use <channel list> to get infos")
	return nil
}

func transferToSomebody(ctx *cli.Context) error {
	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	targetAddr := ctx.String(flags.GetFlagName(flags.TargetAddressFlag))
	amount := ctx.String(flags.GetFlagName(flags.AmountStrFlag))
	paymentID := ctx.Uint64(flags.GetFlagName(flags.PaymentIDFlag))
	if paymentID == 0 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		paymentID = uint64(r.Int31())
	}
	err := utils.MediaTransfer(fmt.Sprintf("%d", paymentID), amount, targetAddr)
	if err != nil {
		return err
	}
	PrintInfoMsg("MediaTransfer success")
	return nil
}

func queryChannelDeposit(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	partnerAddr := ctx.String(flags.GetFlagName(flags.PartnerAddressFlag))
	amount, err := utils.QuerySpecialChannelDeposit(partnerAddr)
	if err != nil {
		PrintErrorMsg("%s", err)
		return err
	}
	PrintJsonData(amount)
	return nil
}

func withdrawChannel(ctx *cli.Context) error {
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
	pwdHash := eUtils.Sha256HexStr(string(pwd))
	_, err = utils.WithdrawChannel(partnerAddr, totalDeposit, pwdHash)
	if err != nil {
		PrintErrorMsg("%s", err)
		return err
	}
	PrintInfoMsg("withdraw to channel success. use <channel list> to get infos")
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
