package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/common"
	"github.com/saveio/themis/common/log"

	"github.com/urfave/cli"
)

//default common parameter
const (
	VERSION              = "0.1"
	DEFAULT_MAX_LOG_SIZE = 20 * 1024 * 1024 //MB
	DEFAULT_CONFIG_DIR   = "/config.json"
)

type BaseConfig struct {
	BaseDir            string `json:"BaseDir"`
	LogPath            string `json:"LogPath"`
	NetworkId          uint32 `json:"NetworkId"`
	PublicIP           string `json:"PublicIP"`
	PortBase           uint32 `json:"PortBase"`
	LogLevel           int    `json:"LogLevel"`
	LocalRpcPortOffset int    `json:"LocalRpcPortOffset"`
	EnableLocalRpc     bool   `json:"EnableLocalRpc"`
	HttpRestPortOffset int    `json:"HttpRestPortOffset"`
	HttpCertPath       string `json:"HttpCertPath"`
	HttpKeyPath        string `json:"HttpKeyPath"`
	RestEnable         bool   `json:"RestEnable"`

	ChannelPortOffset    int    `json:"ChannelPortOffset"`
	ChannelProtocol      string `json:"ChannelProtocol"`
	ChannelClientType    string `json:"ChannelClientType"`
	ChannelRevealTimeout string `json:"ChannelRevealTimeout"`

	DBPath        string   `json:"DBPath"`
	PpListenAddr  string   `json:"PpListenAddr"`
	PpBootstrap   []string `json:"PpBootstrap"`
	ChainRestAddr string   `json:"ChainRestAddr"`
	ChainRpcAddr  string   `json:"ChainRpcAddr"`

	NATProxyServerAddr string `json:"NATProxyServerAddr"`
	DspProtocol        string `json:"DspProtocol"`
	DspPortOffset      int    `json:"DspPortOffset"`

	AutoSetupDNSEnable bool   `json:"AutoSetupDNSEnable"`
	TrackerPortOffset  int    `json:"TrackerPortOffset"`
	DnsNodeMaxNum      int    `json:"DnsNodeMaxNum"`
	SeedInterval       int    `json:"SeedInterval"`
	DnsChannelDeposit  uint64 `json:"DnsChannelDeposit"`

	WalletPwd string `json:"WalletPwd"`
	WalletDir string `json:"WalletDir"`
}

type FsConfig struct {
	FsRepoRoot string `json:"FsRepoRoot"`
	FsFileRoot string `json:"FsFileRoot"`
	FsType     int    `json:"FsType"`
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

func DefConfig() *DspConfig {
	return &DspConfig{
		BaseConfig: BaseConfig{
			BaseDir:              ".",
			LogPath:              "./Log",
			PortBase:             10000,
			LogLevel:             0,
			LocalRpcPortOffset:   338,
			EnableLocalRpc:       true,
			HttpRestPortOffset:   335,
			RestEnable:           true,
			ChannelPortOffset:    3004,
			ChannelProtocol:      "udp",
			ChannelClientType:    "rpc",
			ChannelRevealTimeout: "250",
			DBPath:               "./DB",
			ChainRestAddr:        "http://127.0.0.1:20334",
			ChainRpcAddr:         "http://127.0.0.1:20336",
			NATProxyServerAddr:   "udp://127.0.0.1:6008",
			DspProtocol:          "udp",
			DspPortOffset:        4001,
			AutoSetupDNSEnable:   true,
			DnsNodeMaxNum:        100,
			SeedInterval:         3600,
			DnsChannelDeposit:    1000000000,
			WalletPwd:            "pwd",
			WalletDir:            "./wallet.dat",
		},
		FsConfig: FsConfig{
			FsRepoRoot: "./FS",
			FsFileRoot: "./Downloads",
			FsType:     0,
		},
	}
}

//current default config
var Parameters = DefConfig()
var configDir string
var currUsrWalAddr string

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
		configDir = ctx.String(flags.GetFlagName(flags.ConfigFlag)) + DEFAULT_CONFIG_DIR
	} else {
		configDir = "." + DEFAULT_CONFIG_DIR
	}
	existed := common.FileExisted(configDir)
	if !existed {
		panic(fmt.Sprintf("config file is not exist: %s", configDir))
	}
	log.Debugf("configDir %s", configDir)
	common.GetJsonObjectFromFile(configDir, Parameters)
}

func Save() error {
	data, err := json.Marshal(Parameters)
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
	currUsrWalAddr = addr
}

// WalletDatFilePath. wallet.dat file path
func WalletDatFilePath() string {
	return filepath.Join(Parameters.BaseConfig.BaseDir, Parameters.BaseConfig.WalletDir)
}

// ClientLevelDBPath. todo: replace level db with sqlite db
func ClientLevelDBPath() string {
	return filepath.Join(Parameters.BaseConfig.BaseDir, Parameters.BaseConfig.DBPath, currUsrWalAddr, "client")
}

// ClientSqliteDBPath.
func ClientSqliteDBPath() string {
	return filepath.Join(Parameters.BaseConfig.BaseDir, Parameters.BaseConfig.DBPath, currUsrWalAddr, "oniclient-sqlite.db")
}

// DspDBPath. dsp database path
func DspDBPath() string {
	return filepath.Join(Parameters.BaseConfig.BaseDir, Parameters.BaseConfig.DBPath, currUsrWalAddr, "dsp")
}

// ChannelDBPath. channel database path
func ChannelDBPath() string {
	return filepath.Join(Parameters.BaseConfig.BaseDir, Parameters.BaseConfig.DBPath, currUsrWalAddr, "channel")
}

// FsRepoRootPath. fs repo root path
func FsRepoRootPath() string {
	return filepath.Join(Parameters.BaseConfig.BaseDir, Parameters.FsConfig.FsRepoRoot, currUsrWalAddr)
}

// FsFileRootPath. fs filestore root path
func FsFileRootPath() string {
	return filepath.Join(Parameters.BaseConfig.BaseDir, Parameters.FsConfig.FsFileRoot, currUsrWalAddr)
}
