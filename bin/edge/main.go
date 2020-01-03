package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/saveio/dsp-go-sdk/utils"
	"github.com/saveio/edge/cmd"
	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp"
	"github.com/saveio/edge/dsp/actor/client"
	"github.com/saveio/edge/event/actor/server"
	"github.com/saveio/edge/http"
	"github.com/saveio/edge/http/jsonrpc"
	"github.com/saveio/edge/http/localrpc"
	"github.com/saveio/edge/http/websocket"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/common/password"
	"github.com/urfave/cli"
)

func initAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "save dsp"
	app.Action = dspInit
	app.Version = config.VERSION
	app.Copyright = "Copyright in 2018 The SAVE Authors"
	app.Commands = []cli.Command{
		cmd.AccountCommand,
		cmd.InfoCommand,
		cmd.AssetCommand,
		cmd.FileCommand,
		cmd.UserspaceCommand,
		cmd.NodeCommand,
		cmd.DnsCommand,
		cmd.ChannelCommand,
		cmd.VersionCommand,
	}
	app.Flags = []cli.Flag{
		flags.ProtocolListenPortOffsetFlag,
		flags.ProtocolFsRepoRootFlag,
		flags.ProtocolFsFileRootFlag,
		flags.ConfigFlag,
		flags.LaunchManualFlag,
		flags.WalletPasswordFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		config.Init(context)
		return nil
	}
	return app
}

func main() {
	if err := initAPP().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func dspInit(ctx *cli.Context) {
	launchManual := ctx.Bool(flags.GetFlagName(flags.LaunchManualFlag))

	var walletPwd string
	if !launchManual && common.FileExisted(config.WalletDatFilePath()) {
		if ctx.IsSet(flags.GetFlagName(flags.WalletPasswordFlag)) {
			walletPwd = ctx.String(flags.GetFlagName(flags.WalletPasswordFlag))
		} else {
			pwd, err := password.GetPassword()
			if err != nil {
				log.Errorf("require password: %s", err.Error())
				os.Exit(1)
			}
			walletPwd = string(pwd)
		}
	}
	config.SetDspConfig(ctx)
	initLog(ctx)
	log.Debugf("set dsp config, config %v", config.Parameters)

	eventActorServer, _ := server.NewEventActorServer()
	client.SetEventPid(eventActorServer.GetLocalPID())

	initRest()
	initWebsocket()
	initJsonRpc()

	if launchManual {
		waitToExit()
		return
	}
	endpoint, err := dsp.Init(config.WalletDatFilePath(), walletPwd)
	if endpoint == nil {
		log.Error("dsp init failed: %s", err.Error())
		os.Exit(1)
	}
	if endpoint.Account != nil {
		if err := dsp.StartDspNode(endpoint, true, true, true); err != nil {
			log.Errorf("start dsp node err %s", err.Error())
		}
		if err := initLocalRpc(); err != nil {
			log.Errorf("init local rpc err %s", err.Error())
		}
	} else {
		log.Infof("current wallet is empty, please create one")
	}
	waitToExit()
}

func initLog(ctx *cli.Context) {
	//init log module
	logLevel := config.Parameters.BaseConfig.LogLevel
	logPath := config.Parameters.BaseConfig.LogPath
	baseDir := config.Parameters.BaseConfig.BaseDir
	if len(logPath) == 0 {
		logPath = fmt.Sprintf("./Log_%s", time.Now().Format("2006-01-02"))
	}

	extra := ""
	logFullPath := filepath.Join(baseDir, logPath) + extra + "/"
	_, err := log.FileOpen(logFullPath)
	if err != nil {
		extra = strconv.FormatUint(utils.GetMilliSecTimestamp(), 10)
	}
	logFullPath = filepath.Join(baseDir, logPath) + extra + "/"
	log.InitLog(logLevel, logFullPath, log.Stdout)
	log.Infof("start logging at %s", logFullPath)
}

func initRest() {
	if !config.Parameters.BaseConfig.RestEnable {
		return
	}
	go http.StartRestServer()

	log.Info("Restful init success")
}

func initWebsocket() {
	if !config.WsEnabled() {
		return
	}
	go websocket.StartServer()
}

func initJsonRpc() {
	if !config.Parameters.BaseConfig.EnableJsonRpc {
		return
	}
	go jsonrpc.StartRPCServer()
	log.Info("JsonRpc init success")
}

func waitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("do received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}

func initLocalRpc() error {
	if !config.Parameters.BaseConfig.EnableLocalRpc {
		return nil
	}
	var err error
	exitCh := make(chan interface{}, 0)
	go func() {
		err = localrpc.StartLocalRpcServer()
		close(exitCh)
	}()

	flag := false
	select {
	case <-exitCh:
		if !flag {
			return err
		}
	case <-time.After(time.Millisecond * 5):
		flag = true
	}

	log.Infof("Local rpc init success")
	return nil
}
