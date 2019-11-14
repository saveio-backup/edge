package cmd

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/saveio/edge/cmd/flags"
	cmdutil "github.com/saveio/edge/cmd/utils"
	eUtils "github.com/saveio/edge/utils"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/cmd/common"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/password"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/crypto/signature"

	"github.com/urfave/cli"
)

var (
	AccountCommand = cli.Command{
		Action:    cli.ShowSubcommandHelp,
		Name:      "account",
		Usage:     "Manage accounts",
		ArgsUsage: "[arguments...]",
		Description: `Wallet management commands can be used to add, view, modify, delete, import account, and so on.
You can use ./edge account --help command to view help information of wallet management command.`,
		Subcommands: []cli.Command{
			{
				Action:    accountCreate,
				Name:      "add",
				Usage:     "Add a new account",
				ArgsUsage: "[sub-command options]",
				Flags: []cli.Flag{
					utils.AccountQuantityFlag,
					utils.AccountTypeFlag,
					utils.AccountKeylenFlag,
					utils.AccountSigSchemeFlag,
					utils.AccountDefaultFlag,
					utils.AccountLabelFlag,
					utils.IdentityFlag,
					utils.WalletFileFlag,
				},
				Description: ` Add a new account to wallet.
   Edge support three type of key: ecdsa, sm2 and ed25519, and support 224、256、384、521 bits length of key in ecdsa, but only support 256 bits length of key in sm2 and ed25519.
   Edge support multiple signature scheme.
   For ECDSA support SHA224withECDSA、SHA256withECDSA、SHA384withECDSA、SHA512withEdDSA、SHA3-224withECDSA、SHA3-256withECDSA、SHA3-384withECDSA、SHA3-512withECDSA、RIPEMD160withECDSA;
   For SM2 support SM3withSM2, and for SHA512withEdDSA.
   -------------------------------------------------
      Key   |key-length(bits)|  signature-scheme
   ---------|----------------|----------------------
   1 ecdsa  |  1 P-224: 224  | 1 SHA224withECDSA
            |----------------|----------------------
            |  2 P-256: 256  | 2 SHA256withECDSA
            |----------------|----------------------
            |  3 P-384: 384  | 3 SHA384withECDSA
            |----------------|----------------------
            |  4 P-521: 521  | 4 SHA512withEdDSA
            |----------------|----------------------
            |                | 5 SHA3-224withECDSA
            |                |----------------------
            |                | 6 SHA3-256withECDSA
            |                |----------------------
            |                | 7 SHA3-384withECDSA
            |                |----------------------
            |                | 8 SHA3-512withECDSA
            |                |----------------------
            |                | 9 RIPEMD160withECDSA
   ---------|----------------|----------------------
   2 sm2    | sm2p256v1 256  | SM3withSM2
   ---------|----------------|----------------------
   3 ed25519|   25519 256    | SHA512withEdDSA
   -------------------------------------------------`,
			},
			{
				Action:    accountList,
				Name:      "list",
				Usage:     "List existing accounts",
				ArgsUsage: "[sub-command options] <label|address|index>",
				Flags: []cli.Flag{
					utils.WalletFileFlag,
					utils.AccountVerboseFlag,
				},
				Description: `List existing accounts. If specified in args, will list those account. If not specified in args, will list all accouns in wallet`,
			},
			{
				Action:    accountSet,
				Name:      "set",
				Usage:     "Modify an account",
				ArgsUsage: "[sub-command options] <label|address|index>",
				Flags: []cli.Flag{
					utils.AccountSetDefaultFlag,
					utils.WalletFileFlag,
					utils.AccountLabelFlag,
					utils.AccountChangePasswdFlag,
					utils.AccountSigSchemeFlag,
				},
				Description: `Modify settings for an account. Account is specified by address, label of index. Index start from 1. This can be showed by the 'list' command.`,
			},
			{
				Action:    accountDelete,
				Name:      "del",
				Usage:     "Delete an account",
				ArgsUsage: "[sub-command options] <address|label|index>",
				Flags: []cli.Flag{
					utils.WalletFileFlag,
				},
				Description: `Delete an account specified by address, label of index. Index start from 1. This can be showed by the 'list' command`,
			},
			{
				Action:    accountImport,
				Name:      "import",
				Usage:     "Import accounts of wallet to another",
				ArgsUsage: "[sub-command options]",
				Flags: []cli.Flag{
					flags.WalletFileFlag,
					flags.WalletPasswordFlag,
					flags.ImportOnlineWalletFlag,
				},
				Description: "Import accounts of wallet to another. If not specific accounts in args, all account in source will be import",
			},
			{
				Action:    accountCreateOnline,
				Name:      "create",
				Usage:     "Create online account to start edge",
				ArgsUsage: "[sub-command options]",
				Flags: []cli.Flag{
					flags.WalletLabelFlag,
					flags.WalletKeyTypeFlag,
					flags.WalletCurveFlag,
					flags.WalletSchemeFlag,
				},
			},
			{
				Action:    accountExport,
				Name:      "export",
				Usage:     "Export accounts to a specified wallet file",
				ArgsUsage: "[sub-command options] <filename>",
				Flags: []cli.Flag{
					flags.WalletExportTypeFlag,
				},
			},
			{
				Action:    currentAccount,
				Name:      "current",
				Usage:     "List current account",
				ArgsUsage: "[sub-command options]",
				Flags:     []cli.Flag{},
			},
			{
				Action:    logoutAccount,
				Name:      "logout",
				Usage:     "Logout current account",
				ArgsUsage: "[sub-command options]",
				Flags:     []cli.Flag{},
			},
		},
	}
)

func accountCreate(ctx *cli.Context) error {
	reader := bufio.NewReader(os.Stdin)
	optionType := ""
	optionCurve := ""
	optionScheme := ""

	optionDefault := ctx.IsSet(flags.GetFlagName(utils.AccountDefaultFlag))
	if !optionDefault {
		optionType = checkType(ctx, reader)
		optionCurve = checkCurve(ctx, reader, &optionType)
		optionScheme = checkScheme(ctx, reader, &optionType)
	} else {
		PrintInfoMsg("Use default setting '-t ecdsa -b 256 -s SHA256withECDSA'")
		PrintInfoMsg("	signature algorithm: %s", keyTypeMap[optionType].name)
		PrintInfoMsg("	curve: %s", curveMap[optionCurve].name)
		PrintInfoMsg("	signature scheme: %s", schemeMap[optionScheme].name)
	}
	optionFile := checkFileName(ctx)
	optionNumber := checkNumber(ctx)
	optionLabel := checkLabel(ctx)
	pass, _ := password.GetConfirmedPassword()
	keyType := keyTypeMap[optionType].code
	curve := curveMap[optionCurve].code
	scheme := schemeMap[optionScheme].code
	wallet, err := account.Open(optionFile)
	if err != nil {
		return fmt.Errorf("open wallet error:%s", err)
	}
	defer common.ClearPasswd(pass)
	if ctx.Bool(utils.IdentityFlag.Name) {
		// create ONT ID
		wd := wallet.GetWalletData()
		id, err := account.NewIdentity(optionLabel, keyType, curve, pass)
		if err != nil {
			return fmt.Errorf("create ONT ID error: %s", err)
		}
		wd.AddIdentity(id)
		err = wd.Save(optionFile)
		if err != nil {
			return fmt.Errorf("save to %s error: %s", optionFile, err)
		}
		PrintInfoMsg("ONT ID created:%s", id.ID)
		PrintInfoMsg("Bind public key:%s", id.Control[0].Public)
		return nil
	}
	for i := 0; i < optionNumber; i++ {
		label := optionLabel
		if label != "" && optionNumber > 1 {
			label = fmt.Sprintf("%s%d", label, i+1)
		}
		acc, err := wallet.NewAccount(label, keyType, curve, scheme, pass)
		if err != nil {
			return fmt.Errorf("new account error:%s", err)
		}
		PrintInfoMsg("Index:%d", wallet.GetAccountNum())
		PrintInfoMsg("Label:%s", label)
		PrintInfoMsg("Address:%s", acc.Address.ToBase58())
		PrintInfoMsg("Public key:%s", hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)))
		PrintInfoMsg("Signature scheme:%s", acc.SigScheme.Name())
	}

	PrintInfoMsg("Create account successfully.")
	return nil
}

func accountList(ctx *cli.Context) error {
	optionFile := checkFileName(ctx)
	wallet, err := account.Open(optionFile)
	if err != nil {
		return fmt.Errorf("open wallet:%s error:%s", optionFile, err)
	}
	accNum := wallet.GetAccountNum()
	if accNum == 0 {
		PrintInfoMsg("No account.")
		return nil
	}
	accList := make(map[string]string, ctx.NArg())
	for i := 0; i < ctx.NArg(); i++ {
		addr := ctx.Args().Get(i)
		accMeta := common.GetAccountMetadataMulti(wallet, addr)
		if accMeta == nil {
			PrintWarnMsg("Cannot find account by:%s in wallet:%s", addr, flags.GetFlagName(utils.WalletFileFlag))
			continue
		}
		accList[accMeta.Address] = ""
	}
	for i := 1; i <= accNum; i++ {
		accMeta := wallet.GetAccountMetadataByIndex(i)
		if accMeta == nil {
			continue
		}
		if len(accList) > 0 {
			_, ok := accList[accMeta.Address]
			if !ok {
				continue
			}
		}
		if !ctx.Bool(flags.GetFlagName(utils.AccountVerboseFlag)) {
			if accMeta.IsDefault {
				PrintInfoMsg("Index:%-4d Address:%s  Label:%s (default)", i, accMeta.Address, accMeta.Label)
			} else {
				PrintInfoMsg("Index:%-4d Address:%s  Label:%s", i, accMeta.Address, accMeta.Label)
			}
			continue
		}
		if accMeta.IsDefault {
			PrintInfoMsg("%v\t%v (default)", i, accMeta.Address)
		} else {
			PrintInfoMsg("%v\t%v", i, accMeta.Address)
		}
		PrintInfoMsg("	Label: %v", accMeta.Label)
		PrintInfoMsg("	Signature algorithm: %v", accMeta.KeyType)
		PrintInfoMsg("	Curve: %v", accMeta.Curve)
		PrintInfoMsg("	Key length: %v bits", len(accMeta.Key)*8)
		PrintInfoMsg("	Public key: %v", accMeta.PubKey)
		PrintInfoMsg("	Signature scheme: %v\n", accMeta.SigSch)
	}
	return nil
}

//set signature scheme for an account
func accountSet(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing account argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	address := ctx.Args().First()
	wallet, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	accMeta := common.GetAccountMetadataMulti(wallet, address)
	if accMeta == nil {
		return fmt.Errorf("cannot find account info by:%s", address)
	}
	address = accMeta.Address
	label := accMeta.Label
	if ctx.Bool(flags.GetFlagName(utils.AccountSetDefaultFlag)) {
		err = wallet.SetDefaultAccount(address)
		if err != nil {
			PrintErrorMsg("Set Label:%s Account:%s as default failed, %s", label, address, err)
		} else {
			PrintInfoMsg("Set Label:%s Account:%s as default successfully", label, address)
		}
	}
	if ctx.IsSet(flags.GetFlagName(utils.AccountLabelFlag)) {
		newLabel := ctx.String(flags.GetFlagName(utils.AccountLabelFlag))
		err = wallet.SetLabel(address, newLabel)
		if err != nil {
			PrintErrorMsg("Set Account:%s Label:%s to %s failed, %s", address, label, newLabel, err)
		} else {
			PrintInfoMsg("Set Account:%s Label:%s to %s successfully.", address, label, newLabel)
			label = newLabel
		}
	}
	if ctx.IsSet(flags.GetFlagName(utils.AccountSigSchemeFlag)) {
		find := false
		sigScheme := ctx.String(flags.GetFlagName(utils.AccountSigSchemeFlag))
		var sigSch signature.SignatureScheme
		for key, val := range schemeMap {
			if key == sigScheme {
				find = true
				sigSch = val.code
				break
			}
			if val.name == sigScheme {
				find = true
				sigSch = val.code
				break
			}
		}
		if find {
			err = wallet.ChangeSigScheme(address, sigSch)
			if err != nil {
				PrintErrorMsg("Set Label:%s Account:%s SigScheme to: %s failed, %s", accMeta.Label, accMeta.Address, sigSch.Name(), err)
			} else {
				PrintInfoMsg("Set Label:%s Account:%s SigScheme to: %s successfully.", accMeta.Label, accMeta.Address, sigSch.Name())
			}
		} else {
			PrintInfoMsg("%s is not a valid content for option -s", sigScheme)
		}
	}

	if ctx.Bool(flags.GetFlagName(utils.AccountChangePasswdFlag)) {
		passwd, err := common.GetPasswd(ctx)
		if err != nil {
			return err
		}
		defer common.ClearPasswd(passwd)
		PrintInfoMsg("Please input new password:")
		newPass, err := password.GetConfirmedPassword()
		if err != nil {
			return fmt.Errorf("input password error:%s", err)
		}
		err = wallet.ChangePassword(address, passwd, newPass)
		if err != nil {
			PrintErrorMsg("Change password label:%s account:%s failed, %s", accMeta.Label, address, err)
		} else {
			PrintInfoMsg("Change password label:%s account:%s successfully", accMeta.Label, address)
		}
	}
	return nil
}

//delete an account by index from 'list'
func accountDelete(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing account argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	address := ctx.Args().First()

	wallet, err := common.OpenWallet(ctx)
	if err != nil {
		return err
	}
	accMeta := common.GetAccountMetadataMulti(wallet, address)
	if accMeta == nil {
		return fmt.Errorf("cannot get account by address:%s", address)
	}
	passwd, err := common.GetPasswd(ctx)
	if err != nil {
		return err
	}
	defer common.ClearPasswd(passwd)
	_, err = wallet.DeleteAccount(accMeta.Address, passwd)
	if err != nil {
		PrintErrorMsg("Delete account label:%s address:%s failed, %s", accMeta.Label, accMeta.Address, err)
	} else {
		PrintInfoMsg("Delete account label:%s address:%s successfully.", accMeta.Label, accMeta.Address)
	}
	return nil
}

func accountImport(ctx *cli.Context) error {
	password := ctx.String(flags.GetFlagName(flags.WalletPasswordFlag))
	walletFile := ctx.String(flags.GetFlagName(flags.WalletFileFlag))
	forOnline := ctx.Bool(flags.GetFlagName(flags.ImportOnlineWalletFlag))
	if !forOnline {
		// TODO: support import offline
		return nil
	}
	wallet, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}
	acc, err := cmdutil.ImportWithWalletData(string(wallet), password)
	if err != nil {
		return err
	}
	PrintJsonObject(acc)
	return nil
}

func accountCreateOnline(ctx *cli.Context) error {
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	password := string(pwd)
	label := ctx.String(flags.GetFlagName(flags.WalletLabelFlag))
	keyType := ctx.String(flags.GetFlagName(flags.WalletKeyTypeFlag))
	curve := ctx.String(flags.GetFlagName(flags.WalletCurveFlag))
	scheme := ctx.String(flags.GetFlagName(flags.WalletSchemeFlag))
	acc, err := cmdutil.NewAccount(password, label, keyType, curve, scheme)
	if err != nil {
		return err
	}
	PrintJsonObject(acc)
	return nil
}

func accountExport(ctx *cli.Context) error {
	exportType := ctx.Int(flags.GetFlagName(flags.WalletExportTypeFlag))

	if exportType == 0 && ctx.NArg() <= 0 {
		PrintErrorMsg("Missing target file argument to export.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if exportType == 0 {
		target := ctx.Args().First()

		wal, err := cmdutil.ExportWalletFile()
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(target, []byte(wal), 0666)
		if err != nil {
			return err
		}
		PrintInfoMsg("Export wallet success.")
	} else {
		pwd, err := password.GetPassword()
		if err != nil {
			return err
		}
		pwdHash := eUtils.Sha256HexStr(string(pwd))
		ret, err := cmdutil.ExportPrivateKey(pwdHash)
		if err != nil {
			return err
		}
		PrintInfoMsg(ret)
	}

	return nil
}

func currentAccount(ctx *cli.Context) error {
	acc, err := cmdutil.GetCurrentAccount()
	if err != nil {
		return err
	}
	PrintJsonObject(acc)
	return nil
}

func logoutAccount(ctx *cli.Context) error {
	err := cmdutil.Logout()
	if err != nil {
		return err
	}
	PrintInfoMsg("Logout wallet success.")
	return nil
}
