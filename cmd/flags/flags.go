package flags

import (
	"strings"

	"github.com/urfave/cli"
)

var (
	ConfigFlag = cli.StringFlag{
		Name:  "config",
		Usage: "Use `<filename>` to specifies the config file to connect to customize network.",
	}
	LaunchManualFlag = cli.BoolFlag{
		Name:  "launchManual",
		Usage: "Launch dsp manually",
	}
	//commmon
	LogStderrFlag = cli.BoolFlag{
		Name:  "logstderr",
		Usage: "Log to standard error instead of files,default false",
	}
	LogLevelFlag = cli.UintFlag{
		Name:  "loglevel",
		Usage: "Set the log level to `<level>` (0~6). 0:Trace 1:Debug 2:Info 3:Warn 4:Error 5:Fatal 6:MaxLevel",
	}

	NetworkIDFlag = cli.UintFlag{
		Name:  "networkId",
		Usage: "P2P network identity",
	}
	RpcServerFlag = cli.StringFlag{
		Name:  "rpcServer",
		Usage: "Chain RPC server address",
		Value: "",
	}

	WalletFileFlag = cli.StringFlag{
		Name:  "wallet,w",
		Usage: "Import wallet from file",
	}
	PrivateKeyFlag = cli.StringFlag{
		Name:  "privatekey, privk",
		Usage: "Import wallet from private key (WIF)",
	}
	ImportOnlineWalletFlag = cli.BoolFlag{
		Name:  "online",
		Usage: "Import for online node or not",
	}
	WalletPasswordFlag = cli.StringFlag{
		Name:  "password,p",
		Usage: "Wallet password",
	}
	WalletLabelFlag = cli.StringFlag{
		Name:  "label,l",
		Usage: "Create wallet label",
		Value: "",
	}
	WalletKeyTypeFlag = cli.StringFlag{
		Name:  "keyType,k",
		Usage: "Create wallet keyType",
		Value: "ecdsa",
	}
	WalletCurveFlag = cli.StringFlag{
		Name:  "curve,c",
		Usage: "Create wallet curve",
		Value: "P-256",
	}
	WalletSchemeFlag = cli.StringFlag{
		Name:  "scheme,s",
		Usage: "Create wallet scheme",
		Value: "SHA256withECDSA",
	}
	WalletExportTypeFlag = cli.IntFlag{
		Name:  "type,t",
		Usage: "ExportType. 0: WalletFile, 1: PrivateKey",
	}
	WalletExportFileFlag = cli.StringFlag{
		Name:  "output,o",
		Usage: "Output file",
	}

	////////////////Dsp Command Common Setting///////////////////
	DspRestAddrFlag = cli.StringFlag{
		Name:  "dspRestAddr",
		Usage: "DSP restful sever address",
	}

	////////////////Dsp Node Setting///////////////////
	DspNodeAddrFlag = cli.StringFlag{
		Name:  "nodeAddr",
		Usage: "DSP node listen address",
	}
	DspVolumeFlag = cli.StringFlag{
		Name:  "volume",
		Usage: "DSP node storage volume",
	}
	DspServiceTimeFlag = cli.StringFlag{
		Name:  "serviceTime",
		Usage: "DSP storage node service time",
	}
	DspWalletAddrFlag = cli.StringFlag{
		Name:  "walletAddr",
		Usage: "Account wallet address",
	}

	////////////////Dsp File(download) Setting///////////////////
	DspFileHashFlag = cli.StringFlag{
		Name:  "fileHash",
		Usage: "`<hash>` of file",
	}
	DspFileLinkFlag = cli.StringFlag{
		Name:  "link",
		Usage: "`<link>` of file",
	}
	DspInorderFlag = cli.BoolFlag{
		Name:  "inorder",
		Usage: "Download file in order",
	}
	DspNofeeFlag = cli.BoolFlag{
		Name:  "noFee",
		Usage: "Download file free",
	}
	DspDecryptPwdFlag = cli.StringFlag{
		Name:  "decryptPwd",
		Usage: "Decrypt file password",
	}
	DspMaxPeerCntFlag = cli.Uint64Flag{
		Name:  "maxPeerCnt",
		Usage: "Max number of peer for downloading files",
		Value: 1,
	}
	DspSetFileNameFlag = cli.BoolFlag{
		Name:  "setFileName",
		Usage: "Auto save file with its original file name.",
	}

	////////////////Dsp File(upload) Setting///////////////////
	DspUploadFileNameFlag = cli.StringFlag{
		Name:  "filePath",
		Usage: "Absolute `<path>` of file to be uploaded",
	}
	DspUploadFileDescFlag = cli.StringFlag{
		Name:  "desc",
		Usage: "File description",
	}
	DspUploadDurationFlag = cli.StringFlag{
		Name:  "duration",
		Usage: "File storage block count",
		Value: "172800",
	}
	DspFileUrlFlag = cli.StringFlag{
		Name:  "url",
		Usage: "URL of file for downloading",
	}
	DspUploadProveIntervalFlag = cli.StringFlag{
		Name:  "interval",
		Usage: "File challenge interval block count. Minimum 17280.",
		Value: "17280",
	}
	DspUploadExpiredHeightFlag = cli.StringFlag{
		Name:  "expiredHeight",
		Usage: "File expired block height",
	}
	DspUploadPrivilegeFlag = cli.Uint64Flag{
		Name:  "privilege",
		Usage: "Privilege of file sharing",
		Value: 1,
	}
	DspUploadCopyNumFlag = cli.StringFlag{
		Name:  "copyNum",
		Usage: "Copy Number of file",
	}
	DspUploadEncryptFlag = cli.BoolFlag{
		Name:  "encrypt",
		Usage: "Encrypt file or not",
	}
	DspUploadEncryptPasswordFlag = cli.StringFlag{
		Name:  "encryptPwd",
		Usage: "Encrypt password",
	}
	DspUploadShareFlag = cli.BoolFlag{
		Name:  "share",
		Usage: "Share file or not",
	}
	DspFileTypeFlag = cli.StringFlag{
		Name:  "fileType",
		Usage: "File list type",
		Value: "0",
	}
	DspListOffsetFlag = cli.StringFlag{
		Name:  "offset",
		Usage: "File list offset",
		Value: "0",
	}
	DspListLimitFlag = cli.StringFlag{
		Name:  "limit",
		Usage: "File list size limit",
		Value: "0",
	}
	DspFileTransferTypeFlag = cli.StringFlag{
		Name:  "transferType",
		Usage: "File transfer type",
		Value: "0",
	}
	DspUploadStoreTypeFlag = cli.Int64Flag{
		Name:  "storeType",
		Usage: "Store file type",
	}
	DspFileNameFlag = cli.StringFlag{
		Name:  "fileName",
		Usage: "File name",
	}
	DspFileBlocksRoot = cli.StringFlag{
		Name:  "blocksRoot",
		Usage: "Merkle root hash for all block hash of file",
	}
	DspFileOwner = cli.StringFlag{
		Name:  "fileOwner",
		Usage: "Wallet address of the file owner",
	}
	DspFileSize = cli.Int64Flag{
		Name:  "fileSize",
		Usage: "File size. (KB)",
	}
	DspBlockCountSize = cli.Int64Flag{
		Name:  "blockCount",
		Usage: "File block count.",
	}

	////////////////Dsp File(delete) Setting///////////////////
	DspDeleteLocalFlag = cli.BoolFlag{
		Name:  "local",
		Usage: "Delete remote file or local file",
	}

	DspBlockCountOpFlag = cli.Uint64Flag{
		Name:  "blockCountOp",
		Usage: "User space storage block count operation.0: none, 1: add, 2: revoke",
		Value: 1,
	}
	DspBlockCountFlag = cli.Uint64Flag{
		Name:  "blockCount",
		Usage: "User space storage block count",
		Value: 172800,
	}
	DspSizeFlag = cli.Uint64Flag{
		Name:  "size",
		Usage: "User space size.(KB)",
		Value: 1024000,
	}
	DspSizeOpFlag = cli.Uint64Flag{
		Name:  "sizeOp",
		Usage: "User space size operation.0: none, 1: add, 2: revoke",
		Value: 1,
	}

	////////////////Dsp DNS Command Setting///////////////////
	DnsURLFlag = cli.StringFlag{
		Name:  "url",
		Usage: "`<url>` of the file",
	}
	DnsLinkFlag = cli.StringFlag{
		Name:  "dnsLink",
		Usage: "Dns `<link>`",
	}
	DnsIpFlag = cli.StringFlag{
		Name:  "dnsIp",
		Usage: "Dns `<ip>`",
	}
	DnsPortFlag = cli.StringFlag{
		Name:  "dnsPort",
		Usage: "Dns `<port>`",
	}
	DnsWalletFlag = cli.StringFlag{
		Name:  "walletAddr",
		Usage: "Dns `<walletAddr>`",
	}
	DnsAllFlag = cli.BoolFlag{
		Name:  "all",
		Usage: "All Dns info",
	}
	/////////////Dsp Channel Command Setting////////////
	PartnerAddressFlag = cli.StringFlag{
		Name:  "partnerAddr",
		Usage: "Channel partner `<address>`",
	}
	TargetAddressFlag = cli.StringFlag{
		Name:  "targetAddr",
		Usage: "Channel transfer target `<address>`",
	}
	TotalDepositFlag = cli.Uint64Flag{
		Name:  "totalDeposit",
		Usage: "Channel total `<deposit>`",
	}
	AmountFlag = cli.Uint64Flag{
		Name:  "amount",
		Usage: "Channel payment amount `<amount>`",
	}
	AmountStrFlag = cli.StringFlag{
		Name:  "amount",
		Usage: "Channel payment amount `<amount>`. float",
	}
	PaymentIDFlag = cli.Uint64Flag{
		Name:  "paymentId",
		Usage: "",
	}

	/////////////Dsp Governance Command Setting////////////
	PeerPubkeyFlag = cli.StringFlag{
		Name:  "peerPubkey",
		Usage: "candidate pubkey",
	}
	InitDepositFlag = cli.StringFlag{
		Name:  "initDeposit",
		Usage: "Init `<deposit>`",
	}
	PeerPubkeyListFlag = cli.StringFlag{
		Name:  "peerPubkeyList",
		Usage: "candidate pubkey list",
	}
	WithdrawListFlag = cli.StringFlag{
		Name:  "withdrawList",
		Usage: "withdraw value list",
	}
	DeltaDepositFlag = cli.StringFlag{
		Name:  "deltaDeposit",
		Usage: "Delta `<deposit>`",
	}

	TestFlag = cli.BoolFlag{
		Name:  "test",
		Usage: "Use in test case",
	}
	ProfileFlag = cli.BoolFlag{
		Name:  "profile",
		Usage: "Profile with memory",
	}
	AddressFlag = cli.StringFlag{
		Name:  "address,addr",
		Usage: "Wallet address",
	}
	VerboseFlag = cli.StringFlag{
		Name:  "verbose,v",
		Usage: "Show detail message",
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
