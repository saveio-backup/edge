package cmd

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"math/rand"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	"github.com/saveio/edge/common/config"
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
				flags.DspUploadDurationFlag,
				flags.DspUploadChallengeRateFlag,
				flags.DspUploadChallengeTimesFlag,
				flags.DspUploadPrivilegeFlag,
				flags.DspUploadCopyNumFlag,
				flags.DspUploadEncryptPasswordFlag,
				flags.DspFileUrlFlag,
				flags.DspUploadShareFlag,
				flags.TestFlag,
				flags.DspSizeFlag,
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
				flags.DspFileUrlFlag,
				flags.DspFileLinkFlag,
				flags.DspDecryptPwdFlag,
				flags.DspMaxPeerCntFlag,
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
			},
			Description: "Delete file",
		},
		{
			Action:    getUploadList,
			Name:      "uploadlist",
			Usage:     "Get upload files list",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspFileTypeFlag,
				flags.DspListOffsetFlag,
				flags.DspListLimitFlag,
			},
			Description: "Get upload file list",
		},
		{
			Action:    getDownloadList,
			Name:      "downloadlist",
			Usage:     "Get download files list",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspFileTypeFlag,
				flags.DspListOffsetFlag,
				flags.DspListLimitFlag,
			},
			Description: "Get download file list",
		},
		{
			Action:    getTransferList,
			Name:      "transferlist",
			Usage:     "Get transfer files list",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspFileTransferTypeFlag,
				flags.DspListOffsetFlag,
				flags.DspListLimitFlag,
			},
			Description: "Get transfer file list",
		},
	},
	Description: `./edge file --help command to view help information.`,
}

var UserspaceCommand = cli.Command{
	Name:  "userspace",
	Usage: "Manage user file storage space using dsp",
	Subcommands: []cli.Command{
		{
			Action:    getUserSpace,
			Name:      "show",
			Usage:     "Get user space",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspWalletAddrFlag,
			},
			Description: "Get user space",
		},
		{
			Action:    setUserSpace,
			Name:      "set",
			Usage:     "Set user space",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspWalletAddrFlag,
				flags.DspSecondFlag,
				flags.DspSecondOpFlag,
				flags.DspSizeFlag,
				flags.DspSizeOpFlag,
			},
			Description: "Set user space size/second",
		},
	},
	Description: `./edge userspace --help command to view help information.`,
}

//file command
func fileDownload(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspFileHashFlag)) && !ctx.IsSet(flags.GetFlagName(flags.DspFileUrlFlag)) && !ctx.IsSet(flags.GetFlagName(flags.DspFileLinkFlag)) {
		PrintErrorMsg("Missing file hash/url/link.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	fileHash := ctx.String(flags.GetFlagName(flags.DspFileHashFlag))
	url := ctx.String(flags.GetFlagName(flags.DspFileUrlFlag))
	link := ctx.String(flags.GetFlagName(flags.DspFileLinkFlag))
	password := ctx.String(flags.GetFlagName(flags.DspDecryptPwdFlag))
	maxPeerNum := ctx.Uint64(flags.GetFlagName(flags.DspMaxPeerCntFlag))
	_, err := utils.DownloadFile(fileHash, url, link, password, maxPeerNum)
	if err != nil {
		PrintErrorMsg("download file err %s", err)
		return err
	}
	PrintInfoMsg("download file succes. use <list> to show transfer list")
	return nil
}

//upload can be done without dsp daemon
func fileUpload(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspUploadFileNameFlag)) {
		PrintErrorMsg("Missing file name.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	fileName := ctx.String(flags.GetFlagName(flags.DspUploadFileNameFlag))
	fileDesc := ctx.String(flags.GetFlagName(flags.DspUploadFileDescFlag))
	duration := ctx.String(flags.GetFlagName(flags.DspUploadDurationFlag))
	rate := ctx.String(flags.GetFlagName(flags.DspUploadChallengeRateFlag))
	challengeTimes := ctx.String(flags.GetFlagName(flags.DspUploadChallengeTimesFlag))
	uploadPrivilege := ctx.Uint64(flags.GetFlagName(flags.DspUploadPrivilegeFlag))
	copyNum := ctx.String(flags.GetFlagName(flags.DspUploadCopyNumFlag))
	encryptPwd := ctx.String(flags.GetFlagName(flags.DspUploadEncryptPasswordFlag))
	uploadUrl := ctx.String(flags.GetFlagName(flags.DspFileUrlFlag))
	share := ctx.Bool(flags.GetFlagName(flags.DspUploadShareFlag))
	test := ctx.Bool(flags.GetFlagName(flags.TestFlag))
	if test {
		dataSize := ctx.Uint64(flags.GetFlagName(flags.DspSizeFlag))
		data := make([]byte, dataSize*1024)
		_, err := rand.Read(data)
		if err != nil {
			log.Errorf("make rand data err %s", err)
			return nil
		}
		md5Ret := md5.Sum(data)
		baseName := hex.EncodeToString(md5Ret[:])
		fileDesc = baseName
		fileName = config.FsFileRootPath() + "/" + baseName
		ioutil.WriteFile(fileName, data, 0666)
		PrintInfoMsg("filemd5 is %s", hex.EncodeToString(md5Ret[:]))
	}
	_, err := utils.UploadFile(fileName, fileDesc, nil, encryptPwd, uploadUrl, share, duration, rate, challengeTimes, uploadPrivilege, copyNum)
	if err != nil {
		PrintErrorMsg("upload file err %s", err)
		return err
	}
	PrintInfoMsg("upload file succes. use <list> to show transfer list")
	return nil
}

func fileDelete(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspFileHashFlag)) {
		PrintErrorMsg("Missing file hash.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	hash := ctx.String(flags.GetFlagName(flags.DspFileHashFlag))
	ret, err := utils.DeleteFile(hash)
	if err != nil {
		PrintErrorMsg("delete file err %s", err)
		return err
	}
	PrintJsonData(ret)
	return nil
}

func getUploadList(ctx *cli.Context) error {
	fileType := ctx.String(flags.GetFlagName(flags.DspFileTypeFlag))
	offset := ctx.String(flags.GetFlagName(flags.DspListOffsetFlag))
	limit := ctx.String(flags.GetFlagName(flags.DspListLimitFlag))
	ret, err := utils.GetUploadFiles(fileType, offset, limit)
	if err != nil {
		PrintErrorMsg("get upload file list err %s", err)
		return err
	}
	PrintJsonData(ret)
	return nil
}

func getDownloadList(ctx *cli.Context) error {
	fileType := ctx.String(flags.GetFlagName(flags.DspFileTypeFlag))
	offset := ctx.String(flags.GetFlagName(flags.DspListOffsetFlag))
	limit := ctx.String(flags.GetFlagName(flags.DspListLimitFlag))
	ret, err := utils.GetDownloadFiles(fileType, offset, limit)
	if err != nil {
		PrintErrorMsg("get upload file list err %s", err)
		return err
	}
	PrintJsonData(ret)
	return nil
}

func getTransferList(ctx *cli.Context) error {
	transferType := ctx.String(flags.GetFlagName(flags.DspFileTransferTypeFlag))
	offset := ctx.String(flags.GetFlagName(flags.DspListOffsetFlag))
	limit := ctx.String(flags.GetFlagName(flags.DspListLimitFlag))
	ret, err := utils.GetTransferList(transferType, offset, limit)
	if err != nil {
		PrintErrorMsg("get upload file list err %s", err)
		return err
	}
	PrintJsonData(ret)
	return nil
}

func getUserSpace(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspWalletAddrFlag)) {
		PrintErrorMsg("Missing wallet address.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	addr := ctx.String(flags.GetFlagName(flags.DspWalletAddrFlag))
	ret, err := utils.GetUserSpace(addr)
	if err != nil {
		PrintErrorMsg("get upload file list err %s", err)
		return err
	}
	PrintJsonData(ret)
	return nil
}

func setUserSpace(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspWalletAddrFlag)) {
		PrintErrorMsg("Missing wallet address.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	addr := ctx.String(flags.GetFlagName(flags.DspWalletAddrFlag))
	size := ctx.Uint64(flags.GetFlagName(flags.DspSizeFlag))
	sizeOp := ctx.Uint64(flags.GetFlagName(flags.DspSizeOpFlag))
	second := ctx.Uint64(flags.GetFlagName(flags.DspSecondFlag))
	secondOp := ctx.Uint64(flags.GetFlagName(flags.DspSecondOpFlag))

	sizeMap := make(map[string]interface{}, 0)
	sizeMap["Type"] = sizeOp
	sizeMap["Size"] = size

	secondMap := make(map[string]interface{}, 0)
	secondMap["Type"] = secondOp
	secondMap["Second"] = second
	ret, err := utils.SetUserSpace(addr, sizeMap, secondMap)
	if err != nil {
		PrintErrorMsg("get upload file list err %s", err)
		return err
	}
	PrintJsonData(ret)
	return nil
}
