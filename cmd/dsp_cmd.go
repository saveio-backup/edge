package cmd

import (
	"encoding/json"

	"github.com/saveio/dsp-go-sdk/common"
	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp"
	"github.com/saveio/themis/common/log"

	"github.com/urfave/cli"
)

var FileCommand = cli.Command{
	Name:  "file",
	Usage: "Upload/Download file using dsp",
	Subcommands: []cli.Command{
		{
			Action:    fileUpload,
			Name:      "upload",
			Usage:     "Upload file",
			ArgsUsage: "<hash>",
			Flags: []cli.Flag{
				flags.DspUploadFileNameFlag,
				flags.DspUploadFileDescFlag,
				flags.DspUploadChallengeRateFlag,
				flags.DspUploadChallengeTimesFlag,
				flags.DspUploadPrivilegeFlag,
				flags.DspUploadCopyNumFlag,
				flags.DspUploadEncryptFlag,
				flags.DspUploadEncryptPasswordFlag,
				flags.DspUploadUrlFlag,
			},
			Description: "Upload file",
		},
		{
			Action:    fileDownload,
			Name:      "download",
			Usage:     "Download file",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspFileHashFlag,
				flags.DspInorderFlag,
				flags.DspNofeeFlag,
				flags.DspDecryptPwdFlag,
				flags.DspMaxPeerCntFlag,
				flags.DspProgressEnableFlag,
			},
			Description: "Download file",
		},
		{
			Action:    fileDelete,
			Name:      "delete",
			Usage:     "Delete file",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspFileHashFlag,
				flags.DspDeleteLocalFlag,
			},
			Description: "Delete file",
		},
	},
	Description: `./dsp file --help command to view help information.`,
}

var NodeCommand = cli.Command{
	Name:  "node",
	Usage: "Display information about the dsp",
	Subcommands: []cli.Command{
		{
			Action:    registerNode,
			Name:      "register",
			Usage:     "Register node to themis",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspNodeAddrFlag,
				flags.DspVolumeFlag,
				flags.DspServiceTimeFlag,
			},
		},
		{
			Action:      unregisterNode,
			Name:        "unregister",
			Usage:       "Unregister node from themis",
			Description: "Unregister node",
		},
		{
			Action:    queryNode,
			Name:      "query",
			Usage:     "Query node info",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspWalletAddrFlag,
			},
			Description: "Query node info",
		},
		{
			Action:    updateNode,
			Name:      "update",
			Usage:     "Update node info",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspNodeAddrFlag,
				flags.DspVolumeFlag,
				flags.DspServiceTimeFlag,
			},
			Description: "Update node info",
		},
		{
			Action:      withDraw,
			Name:        "withdraw",
			Usage:       "Withdraw node profit",
			Description: "Withdraw node profit",
		},
	},
	Description: `./dsp node --help command to view help information.`,
}

//file command
func fileDownload(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspFileHashFlag)) {
		PrintErrorMsg("Missing file hash.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	log.InitLog(config.Parameters.BaseConfig.LogLevel, log.PATH, log.Stdout)
	endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	err = dsp.StartDspNode(endpoint, true, false, true)
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}

	hash := ctx.String(flags.GetFlagName(flags.DspFileHashFlag))
	inorder := ctx.Bool(flags.GetFlagName(flags.DspInorderFlag))
	pwd := ctx.String(flags.GetFlagName(flags.DspDecryptPwdFlag))
	nofee := ctx.Bool(flags.GetFlagName(flags.DspNofeeFlag))
	maxPeer := ctx.Uint64(flags.GetFlagName(flags.DspMaxPeerCntFlag))
	progressEnable := ctx.Bool(flags.GetFlagName(flags.DspProgressEnableFlag))

	if progressEnable {
		endpoint.Dsp.RegProgressChannel()
		go func() {
			stop := false
			for {
				v := <-endpoint.Dsp.ProgressChannel()
				for node, cnt := range v.Count {
					PrintInfoMsg("file:%s, hash:%s, total:%d, peer:%s, downloaded:%d, progress:%f", v.FileName, v.FileHash, v.Total, node, cnt, float64(cnt)/float64(v.Total))
					stop = (cnt == v.Total)
				}
				if stop {
					break
				}
			}
			endpoint.Dsp.CloseProgressChannel()
		}()
	}
	err = endpoint.Dsp.DownloadFile(hash, "", common.ASSET_USDT, inorder, pwd, nofee, int(maxPeer))
	if err != nil {
		PrintErrorMsg("download err %s\n", err)
		return err
	}
	//[TODO] need wait download task finish by RegProgressCh
	PrintInfoMsg("Download file:%s Successed", hash)
	return nil
}

//upload can be done without dsp daemon
func fileUpload(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspUploadFileNameFlag)) {
		PrintErrorMsg("Missing file name.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	log.InitLog(config.Parameters.BaseConfig.LogLevel, log.PATH, log.Stdout)

	endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	err = dsp.StartDspNode(endpoint, true, false, false)
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}

	fileName := ctx.String(flags.GetFlagName(flags.DspUploadFileNameFlag))
	fileDesc := ctx.String(flags.GetFlagName(flags.DspUploadFileDescFlag))
	rate := ctx.Uint64(flags.GetFlagName(flags.DspUploadChallengeRateFlag))
	challengeTimes := ctx.Uint64(flags.GetFlagName(flags.DspUploadChallengeTimesFlag))
	uploadPrivilege := ctx.Uint64(flags.GetFlagName(flags.DspUploadPrivilegeFlag))
	copyNum := ctx.Uint64(flags.GetFlagName(flags.DspUploadCopyNumFlag))
	encrypt := ctx.Bool(flags.GetFlagName(flags.DspUploadEncryptFlag))
	encryptPwd := ctx.String(flags.GetFlagName(flags.DspUploadEncryptPasswordFlag))
	uploadUrl := ctx.String(flags.GetFlagName(flags.DspUploadUrlFlag))
	regUrl, bindUrl := len(uploadUrl) > 0, len(uploadUrl) > 0

	opt := &common.UploadOption{
		FileDesc:        fileDesc,
		ProveInterval:   rate,
		ProveTimes:      uint32(challengeTimes),
		Privilege:       uint32(uploadPrivilege),
		CopyNum:         uint32(copyNum),
		Encrypt:         encrypt,
		EncryptPassword: encryptPwd,
		RegisterDns:     regUrl,
		BindDns:         bindUrl,
		DnsUrl:          uploadUrl,
	}

	ret, err := endpoint.Dsp.UploadFile(fileName, opt)
	if err != nil {
		PrintErrorMsg("upload file failed, err: %s", err)
		return err
	}
	retJson, _ := json.Marshal(ret)
	PrintInfoMsg("upload file success, result:\n %s", string(retJson))

	return nil
}

func fileDelete(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspFileHashFlag)) {
		PrintErrorMsg("Missing file hash.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	err = dsp.StartDspNode(endpoint, true, false, false)
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}
	hash := ctx.String(flags.GetFlagName(flags.DspFileHashFlag))
	isLocal := ctx.Bool(flags.GetFlagName(flags.DspDeleteLocalFlag))

	if isLocal {
		err = endpoint.Dsp.DeleteDownloadedFile(hash)
		if err != nil {
			PrintErrorMsg("delete file failed, err: %s", err)
		}
		PrintInfoMsg("delete file success")
		return nil
	}

	ret, err := endpoint.Dsp.DeleteUploadedFile(hash)
	if err != nil {
		PrintErrorMsg("delete file failed, err: %s", err)
	}
	PrintInfoMsg("delete file success, ret: %s", ret)

	return nil
}

//node command
func registerNode(ctx *cli.Context) error {
	if ctx.NumFlags() < 3 {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	err = dsp.StartDspNode(endpoint, false, false, false)
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}
	nodeAddr := ctx.String(flags.GetFlagName(flags.DspNodeAddrFlag))
	volume := ctx.Uint64(flags.GetFlagName(flags.DspVolumeFlag))
	serviceTime := ctx.Uint64(flags.GetFlagName(flags.DspServiceTimeFlag))

	tx, err := endpoint.Dsp.RegisterNode(nodeAddr, volume, serviceTime)
	if err != nil {
		PrintErrorMsg("register node err: %s", err)
		return err
	}
	PrintInfoMsg("register node success, tx: %s", tx)

	return nil
}

func unregisterNode(ctx *cli.Context) error {
	endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	err = dsp.StartDspNode(endpoint, false, false, false)
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}
	tx, err := endpoint.Dsp.UnregisterNode()
	if err != nil {
		PrintErrorMsg("unregister node err: %s", err)
		return err
	}
	PrintInfoMsg("unregister node success, tx: %s", tx)

	return nil
}

func queryNode(ctx *cli.Context) error {
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

	err = dsp.StartDspNode(endpoint, false, false, false)
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}
	walletAddr := ctx.String(flags.GetFlagName(flags.DspWalletAddrFlag))
	info, err := endpoint.Dsp.QueryNode(walletAddr)
	if err != nil {
		PrintErrorMsg("query node err %s", err)
		return err
	}
	PrintInfoMsg("node info pledge %d", info.Pledge)
	PrintInfoMsg("node info profit %d", info.Profit)
	PrintInfoMsg("node info volume %d", info.Volume)
	PrintInfoMsg("node info restvol %d", info.RestVol)
	PrintInfoMsg("node info service time %d", info.ServiceTime)
	PrintInfoMsg("node info wallet address %s", info.WalletAddr.ToBase58())
	PrintInfoMsg("node info node address %s", info.NodeAddr)

	return nil
}

func updateNode(ctx *cli.Context) error {
	endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	err = dsp.StartDspNode(endpoint, false, false, false)
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}
	nodeAddr := ctx.String(flags.GetFlagName(flags.DspNodeAddrFlag))
	volume := ctx.Uint64(flags.GetFlagName(flags.DspVolumeFlag))
	serviceTime := ctx.Uint64(flags.GetFlagName(flags.DspServiceTimeFlag))

	tx, err := endpoint.Dsp.UpdateNode(nodeAddr, volume, serviceTime)
	if err != nil {
		PrintErrorMsg("update node err: %s", err)
		return err
	}
	PrintInfoMsg("update node success, tx: %s", tx)

	return nil
}

func withDraw(ctx *cli.Context) error {
	endpoint, err := dsp.Init(config.WalletDatFilePath(), config.Parameters.BaseConfig.WalletPwd)
	if err != nil {
		PrintErrorMsg("init dsp err:%s\n", err)
		return err
	}

	err = dsp.StartDspNode(endpoint, false, false, false)
	if err != nil {
		PrintErrorMsg("start dsp err:%s\n", err)
		return err
	}
	tx, err := endpoint.Dsp.NodeWithdrawProfit()
	if err != nil {
		PrintErrorMsg("withdraw node profit err: %s", err)
		return err
	}
	PrintInfoMsg("withdraw node profit success, tx: %s", tx)

	return nil
}
