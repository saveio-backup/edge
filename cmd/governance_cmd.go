package cmd

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/password"

	"github.com/urfave/cli"
)

var GovernanceCommand = cli.Command{
	Name:   "governance",
	Usage:  "Manage candidate",
	Before: utils.BeforeFunc,
	Subcommands: []cli.Command{
		{
			Action:    registerCandidate,
			Name:      "register",
			Usage:     "Register consensus candidate",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PeerPubkeyFlag,
				flags.InitDepositFlag,
			},
			Description: "Request register as consensus candidate",
		},
		{
			Action:    unregisterCandidate,
			Name:      "unregister",
			Usage:     "Cancel previous register request",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PeerPubkeyFlag,
			},
			Description: "Cancel previous register request",
		},
		{
			Action:    withdraw,
			Name:      "withdraw",
			Usage:     "Withdraw deposit",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PeerPubkeyListFlag,
				flags.WithdrawListFlag,
			},
			Description: "Withdraw deposit ONT",
		},
		{
			Action:    quitNode,
			Name:      "quit",
			Usage:     "Quit concensus",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PeerPubkeyFlag,
			},
			Description: "Quit concensus, deposit will unfreeze",
		},
		{
			Action:    addInitPos,
			Name:      "addInitPos",
			Usage:     "Increase init deposit",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PeerPubkeyFlag,
				flags.DeltaDepositFlag,
			},
			Description: "Increase init deposit",
		},
		{
			Action:    reduceInitPos,
			Name:      "reduceInitPos",
			Usage:     "Reduce init deposit",
			ArgsUsage: " ",
			Flags: []cli.Flag{
				flags.PeerPubkeyFlag,
				flags.DeltaDepositFlag,
			},
			Description: "Reduce init deposit",
		},
	},
	Flags: []cli.Flag{
		flags.DspRestAddrFlag,
	},
	Description: ` You can use the ./dsp governance --help command to view help information.`,
}

func registerCandidate(ctx *cli.Context) error {
	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	password := string(pwd)

	endpoint, err := dsp.Init(config.WalletDatFilePath(), password)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	peerPubkey := ctx.String(flags.GetFlagName(flags.PeerPubkeyFlag))
	initDeposit := ctx.Uint64(flags.GetFlagName(flags.InitDepositFlag))

	tx, err := endpoint.Dsp.Chain.Native.Governance.RegisterCandidate(peerPubkey, initDeposit)
	if err != nil {
		PrintErrorMsg("Register candidate err:%s\n", err)
		return nil
	}
	PrintInfoMsg("RegisterCandidate Success")
	PrintInfoMsg("tx :%s\n", hex.EncodeToString(common.ToArrayReverse(tx)))

	return nil
}

func unregisterCandidate(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	password := string(pwd)
	endpoint, err := dsp.Init(config.WalletDatFilePath(), password)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	peerPubkey := ctx.String(flags.GetFlagName(flags.PeerPubkeyFlag))

	tx, err := endpoint.Dsp.Chain.Native.Governance.UnRegisterCandidate(peerPubkey)
	if err != nil {
		PrintErrorMsg("Unregister candidate err:%s\n", err)
		return nil
	}
	PrintInfoMsg("UnregisterCandidate Success")
	PrintInfoMsg("tx :%s\n", hex.EncodeToString(common.ToArrayReverse(tx)))

	return nil
}

func withdraw(ctx *cli.Context) error {
	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	password := string(pwd)
	endpoint, err := dsp.Init(config.WalletDatFilePath(), password)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	pkstr := strings.TrimSpace(strings.Trim(ctx.String(flags.GetFlagName(flags.PeerPubkeyListFlag)), ","))
	pks := strings.Split(pkstr, ",")
	peerPubkeyList := make([]string, 0, len(pks))
	for _, pk := range pks {
		pk := strings.TrimSpace(pk)
		if pk == "" {
			continue
		}
		peerPubkeyList = append(peerPubkeyList, pk)
	}

	amountStr := strings.TrimSpace(strings.Trim(ctx.String(flags.GetFlagName(flags.WithdrawListFlag)), ","))
	amounts := strings.Split(amountStr, ",")
	withdrawList := make([]uint64, 0, len(amounts))
	for _, amount := range amounts {
		amount := strings.TrimSpace(amount)
		v, err := strconv.ParseUint(amount, 10, 64)
		if err != nil {
			PrintErrorMsg("parse withdraw value err:%s\n", err)
			return err
		}

		withdrawList = append(withdrawList, v)
	}

	tx, err := endpoint.Dsp.Chain.Native.Governance.Withdraw(peerPubkeyList, withdrawList)
	if err != nil {
		PrintErrorMsg("Withdraw err:%s\n", err)
		return nil
	}

	PrintInfoMsg("Withdraw Success")
	PrintInfoMsg("tx :%s\n", hex.EncodeToString(common.ToArrayReverse(tx)))

	return nil
}

func quitNode(ctx *cli.Context) error {
	if ctx.NumFlags() < 1 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	password := string(pwd)
	endpoint, err := dsp.Init(config.WalletDatFilePath(), password)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	peerPubkey := ctx.String(flags.GetFlagName(flags.PeerPubkeyFlag))

	tx, err := endpoint.Dsp.Chain.Native.Governance.QuitNode(peerPubkey)
	if err != nil {
		PrintErrorMsg("Quit candidate err:%s\n", err)
		return nil
	}
	PrintInfoMsg("Quit candidate Success")
	PrintInfoMsg("tx :%s\n", hex.EncodeToString(common.ToArrayReverse(tx)))

	return nil
}

func addInitPos(ctx *cli.Context) error {
	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	password := string(pwd)
	endpoint, err := dsp.Init(config.WalletDatFilePath(), password)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	peerPubkey := ctx.String(flags.GetFlagName(flags.PeerPubkeyFlag))
	deltaDeposit := ctx.Uint64(flags.GetFlagName(flags.DeltaDepositFlag))

	tx, err := endpoint.Dsp.Chain.Native.Governance.AddInitPos(peerPubkey, deltaDeposit)
	if err != nil {
		PrintErrorMsg("Add init deposit err:%s\n", err)
		return nil
	}
	PrintInfoMsg("Add init deposit Success")
	PrintInfoMsg("tx :%s\n", hex.EncodeToString(common.ToArrayReverse(tx)))

	return nil
}

func reduceInitPos(ctx *cli.Context) error {
	if ctx.NumFlags() < 2 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	password := string(pwd)
	endpoint, err := dsp.Init(config.WalletDatFilePath(), password)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	peerPubkey := ctx.String(flags.GetFlagName(flags.PeerPubkeyFlag))
	deltaDeposit := ctx.Uint64(flags.GetFlagName(flags.DeltaDepositFlag))

	tx, err := endpoint.Dsp.Chain.Native.Governance.ReduceInitPos(peerPubkey, deltaDeposit)
	if err != nil {
		PrintErrorMsg("Reduce init deposit err:%s\n", err)
		return nil
	}
	PrintInfoMsg("Reduce init deposit Success")
	PrintInfoMsg("tx :%s\n", hex.EncodeToString(common.ToArrayReverse(tx)))

	return nil
}
