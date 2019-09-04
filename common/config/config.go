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
	ChainId            string `json:"ChainId"`
	BlockTime          uint64 `json:"BlockTime"`
	NetworkId          uint32 `json:"NetworkId"`
	PublicIP           string `json:"PublicIP"`
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

	ChannelPortOffset    int    `json:"ChannelPortOffset"`
	ChannelProtocol      string `json:"ChannelProtocol"`
	ChannelClientType    string `json:"ChannelClientType"`
	ChannelRevealTimeout string `json:"ChannelRevealTimeout"`
	ChannelSettleTimeout string `json:"ChannelSettleTimeout"`
	BlockDelay           string `json:"BlockDelay"`

	DBPath         string   `json:"DBPath"`
	PpListenAddr   string   `json:"PpListenAddr"`
	PpBootstrap    []string `json:"PpBootstrap"`
	ChainRestAddrs []string `json:"ChainRestAddrs"`
	ChainRpcAddrs  []string `json:"ChainRpcAddrs"`

	NATProxyServerAddrs string `json:"NATProxyServerAddrs"`
	DspProtocol         string `json:"DspProtocol"`
	DspPortOffset       int    `json:"DspPortOffset"`

	AutoSetupDNSEnable bool     `json:"AutoSetupDNSEnable"`
	TrackerPortOffset  int      `json:"TrackerPortOffset"`
	DnsNodeMaxNum      int      `json:"DnsNodeMaxNum"`
	SeedInterval       int      `json:"SeedInterval"`
	DnsChannelDeposit  uint64   `json:"DnsChannelDeposit"`
	Trackers           []string `json:"Trackers"`
	DNSWalletAddrs     []string `json:"DNSWalletAddrs"`

	WalletPwd string `json:"WalletPwd"`
	WalletDir string `json:"WalletDir"`
}

type FsConfig struct {
	FsRepoRoot   string `json:"FsRepoRoot"`
	FsFileRoot   string `json:"FsFileRoot"`
	FsType       int    `json:"FsType"`
	FsGCPeriod   string `json:"FsGCPeriod"`
	EnableBackup bool   `json:"EnableBackup"`
}

type BootstrapConfig struct {
	Url                 string `json:"Url"`
	MasterChainSupport  bool   `json:"MasterChainSupport"`
	MasterChainRestPort uint32 `json:"MasterChainRestPort"`
	TrackerPort         int    `json:"TrackerPort"`
	ChannelProtocol     string `json:"ChannelProtocol"`
	ChannelPort         uint32 `json:"ChannelPort"`
}

type DspConfig struct {
	BaseConfig       BaseConfig        `json:"Base"`
	FsConfig         FsConfig          `json:"Fs"`
	BootstrapsConfig []BootstrapConfig `json:"Bootstraps"`
}

func DefaultConfig() *DspConfig {
	configDir = "./" + DEFAULT_CONFIG_FILENAME
	existed := common.FileExisted(configDir)
	if !existed {
		return TestConfig()
	}
	cfg := &DspConfig{}
	common.GetJsonObjectFromFile(configDir, cfg)
	return cfg
}

func TestConfig() *DspConfig {
	return &DspConfig{
		BaseConfig: BaseConfig{
			BaseDir:              ".",
			LogPath:              "./Log",
			PortBase:             10000,
			LogLevel:             0,
			LocalRpcPortOffset:   205,
			EnableLocalRpc:       false,
			JsonRpcPortOffset:    204,
			EnableJsonRpc:        true,
			HttpRestPortOffset:   203,
			RestEnable:           true,
			ChannelPortOffset:    202,
			ChannelProtocol:      "udp",
			ChannelClientType:    "rpc",
			ChannelRevealTimeout: "250",
			DBPath:               "./DB",
			ChainRestAddrs:       []string{"http://127.0.0.1:20334"},
			ChainRpcAddrs:        []string{"http://127.0.0.1:20336"},
			NATProxyServerAddrs:  "udp://40.73.100.114:6008",
			DspProtocol:          "udp",
			DspPortOffset:        201,
			AutoSetupDNSEnable:   true,
			DnsNodeMaxNum:        100,
			SeedInterval:         3600,
			DnsChannelDeposit:    1000000000,
			WalletPwd:            "pwd",
			WalletDir:            "./wallet.dat",
		},
		FsConfig: FsConfig{
			FsRepoRoot:   "./FS",
			FsFileRoot:   "./Downloads",
			FsType:       0,
			EnableBackup: true,
		},
	}
}

//current testnet config config
var Parameters = DefaultConfig()
var configDir string
var curUsrWalAddr string

func setConfigByCommandParams(dspConfig *DspConfig, ctx *cli.Context) {
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
	common.GetJsonObjectFromFile(configDir, Parameters)
}

func GetConfigFromFile(cfgFileName string) *DspConfig {
	dir, _ := filepath.Split(configDir)
	path := filepath.Join(dir, cfgFileName)
	exist := common.FileExisted(path)
	if !exist {
		return nil
	}
	newCfg := &DspConfig{}
	err := common.GetJsonObjectFromFile(path, newCfg)
	if err != nil {
		return nil
	}
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
	return filepath.Join(Parameters.BaseConfig.BaseDir, Parameters.BaseConfig.WalletDir)
}

// ClientLevelDBPath. todo: replace level db with sqlite db
func ClientLevelDBPath() string {
	return filepath.Join(BaseDataDirPath(), Parameters.BaseConfig.DBPath, curUsrWalAddr, "client")
}

// ClientSqliteDBPath.
func ClientSqliteDBPath() string {
	return filepath.Join(BaseDataDirPath(), Parameters.BaseConfig.DBPath, curUsrWalAddr, "edge-sqlite.db")
}

// DspDBPath. dsp database path
func DspDBPath() string {
	return filepath.Join(BaseDataDirPath(), Parameters.BaseConfig.DBPath, curUsrWalAddr, "dsp")
}

// ChannelDBPath. channel database path
func ChannelDBPath() string {
	return filepath.Join(BaseDataDirPath(), Parameters.BaseConfig.DBPath, curUsrWalAddr, "channel")
}

// FsRepoRootPath. fs repo root path
func FsRepoRootPath() string {
	return filepath.Join(BaseDataDirPath(), Parameters.FsConfig.FsRepoRoot, curUsrWalAddr)
}

// FsFileRootPath. fs filestore root path
func FsFileRootPath() string {
	return filepath.Join(BaseDataDirPath(), Parameters.FsConfig.FsFileRoot, curUsrWalAddr)
}
