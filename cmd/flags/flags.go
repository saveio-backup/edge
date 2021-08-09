package flags

import (
	"strings"

	"github.com/urfave/cli"
)

var (
	ConfigFlag = cli.StringFlag{
		Name:  "config",
		Usage: "Use `<filename>` to specifies the config file to connect to customize network. [string]",
	}
	LaunchManualFlag = cli.BoolFlag{
		Name:  "launchManual",
		Usage: "Launch dsp manually. [bool]",
	}
	//commmon
	LogStderrFlag = cli.BoolFlag{
		Name:  "logstderr",
		Usage: "Log to standard error instead of files,default false. [bool]",
	}
	LogLevelFlag = cli.UintFlag{
		Name:  "loglevel",
		Usage: "Set the log level to `<level>` (0~6). 0:Trace 1:Debug 2:Info 3:Warn 4:Error 5:Fatal 6:MaxLevel. [uint]. [int]",
	}

	NetworkIDFlag = cli.UintFlag{
		Name:  "networkId",
		Usage: "P2P network identity. [uint]. [int]",
	}
	RpcServerFlag = cli.StringFlag{
		Name:  "rpcServer",
		Usage: "Chain RPC server address. [string]",
		Value: "",
	}

	WalletFileFlag = cli.StringFlag{
		Name:  "wallet,w",
		Usage: "Import wallet from file. [string]",
	}
	PrivateKeyFlag = cli.StringFlag{
		Name:  "privatekey, privk",
		Usage: "Import wallet from private key (WIF). [string]",
	}
	ImportOnlineWalletFlag = cli.BoolFlag{
		Name:  "online",
		Usage: "Import for online node or not. [bool]",
	}
	WalletPasswordFlag = cli.StringFlag{
		Name:  "password,p",
		Usage: "Wallet password. [string]",
	}
	WalletLabelFlag = cli.StringFlag{
		Name:  "label,l",
		Usage: "Create wallet label. [string]",
		Value: "",
	}
	WalletKeyTypeFlag = cli.StringFlag{
		Name:  "keyType,k",
		Usage: "Create wallet keyType. [string]",
		Value: "ecdsa",
	}
	WalletCurveFlag = cli.StringFlag{
		Name:  "curve,c",
		Usage: "Create wallet curve. [string]",
		Value: "P-256",
	}
	WalletSchemeFlag = cli.StringFlag{
		Name:  "scheme,s",
		Usage: "Create wallet scheme. [string]",
		Value: "SHA256withECDSA",
	}
	WalletExportTypeFlag = cli.IntFlag{
		Name:  "type,t",
		Usage: "ExportType. 0: WalletFile, 1: PrivateKey. [int]",
	}
	WalletExportFileFlag = cli.StringFlag{
		Name:  "output,o",
		Usage: "Output file. [string]",
	}

	////////////////Dsp Command Common Setting///////////////////
	DspRestAddrFlag = cli.StringFlag{
		Name:  "dspRestAddr",
		Usage: "DSP restful sever address. [string]",
	}

	////////////////Dsp Node Setting///////////////////
	DspNodeAddrFlag = cli.StringFlag{
		Name:  "nodeAddr",
		Usage: "DSP node listen address. [string]",
	}
	DspVolumeFlag = cli.StringFlag{
		Name:  "volume",
		Usage: "DSP node storage volume. [string]",
	}
	DspServiceTimeFlag = cli.StringFlag{
		Name:  "serviceTime",
		Usage: "DSP storage node service time. [string]",
	}
	DspWalletAddrFlag = cli.StringFlag{
		Name:  "walletAddr",
		Usage: "Account wallet address. [string]",
	}

	////////////////Dsp File(download) Setting///////////////////
	DspFileHashFlag = cli.StringFlag{
		Name:  "fileHash",
		Usage: "`<hash>` of file. [string]",
	}
	DspFileLinkFlag = cli.StringFlag{
		Name:  "link",
		Usage: "`<link>` of file. [string]",
	}
	DspInorderFlag = cli.BoolFlag{
		Name:  "inorder",
		Usage: "Download file in order. [bool]",
	}
	DspNofeeFlag = cli.BoolFlag{
		Name:  "noFee",
		Usage: "Download file free. [bool]",
	}
	DspDecryptPwdFlag = cli.StringFlag{
		Name:  "decryptPwd",
		Usage: "Decrypt file password. [string]",
	}
	DspMaxPeerCntFlag = cli.Uint64Flag{
		Name:  "maxPeerCnt",
		Usage: "Max number of peer for downloading files. [uint64]",
		Value: 1,
	}
	DspSetFileNameFlag = cli.BoolFlag{
		Name:  "setFileName",
		Usage: "Auto save file with its original file name. [bool]",
	}

	////////////////Dsp File(upload) Setting///////////////////
	DspUploadFilePathFlag = cli.StringFlag{
		Name:  "filePath",
		Usage: "Absolute `<path>` of file to be uploaded. [string]",
	}
	DspFilePathFlag = cli.StringFlag{
		Name:  "filePath",
		Usage: "Absolute `<path>` of the file. [string]",
	}
	DspEncryptPwdFlag = cli.StringFlag{
		Name:  "encryptPwd",
		Usage: "Password for encrypting the file [string]",
	}
	DspUploadFileDescFlag = cli.StringFlag{
		Name:  "desc",
		Usage: "File description. [string]",
	}
	DspUploadDurationFlag = cli.StringFlag{
		Name:  "duration",
		Usage: "File storage block count. [string]",
		Value: "172800",
	}
	DspFileUrlFlag = cli.StringFlag{
		Name:  "url",
		Usage: "URL of file for downloading. [string]",
	}
	DspUploadProveLevelFlag = cli.StringFlag{
		Name:  "proveLevel",
		Usage: "File prove level, 1:high 2:medium 3:low",
		Value: "1",
	}
	DspUploadExpiredHeightFlag = cli.StringFlag{
		Name:  "expiredHeight",
		Usage: "File expired block height. [string]",
	}
	DspUploadPrivilegeFlag = cli.Uint64Flag{
		Name:  "privilege",
		Usage: "Privilege of file sharing. [uint64]",
		Value: 1,
	}
	DspUploadCopyNumFlag = cli.StringFlag{
		Name:  "copyNum",
		Usage: "Copy Number of file. [string]",
	}
	DspUploadEncryptFlag = cli.BoolFlag{
		Name:  "encrypt",
		Usage: "Encrypt file or not. [bool]",
	}
	DspUploadEncryptPasswordFlag = cli.StringFlag{
		Name:  "encryptPwd",
		Usage: "Encrypt password. [string]",
	}
	DspUploadShareFlag = cli.BoolFlag{
		Name:  "share",
		Usage: "Share file or not. [bool]",
	}
	DspFileTypeFlag = cli.StringFlag{
		Name:  "fileType",
		Usage: "File list type. [string]",
		Value: "0",
	}
	DspListOffsetFlag = cli.StringFlag{
		Name:  "offset",
		Usage: "File list offset. [string]",
		Value: "0",
	}
	DspListLimitFlag = cli.StringFlag{
		Name:  "limit",
		Usage: "File list size limit. [string]",
		Value: "0",
	}
	DspFileTransferTypeFlag = cli.StringFlag{
		Name:  "transferType",
		Usage: "File transfer type. [string]",
		Value: "0",
	}
	DspUploadStoreTypeFlag = cli.Int64Flag{
		Name:  "storeType",
		Usage: "Store file type. 0, space mode, 1 advance mode.[int64]",
	}
	DspFileNameFlag = cli.StringFlag{
		Name:  "fileName",
		Usage: "File name. [string]",
	}
	DspFileBlocksRoot = cli.StringFlag{
		Name:  "blocksRoot",
		Usage: "Merkle root hash for all block hash of file. [string]",
	}
	DspFileOwner = cli.StringFlag{
		Name:  "fileOwner",
		Usage: "Wallet address of the file owner. [string]",
	}
	DspFileSize = cli.Int64Flag{
		Name:  "fileSize",
		Usage: "File size. (KB). [int64]",
	}
	DspBlockCountSize = cli.Int64Flag{
		Name:  "blockCount",
		Usage: "File block count. [int64]",
	}
	DspUploadFileTestCountSize = cli.Int64Flag{
		Name:  "count",
		Usage: "Upload test file count. [int64]",
	}

	////////////////Dsp File(delete) Setting///////////////////
	DspDeleteLocalFlag = cli.BoolFlag{
		Name:  "local",
		Usage: "Delete remote file or local file. [bool]",
	}

	DspBlockCountOpFlag = cli.Uint64Flag{
		Name:  "blockCountOp",
		Usage: "User space storage block count operation.0: none, 1: add, 2: revoke. [uint64]",
		Value: 1,
	}
	DspBlockCountFlag = cli.Uint64Flag{
		Name:  "blockCount",
		Usage: "User space storage block count. [uint64]",
		Value: 172800,
	}
	DspSizeFlag = cli.Uint64Flag{
		Name:  "size",
		Usage: "User space size.(KB). [uint64]",
		Value: 1024000,
	}
	DspSizeOpFlag = cli.Uint64Flag{
		Name:  "sizeOp",
		Usage: "User space size operation.0: none, 1: add, 2: revoke. [uint64]",
		Value: 1,
	}

	////////////////Dsp DNS Command Setting///////////////////
	DnsURLFlag = cli.StringFlag{
		Name:  "url",
		Usage: "`<url>` of the file. [string]",
	}
	DnsLinkFlag = cli.StringFlag{
		Name:  "dnsLink",
		Usage: "Dns `<link>`. [string]",
	}
	DnsHeaderFlag = cli.StringFlag{
		Name:  "header",
		Usage: "`<header>` of the dns url. [string]",
	}
	DnsDescFlag = cli.StringFlag{
		Name:  "desc",
		Usage: "`<desc>` of the header. [string]",
	}
	DnsTTLFlag = cli.Uint64Flag{
		Name:  "ttl",
		Usage: "header ttl. [uint64]",
	}
	DnsIpFlag = cli.StringFlag{
		Name:  "dnsIp",
		Usage: "Dns `<ip>`. [string]",
	}
	DnsPortFlag = cli.StringFlag{
		Name:  "dnsPort",
		Usage: "Dns `<port>`. [string]",
	}
	DnsWalletFlag = cli.StringFlag{
		Name:  "walletAddr",
		Usage: "Dns `<walletAddr>`. [string]",
	}
	DnsAllFlag = cli.BoolFlag{
		Name:  "all",
		Usage: "All Dns info. [bool]",
	}
	/////////////Dsp Channel Command Setting////////////
	PartnerAddressFlag = cli.StringFlag{
		Name:  "partnerAddr",
		Usage: "Channel partner `<address>`. [string]",
	}
	TargetAddressFlag = cli.StringFlag{
		Name:  "targetAddr",
		Usage: "Channel transfer target `<address>`. [string]",
	}
	TotalDepositFlag = cli.Uint64Flag{
		Name:  "totalDeposit",
		Usage: "Channel total `<deposit>`. [uint64]",
	}
	AmountFlag = cli.Uint64Flag{
		Name:  "amount",
		Usage: "Channel payment amount `<amount>`. [uint64]",
	}
	AmountStrFlag = cli.StringFlag{
		Name:  "amount",
		Usage: "Channel payment amount `<amount>`. float. [string]",
	}
	PaymentIDFlag = cli.Uint64Flag{
		Name:  "paymentId",
		Usage: ". [uint64]",
	}

	/////////////Dsp Governance Command Setting////////////
	PeerPubkeyFlag = cli.StringFlag{
		Name:  "peerPubkey",
		Usage: "candidate pubkey. [string]",
	}
	InitDepositFlag = cli.StringFlag{
		Name:  "initDeposit",
		Usage: "Init `<deposit>`. [string]",
	}
	PeerPubkeyListFlag = cli.StringFlag{
		Name:  "peerPubkeyList",
		Usage: "candidate pubkey list. [string]",
	}
	WithdrawListFlag = cli.StringFlag{
		Name:  "withdrawList",
		Usage: "withdraw value list. [string]",
	}
	DeltaDepositFlag = cli.StringFlag{
		Name:  "deltaDeposit",
		Usage: "Delta `<deposit>`. [string]",
	}

	TestFlag = cli.BoolFlag{
		Name:  "test",
		Usage: "Use in test case. [bool]",
	}
	ProfileFlag = cli.BoolFlag{
		Name:  "profile",
		Usage: "Profile with memory. [bool]",
	}
	AddressFlag = cli.StringFlag{
		Name:  "address,addr",
		Usage: "Wallet address. [string]",
	}
	GasPriceFlag = cli.Uint64Flag{
		Name:  "gasPrice,gp",
		Usage: "Invoke smart contract gas price. [uint64]",
		Value: 500,
	}
	GasLimitFlag = cli.Uint64Flag{
		Name:  "gasLimit,gl",
		Usage: "Invoke smart contract gas limit. [uint64]",
		Value: 20000,
	}
	VerboseFlag = cli.StringFlag{
		Name:  "verbose,v",
		Usage: "Show detail message. [string]",
	}

	////////////////Dsp Sector Setting///////////////////
	DspSectorIdFlag = cli.StringFlag{
		Name:  "sectorId",
		Usage: "Sector id. Sector id should be bigger than 0. [string]",
	}
	DspSectorProveLevelFlag = cli.Uint64Flag{
		Name:  "proveLevel",
		Usage: "Sector prove level.1: high, 2: medium, 3:low. [uint64]",
		Value: 1,
	}
	DspSectorSizeFlag = cli.Uint64Flag{
		Name:  "sectorSize",
		Usage: "Sector size. (KB). Minimum sector size is 1GB. [uint64]",
		Value: 1048576,
	}
	DspSectorPlotFlag = cli.BoolFlag{
		Name:  "isPlot",
		Usage: "Set if the sector is used to store plot files",
	}

	////////////////Dsp Plot Setting///////////////////
	PlotSystemFlag = cli.StringFlag{
		Name:  "system",
		Usage: "On which system to run the plot tool.\"win\" for windows, \"linux\" for linux",
	}
	PlotNumericIDFlag = cli.StringFlag{
		Name:  "numericID",
		Usage: "Numeric Account ID",
	}
	PlotSizeFlag = cli.Uint64Flag{
		Name:  "size",
		Usage: "Plot file size(KB), should be an integer multiple of 2048",
		Value: 31457280,
	}
	PlotNumFlag = cli.Uint64Flag{
		Name:  "num",
		Usage: "Plot file num",
		Value: 1,
	}
	PlotStartNonceFlag = cli.Uint64Flag{
		Name:  "startNonce",
		Usage: "Where you want to start plotting",
	}
	PlotNoncesFlag = cli.Uint64Flag{
		Name:  "nonces",
		Usage: "How many nonces you want to plot, should be an integer multiple of 8. 1 nonces = 256KB",
	}
	PlotPathFlag = cli.StringFlag{
		Name:  "path",
		Usage: "Target path for plotfile",
	}

	CreateSectorFlag = cli.BoolFlag{
		Name:  "create-sector",
		Usage: "Auto create sectors if space not enough",
	}
)

//GetFlagName deal with short flag, and return the flag name whether flag name have short name
func GetFlagName(flag cli.Flag) string {
	name := flag.GetName()
	if name == "" {
		return ""
	}
	return strings.TrimSpace(strings.Split(name, ",")[0])
}
