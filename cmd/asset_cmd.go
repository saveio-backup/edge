package cmd

import (
	"fmt"
	"strconv"
	"strings"

	cutils "github.com/saveio/edge/cmd/utils"
	"github.com/saveio/themis/account"
	cmdcom "github.com/saveio/themis/cmd/common"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/config"

	"github.com/urfave/cli"
)

var AssetCommand = cli.Command{
	Name:        "asset",
	Usage:       "Handle assets",
	Description: "Asset management commands can check account balance, ONT/ONG transfers, extract ONGs, and view unbound ONGs, and so on.",
	Subcommands: []cli.Command{
		{
			Action:      transfer,
			Name:        "transfer",
			Usage:       "Transfer ont or ong to another account",
			ArgsUsage:   " ",
			Description: "Transfer ont or ong to another account. If from address does not specified, using default account",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
				utils.TransactionGasPriceFlag,
				utils.TransactionGasLimitFlag,
				utils.TransactionAssetFlag,
				utils.TransactionFromFlag,
				utils.TransactionToFlag,
				utils.TransactionAmountFlag,
				utils.ForceSendTxFlag,
				utils.WalletFileFlag,
			},
		},
		{
			Action:    approve,
			Name:      "approve",
			ArgsUsage: " ",
			Usage:     "Approve another user can transfer asset",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
				utils.TransactionGasPriceFlag,
				utils.TransactionGasLimitFlag,
				utils.ApproveAssetFlag,
				utils.ApproveAssetFromFlag,
				utils.ApproveAssetToFlag,
				utils.ApproveAmountFlag,
				utils.WalletFileFlag,
			},
		},
		{
			Action:    transferFrom,
			Name:      "transferfrom",
			ArgsUsage: " ",
			Usage:     "Using to transfer asset after approve",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
				utils.TransactionGasPriceFlag,
				utils.TransactionGasLimitFlag,
				utils.ApproveAssetFlag,
				utils.TransferFromSenderFlag,
				utils.ApproveAssetFromFlag,
				utils.ApproveAssetToFlag,
				utils.TransferFromAmountFlag,
				utils.ForceSendTxFlag,
				utils.WalletFileFlag,
			},
		},
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
		{
			Action: getAllowance,
			Name:   "allowance",
			Usage:  "Show approve balance of ont or ong of specified account",
			Flags: []cli.Flag{
				utils.RPCPortFlag,
				utils.ApproveAssetFlag,
				utils.ApproveAssetFromFlag,
				utils.ApproveAssetToFlag,
				utils.WalletFileFlag,
			},
		},
	},
}

func transfer(ctx *cli.Context) error {
	if !ctx.IsSet(utils.GetFlagName(utils.TransactionToFlag)) ||
		!ctx.IsSet(utils.GetFlagName(utils.TransactionFromFlag)) ||
		!ctx.IsSet(utils.GetFlagName(utils.TransactionAmountFlag)) {
		PrintErrorMsg("Missing %s %s or %s argument.", utils.TransactionToFlag.Name, utils.TransactionFromFlag.Name, utils.TransactionAmountFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	asset := ctx.String(utils.GetFlagName(utils.TransactionAssetFlag))
	if asset == "" {
		asset = utils.ASSET_USDT
	}
	from := ctx.String(utils.TransactionFromFlag.Name)
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return err
	}
	to := ctx.String(utils.TransactionToFlag.Name)
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return err
	}

	var amount uint64
	amountStr := ctx.String(utils.TransactionAmountFlag.Name)
	switch strings.ToLower(asset) {
	case "usdt":
		amount = utils.ParseUsdt(amountStr)
		amountStr = utils.FormatUsdt(amount)

	default:
		return fmt.Errorf("unsupport asset:%s", asset)
	}

	err = utils.CheckAssetAmount(asset, amount)
	if err != nil {
		return err
	}

	force := ctx.Bool(utils.GetFlagName(utils.ForceSendTxFlag))
	if !force {
		balance, err := cutils.GetAccountBalance(fromAddr, asset)
		if err != nil {
			return err
		}
		if balance < amount {
			PrintErrorMsg("Account:%s balance not enough.", fromAddr)
			PrintInfoMsg("\nTip:")
			PrintInfoMsg("  If you want to send transaction compulsively, please using %s flag.", utils.GetFlagName(utils.ForceSendTxFlag))
			return nil
		}
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)
	networkId, err := cutils.GetNetworkId()
	if err != nil {
		return err
	}
	if networkId == config.NETWORK_ID_SOLO_NET {
		gasPrice = 0
	}

	var signer *account.Account
	signer, err = cmdcom.GetAccount(ctx, fromAddr)
	if err != nil {
		return err
	}
	txHash, err := cutils.Transfer(gasPrice, gasLimit, signer, asset, fromAddr, toAddr, amount)
	if err != nil {
		return fmt.Errorf("transfer error:%s", err)
	}
	PrintInfoMsg("Transfer %s", strings.ToUpper(asset))
	PrintInfoMsg("  From:%s", fromAddr)
	PrintInfoMsg("  To:%s", toAddr)
	PrintInfoMsg("  Amount:%s", amountStr)
	PrintInfoMsg("  TxHash:%s", txHash)
	PrintInfoMsg("\nTip:")
	PrintInfoMsg("  Using './edge info status %s' to query transaction status.", txHash)
	return nil
}

func getBalance(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing account argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	addrArg := ctx.Args().First()
	accAddr, err := cmdcom.ParseAddress(addrArg, ctx)
	if err != nil {
		return err
	}
	balance, err := cutils.GetBalance(accAddr)
	if err != nil {
		return err
	}

	PrintInfoMsg("BalanceOf:%s", accAddr)
	// PrintInfoMsg("  ONT:%s", balance.Usdt)
	usdt, err := strconv.ParseUint(balance.Usdt, 10, 64)
	if err != nil {
		return err
	}
	PrintInfoMsg("  USDT:%s", utils.FormatUsdt(usdt))
	return nil
}

func getAllowance(ctx *cli.Context) error {
	from := ctx.String(utils.GetFlagName(utils.ApproveAssetFromFlag))
	to := ctx.String(utils.GetFlagName(utils.ApproveAssetToFlag))
	if from == "" || to == "" {
		PrintErrorMsg("Missing %s or %s argument.", utils.ApproveAssetFromFlag.Name, utils.ApproveAssetToFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	asset := ctx.String(utils.GetFlagName(utils.ApproveAssetFlag))
	if asset == "" {
		asset = utils.ASSET_USDT
	}
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return err
	}
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return err
	}
	balanceStr, err := cutils.GetAllowance(asset, fromAddr, toAddr)
	if err != nil {
		return err
	}
	switch strings.ToLower(asset) {
	case "usdt":
	default:
		return fmt.Errorf("unsupport asset:%s", asset)
	}
	PrintInfoMsg("Allowance:%s", asset)
	PrintInfoMsg("  From:%s", fromAddr)
	PrintInfoMsg("  To:%s", toAddr)
	PrintInfoMsg("  Balance:%s", balanceStr)
	return nil
}

func approve(ctx *cli.Context) error {
	asset := ctx.String(utils.GetFlagName(utils.ApproveAssetFlag))
	from := ctx.String(utils.GetFlagName(utils.ApproveAssetFromFlag))
	to := ctx.String(utils.GetFlagName(utils.ApproveAssetToFlag))
	amountStr := ctx.String(utils.GetFlagName(utils.ApproveAmountFlag))
	if asset == "" ||
		from == "" ||
		to == "" ||
		amountStr == "" {
		PrintErrorMsg("Missing %s %s %s or %s argument.", utils.ApproveAssetFlag.Name, utils.ApproveAssetFromFlag.Name, utils.ApproveAssetToFlag.Name, utils.ApproveAmountFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return err
	}
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return err
	}
	var amount uint64
	switch strings.ToLower(asset) {
	case "usdt":
		amount = utils.ParseUsdt(amountStr)
		amountStr = utils.FormatUsdt(amount)
	default:
		return fmt.Errorf("unsupport asset:%s", asset)
	}

	err = utils.CheckAssetAmount(asset, amount)
	if err != nil {
		return err
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)
	networkId, err := cutils.GetNetworkId()
	if err != nil {
		return err
	}
	if networkId == config.NETWORK_ID_SOLO_NET {
		gasPrice = 0
	}

	var signer *account.Account
	signer, err = cmdcom.GetAccount(ctx, fromAddr)
	if err != nil {
		return err
	}

	txHash, err := cutils.Approve(gasPrice, gasLimit, signer, asset, fromAddr, toAddr, amount)
	if err != nil {
		return fmt.Errorf("approve error:%s", err)
	}

	PrintInfoMsg("Approve:")
	PrintInfoMsg("  Asset:%s", asset)
	PrintInfoMsg("  From:%s", fromAddr)
	PrintInfoMsg("  To:%s", toAddr)
	PrintInfoMsg("  Amount:%s", amountStr)
	PrintInfoMsg("  TxHash:%s", txHash)
	PrintInfoMsg("\nTip:")
	PrintInfoMsg("  Using './edge info status %s' to query transaction status.", txHash)
	return nil
}

func transferFrom(ctx *cli.Context) error {
	asset := ctx.String(utils.GetFlagName(utils.ApproveAssetFlag))
	from := ctx.String(utils.GetFlagName(utils.ApproveAssetFromFlag))
	to := ctx.String(utils.GetFlagName(utils.ApproveAssetToFlag))
	amountStr := ctx.String(utils.GetFlagName(utils.TransferFromAmountFlag))
	if asset == "" ||
		from == "" ||
		to == "" ||
		amountStr == "" {
		PrintErrorMsg("Missing %s %s %s or %s argument.", utils.ApproveAssetFlag.Name, utils.ApproveAssetFromFlag.Name, utils.ApproveAssetToFlag.Name, utils.TransferFromAmountFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return err
	}
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return err
	}

	var sendAddr string
	sender := ctx.String(utils.GetFlagName(utils.TransferFromSenderFlag))
	if sender == "" {
		sendAddr = toAddr
	} else {
		sendAddr, err = cmdcom.ParseAddress(sender, ctx)
		if err != nil {
			return err
		}
	}

	var signer *account.Account
	signer, err = cmdcom.GetAccount(ctx, sendAddr)
	if err != nil {
		return err
	}

	var amount uint64
	switch strings.ToLower(asset) {
	case "usdt":
		amount = utils.ParseUsdt(amountStr)
		amountStr = utils.FormatUsdt(amount)
	default:
		return fmt.Errorf("unsupport asset:%s", asset)
	}

	err = utils.CheckAssetAmount(asset, amount)
	if err != nil {
		return err
	}

	force := ctx.Bool(utils.GetFlagName(utils.ForceSendTxFlag))
	if !force {
		balance, err := cutils.GetAccountBalance(fromAddr, asset)
		if err != nil {
			return err
		}
		if balance < amount {
			PrintErrorMsg("Account:%s balance not enough.", fromAddr)
			PrintInfoMsg("\nTip:")
			PrintInfoMsg("  If you want to send transaction compulsively, please using %s flag.", utils.GetFlagName(utils.ForceSendTxFlag))
			return nil
		}
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)
	networkId, err := cutils.GetNetworkId()
	if err != nil {
		return err
	}
	if networkId == config.NETWORK_ID_SOLO_NET {
		gasPrice = 0
	}

	txHash, err := cutils.TransferFrom(gasPrice, gasLimit, signer, asset, sendAddr, fromAddr, toAddr, amount)
	if err != nil {
		return err
	}

	PrintInfoMsg("Transfer from:")
	PrintInfoMsg("  Asset:%s", asset)
	PrintInfoMsg("  Sender:%s", sendAddr)
	PrintInfoMsg("  From:%s", fromAddr)
	PrintInfoMsg("  To:%s", toAddr)
	PrintInfoMsg("  Amount:%s", amountStr)
	PrintInfoMsg("  TxHash:%s", txHash)
	PrintInfoMsg("\nTip:")
	PrintInfoMsg("  Using './edge info status %s' to query transaction status.", txHash)
	return nil
}
