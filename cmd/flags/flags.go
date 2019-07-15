package flags

import (
	"strings"

	"github.com/urfave/cli"
)

var (
	ConfigFlag = cli.StringFlag{
		Name:  "config",
		Usage: "Use `<filename>` to specifies the config file to connect to cunstomize network.",
	}
	//commmon
	LogStderrFlag = cli.BoolFlag{
		Name:  "logstderr",
		Usage: "log to standard error instead of files,default false",
	}
	LogLevelFlag = cli.UintFlag{
		Name:  "loglevel",
		Usage: "Set the log level to `<level>` (0~6). 0:Trace 1:Debug 2:Info 3:Warn 4:Error 5:Fatal 6:MaxLevel",
	}

	NetworkIDFlag = cli.UintFlag{
		Name:  "networkId",
		Usage: "",
	}
	RpcServerFlag = cli.StringFlag{
		Name:  "rpcServer",
		Usage: "",
		Value: "",
	}

	WalletFileFlag = cli.StringFlag{
		Name:  "wallet,w",
		Usage: "Import wallet from file",
	}
	ImportOnlineWalletFlag = cli.BoolFlag{
		Name:  "online",
		Usage: "Import for online node or not",
	}
	WalletPasswordFlag = cli.StringFlag{
		Name:  "password,p",
		Usage: "Create wallet password",
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

	/////////////Dsp Protocol Setting////////////
	ProtocolListenPortOffsetFlag = cli.UintFlag{
		Name:  "protocolListenPortOffset",
		Usage: "",
	}
	ProtocolFsRepoRootFlag = cli.StringFlag{
		Name:  "protocolFsRepoRoot",
		Usage: "",
		Value: "",
	}
	ProtocolFsFileRootFlag = cli.StringFlag{
		Name:  "protocolFsFileRoot",
		Usage: "",
		Value: "",
	}
	ProtocolSeedFlag = cli.BoolFlag{
		Name:  "protocolSeed",
		Usage: "",
	}
	ProtocolDebugFlag = cli.BoolFlag{
		Name:  "protocolDebug",
		Usage: "",
	}
	ProtocolProxyAddrFlag = cli.StringFlag{
		Name:  "protocolProxyAddr",
		Usage: "",
	}

	////////////////Dsp Tracker Setting///////////////////
	TrackerSizeFlag = cli.IntFlag{
		Name:  "trackerSize",
		Usage: "",
	}
	TrackerPortFlag = cli.IntFlag{
		Name:  "trackerPort",
		Usage: "",
	}
	TrackerUrlFlag = cli.StringFlag{
		Name:  "trackerUrl",
		Usage: "",
	}

	////////////////Dsp Channel Setting///////////////////
	ChannelChainRpcURLFlag = cli.StringFlag{
		Name:  "channelChainRpcURL",
		Usage: "",
	}
	ChannelProtocolFlag = cli.StringFlag{
		Name:  "channelProtocol",
		Usage: "",
	}
	ChannelDBPathFlag = cli.StringFlag{
		Name:  "channelDbPath",
		Usage: "",
	}

	////////////////Dsp Command Common Setting///////////////////
	DspRestAddrFlag = cli.StringFlag{
		Name:  "dspRestAddr",
		Usage: "dsp rest sever address",
	}

	////////////////Dsp Node Setting///////////////////
	DspNodeAddrFlag = cli.StringFlag{
		Name:  "nodeAddr",
		Usage: "Node address",
	}
	DspVolumeFlag = cli.StringFlag{
		Name:  "volume",
		Usage: "",
	}
	DspServiceTimeFlag = cli.StringFlag{
		Name:  "serviceTime",
		Usage: "",
	}
	DspWalletAddrFlag = cli.StringFlag{
		Name:  "walletAddr",
		Usage: "",
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
		Usage: "",
	}
	DspNofeeFlag = cli.BoolFlag{
		Name:  "noFee",
		Usage: "",
	}
	DspDecryptPwdFlag = cli.StringFlag{
		Name:  "decryptPwd",
		Usage: "",
	}
	DspMaxPeerCntFlag = cli.Uint64Flag{
		Name:  "maxPeerCnt",
		Usage: "",
		Value: 1,
	}
	DspProgressEnableFlag = cli.BoolFlag{
		Name:  "progressEnable",
		Usage: "",
	}
	DspSetFileNameFlag = cli.BoolFlag{
		Name:  "setFileName",
		Usage: "",
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
		Usage: "File storage life cycle",
	}
	DspFileUrlFlag = cli.StringFlag{
		Name:  "url",
		Usage: "File search url",
	}
	DspUploadChallengeRateFlag = cli.StringFlag{
		Name:  "interval",
		Usage: "File challenge interval",
	}
	DspUploadChallengeTimesFlag = cli.StringFlag{
		Name:  "challengeTimes",
		Usage: "Total challenge times",
	}
	DspUploadPrivilegeFlag = cli.Uint64Flag{
		Name:  "privilege",
		Usage: "Privilege of file sharing",
		Value: 1,
	}
	DspUploadCopyNumFlag = cli.StringFlag{
		Name:  "copyNum",
		Usage: "Copy Number of file storage",
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

	////////////////Dsp File(delete) Setting///////////////////
	DspDeleteLocalFlag = cli.BoolFlag{
		Name:  "local",
		Usage: "Delete remote file or local file",
	}

	DspSecondOpFlag = cli.Uint64Flag{
		Name:  "secondOp",
		Usage: "User space second operation",
		Value: 0,
	}
	DspSecondFlag = cli.Uint64Flag{
		Name:  "second",
		Usage: "User space second",
		Value: 0,
	}
	DspSizeFlag = cli.Uint64Flag{
		Name:  "size",
		Usage: "User space size",
		Value: 0,
	}
	DspSizeOpFlag = cli.Uint64Flag{
		Name:  "sizeOp",
		Usage: "User space size operation",
		Value: 0,
	}

	////////////////Dsp DNS Command Setting///////////////////
	DnsURLFlag = cli.StringFlag{
		Name:  "dnsUrl",
		Usage: "Dns `<url>`",
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
		Usage: "use in test case",
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
