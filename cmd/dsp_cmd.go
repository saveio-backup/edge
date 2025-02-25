package cmd

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	"github.com/saveio/edge/common/config"
	eUtils "github.com/saveio/edge/utils"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/common/password"
	fs "github.com/saveio/themis/smartcontract/service/native/savefs"

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
				flags.DspUploadFilePathFlag,
				flags.DspUploadFileDescFlag,
				flags.DspUploadDurationFlag,
				flags.DspUploadProveLevelFlag,
				flags.DspUploadPrivilegeFlag,
				flags.DspUploadCopyNumFlag,
				flags.DspUploadEncryptPasswordFlag,
				flags.DspEncryptNodeAddrFlag,
				flags.DspFileUrlFlag,
				flags.DspUploadShareFlag,
				flags.DspUploadStoreTypeFlag,
				flags.TestFlag,
				flags.DspUploadFileTestCountSize,
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
				flags.DspSetFileNameFlag,
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
				flags.GasLimitFlag,
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
		{
			Action:    encryptFile,
			Name:      "encrypt",
			Usage:     "Encrypt file",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspFilePathFlag,
				flags.DspEncryptPwdFlag,
			},
			Description: "Encrypt file",
		},
		{
			Action:    decryptFile,
			Name:      "decrypt",
			Usage:     "Decrypt file",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspFilePathFlag,
				flags.DspDecryptPwdFlag,
			},
			Description: "Decrypt file",
		},
		{
			Action:    encryptFileA,
			Name:      "encrypta",
			Usage:     "Encrypt file asymmetrically",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspFilePathFlag,
				flags.DspEncryptNodeAddrFlag,
			},
			Description: "Encrypt file asymmetrically",
		},
		{
			Action:    decryptFileA,
			Name:      "decrypta",
			Usage:     "Decrypt file asymmetrically",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspFilePathFlag,
				flags.DspPrivateKeyFlag,
			},
			Description: "Decrypt file asymmetrically",
		},
		{
			Action:    getProveDeatil,
			Name:      "provedetail",
			Usage:     "Get uploaded file prove detail",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspFileHashFlag,
			},
			Description: "Get uploaded file prove detail",
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
				flags.DspBlockCountFlag,
				flags.DspBlockCountOpFlag,
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
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	pwdHash := eUtils.Sha256HexStr(string(pwd))
	fileHash := ctx.String(flags.GetFlagName(flags.DspFileHashFlag))
	url := ctx.String(flags.GetFlagName(flags.DspFileUrlFlag))
	link := ctx.String(flags.GetFlagName(flags.DspFileLinkFlag))
	decryptPwd := ctx.String(flags.GetFlagName(flags.DspDecryptPwdFlag))
	maxPeerNum := ctx.Uint64(flags.GetFlagName(flags.DspMaxPeerCntFlag))
	setFileName := ctx.Bool(flags.GetFlagName(flags.DspSetFileNameFlag))
	_, err = utils.DownloadFile(fileHash, url, link, decryptPwd, maxPeerNum, setFileName, pwdHash)
	if err != nil {
		PrintErrorMsg("download file err %s", err)
		return err
	}
	PrintInfoMsg("download file success. use <transferlist> to show transfer list")
	return nil
}

//upload can be done without dsp daemon
func fileUpload(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspUploadFilePathFlag)) &&
		!ctx.IsSet(flags.GetFlagName(flags.TestFlag)) {
		PrintErrorMsg("Missing file name.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	pwdHash := eUtils.Sha256HexStr(string(pwd))
	fileName := ctx.String(flags.GetFlagName(flags.DspUploadFilePathFlag))
	fileDesc := ctx.String(flags.GetFlagName(flags.DspUploadFileDescFlag))
	duration := ctx.String(flags.GetFlagName(flags.DspUploadDurationFlag))
	proveLevel := ctx.String(flags.GetFlagName(flags.DspUploadProveLevelFlag))
	uploadPrivilege := ctx.Uint64(flags.GetFlagName(flags.DspUploadPrivilegeFlag))
	copyNum := ctx.String(flags.GetFlagName(flags.DspUploadCopyNumFlag))
	encryptPwd := ctx.String(flags.GetFlagName(flags.DspUploadEncryptPasswordFlag))
	encryptNodeAddr := ctx.String(flags.GetFlagName(flags.DspEncryptNodeAddrFlag))
	uploadUrl := ctx.String(flags.GetFlagName(flags.DspFileUrlFlag))
	share := ctx.Bool(flags.GetFlagName(flags.DspUploadShareFlag))
	storeType := ctx.Int64(flags.GetFlagName(flags.DspUploadStoreTypeFlag))

	realFileSize := uint64(0)
	if ctx.IsSet(flags.GetFlagName(flags.DspSizeFlag)) {
		realFileSize = ctx.Uint64(flags.GetFlagName(flags.DspSizeFlag))
	} else {
		realFileSize, err = eUtils.GetFileRealSize(fileName)
		if err != nil {
			return err
		}
		realFileSize = realFileSize / 1000
		if realFileSize == 0 {
			realFileSize = 1
		}
	}
	test := ctx.Bool(flags.GetFlagName(flags.TestFlag))
	testCount := ctx.Int64(flags.GetFlagName(flags.DspUploadFileTestCountSize))
	if !test {
		_, err = utils.UploadFile(fileName, pwdHash, fileDesc, nil, encryptPwd, encryptNodeAddr,
			uploadUrl, share, duration, proveLevel, uploadPrivilege, copyNum, storeType, realFileSize)
		if err != nil {
			PrintErrorMsg("upload file err %s", err)
			return err
		}
		PrintInfoMsg("upload file success. use <transferlist> to show transfer list")
		return nil
	}
	for i := 0; i < int(testCount); i++ {
		data := make([]byte, realFileSize*1024)
		_, err = rand.Read(data)
		if err != nil {
			log.Errorf("make rand data err %s", err)
			return nil
		}
		md5Ret := md5.Sum(data)
		baseName := hex.EncodeToString(md5Ret[:])
		exts := []string{"txt", "jpg", "mp3", "mp4"}
		baseName = baseName + "." + exts[rand.Int31n(int32(len(exts)))]
		fileDesc = baseName
		fileName = filepath.Join(config.FsFileRootPath(), "/", baseName)
		ioutil.WriteFile(fileName, data, 0666)
		PrintInfoMsg("filemd5 is %s", hex.EncodeToString(md5Ret[:]))
		_, err = utils.UploadFile(fileName, pwdHash, fileDesc, nil, encryptPwd, encryptNodeAddr, uploadUrl, share, duration, proveLevel, uploadPrivilege, copyNum, storeType, realFileSize)
		if err != nil {
			PrintErrorMsg("upload file err %s", err)
			return err
		}
		PrintInfoMsg("upload file success. use <transferlist> to show transfer list")
	}

	return nil
}

func fileDelete(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspFileHashFlag)) {
		PrintErrorMsg("Missing file hash.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	pwdHash := eUtils.Sha256HexStr(string(pwd))
	hash := ctx.String(flags.GetFlagName(flags.DspFileHashFlag))
	gasLimit := ctx.Uint64(flags.GetFlagName(flags.GasLimitFlag))
	ret, err := utils.DeleteFile(hash, pwdHash, fmt.Sprintf("%v", gasLimit))
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
	// if !ctx.IsSet(flags.GetFlagName(flags.DspWalletAddrFlag)) {
	// 	PrintErrorMsg("Missing wallet address.")
	// 	cli.ShowSubcommandHelp(ctx)
	// 	return nil
	// }
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
	pwd, err := password.GetPassword()
	if err != nil {
		return err
	}
	pwdHash := eUtils.Sha256HexStr(string(pwd))
	addr := ctx.String(flags.GetFlagName(flags.DspWalletAddrFlag))
	size := ctx.Uint64(flags.GetFlagName(flags.DspSizeFlag))
	sizeOp := ctx.Uint64(flags.GetFlagName(flags.DspSizeOpFlag))
	second := ctx.Uint64(flags.GetFlagName(flags.DspBlockCountFlag))
	secondOp := ctx.Uint64(flags.GetFlagName(flags.DspBlockCountOpFlag))

	sizeMap := make(map[string]interface{}, 0)
	sizeMap["Type"] = sizeOp
	sizeMap["Value"] = size

	secondMap := make(map[string]interface{}, 0)
	secondMap["Type"] = secondOp
	secondMap["Value"] = second * config.Parameters.BaseConfig.BlockTime
	log.Debugf("set userspace: addr %v, size %v second %v", addr, sizeMap, secondMap)
	ret, err := utils.SetUserSpace(addr, pwdHash, sizeMap, secondMap)
	if err != nil {
		PrintErrorMsg("set user space err %s", err)
		return err
	}
	PrintJsonData(ret)
	return nil
}

func encryptFile(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspFilePathFlag)) {
		PrintErrorMsg("Missing file path. --filePath")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if !ctx.IsSet(flags.GetFlagName(flags.DspEncryptPwdFlag)) {
		PrintErrorMsg("Missing encrypt password. --encryptPwd")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	filePath := ctx.String(flags.GetFlagName(flags.DspFilePathFlag))
	pwd := ctx.String(flags.GetFlagName(flags.DspEncryptPwdFlag))
	_, err := utils.EncryptFile(filePath, pwd)
	if err != nil {
		PrintErrorMsg("encrypt file err %s", err)
		return err
	}
	PrintInfoMsg("Encrypt file success. See %s.ept for detail", filePath)

	return nil
}

func decryptFile(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspFilePathFlag)) {
		PrintErrorMsg("Missing file path. --filePath")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if !ctx.IsSet(flags.GetFlagName(flags.DspDecryptPwdFlag)) {
		PrintErrorMsg("Missing decrypt password. --decryptPwd")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	filePath := ctx.String(flags.GetFlagName(flags.DspFilePathFlag))
	pwd := ctx.String(flags.GetFlagName(flags.DspDecryptPwdFlag))
	ret, err := utils.DecryptFile(filePath, pwd)
	if err != nil {
		PrintErrorMsg("decrypt file err %s", err)
		return err
	}
	PrintInfoMsg("Decrypt file success")
	PrintJsonData(ret)
	return nil
}

func getProveDeatil(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspFileHashFlag)) {
		PrintErrorMsg("Missing file path. --fileHash")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	fileHash := ctx.String(flags.GetFlagName(flags.DspFileHashFlag))
	ret, err := utils.GetFileProveDetail(fileHash)
	if err != nil {
		PrintErrorMsg("encrypt file err %s", err)
		return err
	}

	details := new(fs.FsProveDetails)
	err = json.Unmarshal(ret, details)
	if err != nil {
		return err
	}

	PrintJsonObject(formatProveDetails(details))
	return nil
}

func encryptFileA(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspFilePathFlag)) {
		PrintErrorMsg("Missing file path. --filePath")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if !ctx.IsSet(flags.GetFlagName(flags.DspWalletAddrFlag)) {
		PrintErrorMsg("Missing encrypt wallet address. --walletAddr")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	filePath := ctx.String(flags.GetFlagName(flags.DspFilePathFlag))
	wa := ctx.String(flags.GetFlagName(flags.DspWalletAddrFlag))
	_, err := utils.EncryptFileA(filePath, wa)
	if err != nil {
		PrintErrorMsg("encrypt file asymmetrically err %s", err)
		return err
	}
	PrintInfoMsg("Encrypt file asymmetrically success. See %s.ept for detail", filePath)

	return nil
}

func decryptFileA(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspFilePathFlag)) {
		PrintErrorMsg("Missing file path. --filePath")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if !ctx.IsSet(flags.GetFlagName(flags.DspPrivateKeyFlag)) {
		PrintErrorMsg("Missing decrypt private key. --privateKey")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	filePath := ctx.String(flags.GetFlagName(flags.DspFilePathFlag))
	pk := ctx.String(flags.GetFlagName(flags.DspPrivateKeyFlag))
	ret, err := utils.DecryptFileA(filePath, pk)
	if err != nil {
		PrintErrorMsg("decrypt file asymmetrically err %s", err)
		return err
	}
	PrintInfoMsg("Decrypt file asymmetrically success")
	PrintJsonData(ret)
	return nil
}

type ProveDetail struct {
	NodeAddr    string
	WalletAddr  string
	ProveTimes  uint64
	BlockHeight uint64
	Finished    bool
}

type ProveDetails struct {
	CopyNum        uint64
	ProveDetailNum uint64
	ProveDetails   []ProveDetail
}

func formatProveDetails(details *fs.FsProveDetails) *ProveDetails {
	proveDetails := &ProveDetails{
		CopyNum:        details.CopyNum,
		ProveDetailNum: details.ProveDetailNum,
		ProveDetails:   make([]ProveDetail, 0),
	}

	for _, detail := range details.ProveDetails {
		proveDetail := ProveDetail{
			NodeAddr:    string(detail.NodeAddr),
			WalletAddr:  detail.WalletAddr.ToBase58(),
			ProveTimes:  detail.ProveTimes,
			BlockHeight: detail.BlockHeight,
			Finished:    detail.Finished,
		}

		proveDetails.ProveDetails = append(proveDetails.ProveDetails, proveDetail)
	}
	return proveDetails
}
