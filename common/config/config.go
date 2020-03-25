package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/common"
	"github.com/saveio/themis/common/log"

	"github.com/urfave/cli"
)

//default common parameter
const (
	VERSION                 = "0.1"
	DEFAULT_MAX_LOG_SIZE    = 20 * 1024 * 1024 //MB
	DEFAULT_CONFIG_FILENAME = "config.json"
)

type BaseConfig struct {
	BaseDir            string `json:"BaseDir"`
	LogPath            string `json:"LogPath"`
	LogMaxSize         uint64 `json:"LogMaxSize"`
	ChainId            string `json:"ChainId"`
	BlockTime          uint64 `json:"BlockTime"`
	NetworkId          uint32 `json:"NetworkId"`
	PublicIP           string `json:"PublicIP"`
	IntranetIP         string `json:"IntranetIP"`
	PortBase           uint32 `json:"PortBase"`
	LogLevel           int    `json:"LogLevel"`
	LocalRpcPortOffset int    `json:"LocalRpcPortOffset"`
	EnableLocalRpc     bool   `json:"EnableLocalRpc"`
	JsonRpcPortOffset  int    `json:"JsonRpcPortOffset"`
	EnableJsonRpc      bool   `json:"EnableJsonRpc"`
	HttpRestPortOffset int    `json:"HttpRestPortOffset"`
	HttpCertPath       string `json:"HttpCertPath"`
	HttpKeyPath        string `json:"HttpKeyPath"`
	RestEnable         bool   `json:"RestEnable"`
	WsPortOffset       int    `json:"WsPortOffset"`
	WsCertPath         string `json:"WsCertPath"`
	WsKeyPath          string `json:"WsKeyPath"`

	ChannelPortOffset int    `json:"ChannelPortOffset"`
	ChannelProtocol   string `json:"ChannelProtocol"`

	DBPath         string   `json:"DBPath"`
	ChainRestAddrs []string `json:"ChainRestAddrs"`
	ChainRpcAddrs  []string `json:"ChainRpcAddrs"`

	NATProxyServerAddrs string `json:"NATProxyServerAddrs"`
	DspProtocol         string `json:"DspProtocol"`
	DspPortOffset       int    `json:"DspPortOffset"`

	TrackerNetworkId  uint32 `json:"TrackerNetworkId"`
	TrackerProtocol   string `json:"TrackerProtocol"`
	TrackerPortOffset int    `json:"TrackerPortOffset"`

	ProfilePortOffset int `json:"ProfilePortOffset"`

	WalletDir string `json:"WalletDir"`
}

type FsConfig struct {
	FsRepoRoot   string `json:"FsRepoRoot"`
	FsFileRoot   string `json:"FsFileRoot"`
	FsType       int    `json:"FsType"`
	FsGCPeriod   string `json:"FsGCPeriod"`
	FsMaxStorage string `json:"FsMaxStorage"`
	EnableBackup bool   `json:"EnableBackup"`
}

type DspConfig struct {
	BlockConfirm uint32 `json:"BlockConfirm"`

	ChannelClientType    string `json:"ChannelClientType"`
	ChannelRevealTimeout string `json:"ChannelRevealTimeout"`
	ChannelSettleTimeout string `json:"ChannelSettleTimeout"`
	BlockDelay           string `json:"BlockDelay"`
	MaxUnpaidPayment     int32  `json:"MaxUnpaidPayment"`

	AutoSetupDNSEnable bool     `json:"AutoSetupDNSEnable"`
	DnsNodeMaxNum      int      `json:"DnsNodeMaxNum"`
	DnsChannelDeposit  uint64   `json:"DnsChannelDeposit"`
	DNSWalletAddrs     []string `json:"DNSWalletAddrs"`
	HealthCheckDNS     bool     `json:"HealthCheckDNS"`
	SeedInterval       int      `json:"SeedInterval"`
	Trackers           []string `json:"Trackers"`

	MaxUploadTask   uint32 `json:"MaxUploadTask"`
	MaxDownloadTask uint32 `json:"MaxDownloadTask"`
	MaxShareTask    uint32 `json:"MaxShareTask"`
}

type BootstrapConfig struct {
	Url                 string `json:"Url"`
	MasterChainSupport  bool   `json:"MasterChainSupport"`
	MasterChainRestPort uint32 `json:"MasterChainRestPort"`
	TrackerPort         int    `json:"TrackerPort"`
	ChannelProtocol     string `json:"ChannelProtocol"`
	ChannelPort         uint32 `json:"ChannelPort"`
}

type EdgeConfig struct {
	BaseConfig       BaseConfig        `json:"Base"`
	FsConfig         FsConfig          `json:"Fs"`
	DspConfig        DspConfig         `json:"Dsp"`
	BootstrapsConfig []BootstrapConfig `json:"Bootstraps"`
}

func DefaultConfig() *EdgeConfig {
	configDir = "./" + DEFAULT_CONFIG_FILENAME
	existed := common.FileExisted(configDir)
	if !existed {
		return TestConfig()
	}
	cfg := ParseConfigFromFile(configDir)
	return cfg
}

func TestConfig() *EdgeConfig {
	curPath, _ := filepath.Abs(".")
	return &EdgeConfig{
		BaseConfig: BaseConfig{
			BaseDir:             curPath,
			LogPath:             "./Log",
			ChainId:             "0",
			BlockTime:           5,
			NetworkId:           1567481543,
			PublicIP:            "",
			PortBase:            10000,
			LogLevel:            1,
			LocalRpcPortOffset:  338,
			EnableLocalRpc:      false,
			JsonRpcPortOffset:   336,
			EnableJsonRpc:       true,
			HttpRestPortOffset:  335,
			HttpCertPath:        "",
			HttpKeyPath:         "",
			RestEnable:          true,
			ChannelPortOffset:   3100,
			ChannelProtocol:     "tcp",
			DBPath:              "./DB",
			ChainRestAddrs:      []string{"http://139.219.136.38:20334"},
			ChainRpcAddrs:       []string{"http://139.219.136.38:20336"},
			NATProxyServerAddrs: "tcp://40.73.100.114:6007",
			DspProtocol:         "tcp",
			DspPortOffset:       4100,
			TrackerNetworkId:    1567481543,
			TrackerProtocol:     "tcp",
			TrackerPortOffset:   337,
			WalletDir:           "./wallet.dat",
		},
		DspConfig: DspConfig{
			BlockDelay:           "3",
			BlockConfirm:         2,
			ChannelClientType:    "rpc",
			ChannelRevealTimeout: "20",
			ChannelSettleTimeout: "50",
			MaxUnpaidPayment:     5,
			AutoSetupDNSEnable:   false,
			DnsNodeMaxNum:        100,
			DnsChannelDeposit:    0,
			DNSWalletAddrs:       nil,
			HealthCheckDNS:       true,
			SeedInterval:         600,
			Trackers:             nil,
			MaxUploadTask:        1000,
			MaxDownloadTask:      1000,
			MaxShareTask:         1000,
		},
		FsConfig: FsConfig{
			FsRepoRoot:   "./FS",
			FsFileRoot:   "./Downloads",
			FsType:       0,
			EnableBackup: true,
			FsGCPeriod:   "24h",
			FsMaxStorage: "1T",
		},
	}
}

func LocalConfig() *EdgeConfig {
	curPath, _ := filepath.Abs(".")
	return &EdgeConfig{
		BaseConfig: BaseConfig{
			BaseDir:             curPath,
			LogPath:             "./Log",
			ChainId:             "0",
			BlockTime:           5,
			NetworkId:           1567481545,
			PublicIP:            "",
			PortBase:            10000,
			LogLevel:            1,
			LocalRpcPortOffset:  338,
			EnableLocalRpc:      false,
			JsonRpcPortOffset:   336,
			EnableJsonRpc:       true,
			HttpRestPortOffset:  335,
			HttpCertPath:        "",
			HttpKeyPath:         "",
			RestEnable:          true,
			ChannelPortOffset:   3100,
			ChannelProtocol:     "tcp",
			DBPath:              "./DB",
			ChainRestAddrs:      []string{"http://127.0.0.1:20334"},
			ChainRpcAddrs:       []string{"http://127.0.0.1:20336"},
			NATProxyServerAddrs: "tcp://127.0.0.1:6007",
			DspProtocol:         "tcp",
			DspPortOffset:       4100,
			TrackerNetworkId:    1567481545,
			TrackerProtocol:     "tcp",
			TrackerPortOffset:   337,
			WalletDir:           "./wallet.dat",
		},
		DspConfig: DspConfig{
			BlockDelay:           "3",
			BlockConfirm:         2,
			ChannelClientType:    "rpc",
			ChannelRevealTimeout: "20",
			ChannelSettleTimeout: "50",
			MaxUnpaidPayment:     5,
			AutoSetupDNSEnable:   false,
			DnsNodeMaxNum:        100,
			DnsChannelDeposit:    0,
			DNSWalletAddrs:       nil,
			HealthCheckDNS:       true,
			SeedInterval:         600,
			Trackers:             nil,
			MaxUploadTask:        1000,
			MaxDownloadTask:      1000,
			MaxShareTask:         1000,
		},
		FsConfig: FsConfig{
			FsRepoRoot:   "./FS",
			FsFileRoot:   "./Downloads",
			FsType:       0,
			EnableBackup: true,
			FsGCPeriod:   "24h",
			FsMaxStorage: "1T",
		},
	}
}

//current testnet config config
var Parameters = DefaultConfig()
var configDir string
var curUsrWalAddr string

func UserLocalCfg() {
	Parameters = LocalConfig()
}

func setConfigByCommandParams(dspConfig *EdgeConfig, ctx *cli.Context) {
	///////////////////// protocol setting ///////////////////////////
	if ctx.GlobalIsSet(flags.GetFlagName(flags.ProtocolFsRepoRootFlag)) {
		dspConfig.FsConfig.FsRepoRoot = ctx.String(flags.GetFlagName(flags.ProtocolFsRepoRootFlag))
	}
	if ctx.GlobalIsSet(flags.GetFlagName(flags.ProtocolFsFileRootFlag)) {
		dspConfig.FsConfig.FsFileRoot = ctx.String(flags.GetFlagName(flags.ProtocolFsFileRootFlag))
	}

	if ctx.GlobalIsSet(flags.GetFlagName(flags.ProtocolListenPortOffsetFlag)) {
		dspConfig.BaseConfig.DspPortOffset = ctx.Int(flags.GetFlagName(flags.ProtocolListenPortOffsetFlag))
	}

}

func SetDspConfig(ctx *cli.Context) {
	setConfigByCommandParams(Parameters, ctx)
}

func Init(ctx *cli.Context) {
	if ctx.GlobalIsSet(flags.GetFlagName(flags.ConfigFlag)) {
		path := ctx.String(flags.GetFlagName(flags.ConfigFlag))
		if strings.Contains(path, ".json") {
			configDir = path
		} else {
			configDir = ctx.String(flags.GetFlagName(flags.ConfigFlag)) + "/" + DEFAULT_CONFIG_FILENAME
		}
	} else {
		configDir = "./" + DEFAULT_CONFIG_FILENAME
	}
	existed := common.FileExisted(configDir)
	if !existed {
		log.Infof("config file is not exist: %s, use default config", configDir)
		return
	}
	Parameters = ParseConfigFromFile(configDir)
}

// ParseConfigFromFile. parse config from json file
func ParseConfigFromFile(file string) *EdgeConfig {
	cfg := &EdgeConfig{}
	common.GetJsonObjectFromFile(file, cfg)
	SetDefaultFieldForConfig(cfg)
	return cfg
}

// SetDefaultFieldForConfig. set up default value for some field if missing
func SetDefaultFieldForConfig(cfg *EdgeConfig) {
	if cfg == nil {
		return
	}

	if len(cfg.DspConfig.BlockDelay) == 0 {
		cfg.DspConfig.BlockDelay = fmt.Sprintf("%d", common.BLOCK_DELAY)
	}

	if cfg.DspConfig.BlockConfirm == 0 {
		cfg.DspConfig.BlockConfirm = common.BLOCK_CONFIRM
	}

	if len(cfg.BaseConfig.TrackerProtocol) == 0 {
		cfg.BaseConfig.TrackerProtocol = common.DEFAULT_TRACKER_PROTOCOL
	}
	if cfg.BaseConfig.TrackerNetworkId == 0 {
		cfg.BaseConfig.TrackerNetworkId = cfg.BaseConfig.NetworkId
	}
	if cfg.BaseConfig.TrackerPortOffset == 0 {
		cfg.BaseConfig.TrackerPortOffset = common.DEFAULT_TRACKER_PORT_OFFSET
	}

	if cfg.BaseConfig.WsPortOffset == 0 {
		cfg.BaseConfig.WsPortOffset = common.DEFAULT_WS_PORT_OFFSET
	}

	if len(cfg.DspConfig.ChannelClientType) == 0 {
		cfg.DspConfig.ChannelClientType = "rpc"
	}
	if len(cfg.DspConfig.ChannelRevealTimeout) == 0 {
		cfg.DspConfig.ChannelRevealTimeout = "20"
	}

	if len(cfg.DspConfig.ChannelSettleTimeout) == 0 {
		cfg.DspConfig.ChannelSettleTimeout = "50"
	}
	if cfg.DspConfig.MaxUnpaidPayment == 0 {
		cfg.DspConfig.MaxUnpaidPayment = common.DEFAULT_MAX_UNPAID_PAYMENT
	}

	if cfg.BaseConfig.LogMaxSize == 0 {
		cfg.BaseConfig.LogMaxSize = common.DEFAULT_MAX_LOG_SIZE
	}

	if cfg.DspConfig.DnsNodeMaxNum == 0 {
		cfg.DspConfig.DnsNodeMaxNum = common.DEFAULT_MAX_DNS_NODE_NUM
	}

	if cfg.DspConfig.SeedInterval == 0 {
		cfg.DspConfig.SeedInterval = common.DEFAULT_SEED_INTERVAL
	}
	if cfg.DspConfig.MaxUploadTask == 0 {
		cfg.DspConfig.MaxUploadTask = common.DEFAULT_MAX_UPLOAD_TASK_NUM
	}
	if cfg.DspConfig.MaxDownloadTask == 0 {
		cfg.DspConfig.MaxDownloadTask = common.DEFAULT_MAX_DOWNLOAD_TASK_NUM
	}
	if cfg.DspConfig.MaxShareTask == 0 {
		cfg.DspConfig.MaxShareTask = common.DEFAULT_MAX_SHARE_TASK_NUM
	}
	if cfg.BaseConfig.ProfilePortOffset == 0 {
		cfg.BaseConfig.ProfilePortOffset = 332
	}
}

func GetConfigFromFile(cfgFileName string) *EdgeConfig {
	dir, _ := filepath.Split(configDir)
	path := filepath.Join(dir, cfgFileName)
	exist := common.FileExisted(path)
	if !exist {
		return nil
	}
	newCfg := ParseConfigFromFile(path)
	return newCfg
}

func SwitchConfig(cfgFileName string) error {
	// configDir
	newCfg := GetConfigFromFile(cfgFileName)
	if newCfg == nil {
		return fmt.Errorf("config file not exist: %s", cfgFileName)
	}
	dir, _ := filepath.Split(configDir)
	configDir = filepath.Join(dir, cfgFileName)
	newCfg.BaseConfig.BaseDir = Parameters.BaseConfig.BaseDir
	Parameters = newCfg
	return nil
}

func Save() error {
	data, err := json.MarshalIndent(Parameters, "", "\t")
	if err != nil {
		return err
	}
	err = os.Remove(configDir)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configDir, data, 0666)
}

func SetCurrentUserWalletAddress(addr string) {
	curUsrWalAddr = addr
}

func BaseDataDirPath() string {
	chainId := Parameters.BaseConfig.ChainId
	if len(chainId) == 0 {
		chainId = "0"
	}
	return filepath.Join(Parameters.BaseConfig.BaseDir, "Chain-"+chainId)
}

func BlockTime() uint64 {
	if Parameters.BaseConfig.BlockTime > 0 {
		return Parameters.BaseConfig.BlockTime
	}
	return uint64(common.BLOCK_TIME)
}

// WalletDatFilePath. wallet.dat file path
func WalletDatFilePath() string {
	if common.IsAbsPath(Parameters.BaseConfig.WalletDir) {
		return Parameters.BaseConfig.WalletDir
	}
	return filepath.Join(Parameters.BaseConfig.BaseDir, Parameters.BaseConfig.WalletDir)
}

// ClientSqliteDBPath.
func ClientSqliteDBPath() string {
	return filepath.Join(BaseDataDirPath(), Parameters.BaseConfig.DBPath, curUsrWalAddr, common.EDGE_DB_NAME)
}

// DspDBPath. dsp database path
func DspDBPath() string {
	return filepath.Join(BaseDataDirPath(), Parameters.BaseConfig.DBPath, curUsrWalAddr, common.DSP_DB_NAME)
}

// ChannelDBPath. channel database path
func ChannelDBPath() string {
	return filepath.Join(BaseDataDirPath(), Parameters.BaseConfig.DBPath, curUsrWalAddr, common.PYLONS_DB_NAME)
}

// FsRepoRootPath. fs repo root path
func FsRepoRootPath() string {
	return filepath.Join(BaseDataDirPath(), Parameters.FsConfig.FsRepoRoot, curUsrWalAddr)
}

// FsFileRootPath. fs filestore root path
func FsFileRootPath() string {
	if filepath.IsAbs(Parameters.FsConfig.FsFileRoot) {
		return Parameters.FsConfig.FsFileRoot
	}
	return filepath.Join(BaseDataDirPath(), Parameters.FsConfig.FsFileRoot, curUsrWalAddr)
}

func WsEnabled() bool {
	if Parameters.BaseConfig.WsPortOffset == 0 {
		return false
	}
	return true
}
