package dsp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/saveio/max/max"
	"github.com/saveio/pylons"

	carNet "github.com/saveio/carrier/network"
	dsp "github.com/saveio/dsp-go-sdk"
	dspActorClient "github.com/saveio/dsp-go-sdk/actor/client"
	dspCom "github.com/saveio/dsp-go-sdk/common"
	dspCfg "github.com/saveio/dsp-go-sdk/config"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp/storage"
	"github.com/saveio/edge/p2p/actor/req"
	p2p_actor "github.com/saveio/edge/p2p/actor/server"
	"github.com/saveio/edge/p2p/network"
	"github.com/saveio/themis-go-sdk/wallet"
	"github.com/saveio/themis/account"
	chainCom "github.com/saveio/themis/common"
	chainCfg "github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"

	"github.com/saveio/dsp-go-sdk/utils"
	tkActClient "github.com/saveio/scan/p2p/actor/tracker/client"
	tkActServer "github.com/saveio/scan/p2p/actor/tracker/server"
	tk_net "github.com/saveio/scan/p2p/networks/tracker"
	"github.com/saveio/scan/service/tk"
	chainSdk "github.com/saveio/themis-go-sdk/utils"
)

var DspService *Endpoint

type Endpoint struct {
	Dsp               *dsp.Dsp
	Account           *account.Account
	Password          string
	AccountLabel      string
	progress          sync.Map
	sqliteDB          *storage.SQLiteStorage
	closeCh           chan struct{}
	p2pActor          *p2p_actor.P2PActor
	dspNet            *network.Network ``
	dspPublicAddr     string
	channelNet        *network.Network
	channelPublicAddr string
	eventHub          *EventHub
}

func Init(walletDir, pwd string) (*Endpoint, error) {
	e := &Endpoint{
		closeCh:  make(chan struct{}, 1),
		eventHub: NewEventHub(),
	}
	log.Debugf("walletDir: %s, %d", walletDir, len(walletDir))
	if len(walletDir) == 0 {
		return e, nil
	}
	_, err := os.Open(walletDir)
	if err != nil {
		return e, nil
	}
	wallet, err := wallet.OpenWallet(walletDir)
	if err != nil {
		log.Error("Client open wallet Error, msg:", err)
		return nil, err
	}
	acc, err := wallet.GetDefaultAccount([]byte(pwd))
	if err != nil {
		log.Error("Client get default account Error, msg:", err)
		return nil, err
	}
	accData, err := wallet.GetDefaultAccountData()
	if err != nil {
		return nil, err
	}
	e.Account = acc
	e.AccountLabel = accData.Label
	e.Password = pwd
	config.SetCurrentUserWalletAddress(e.Account.Address.ToBase58())
	log.Debug("Endpoint init success")
	DspService = e
	return e, nil
}

func StartDspNode(endpoint *Endpoint, startListen, startShare, startChannel bool) error {
	listenHost := "127.0.0.1"
	if len(config.Parameters.BaseConfig.PublicIP) > 0 {
		listenHost = config.Parameters.BaseConfig.PublicIP
	}
	channelListenAddr := fmt.Sprintf("%s:%d", listenHost, int(config.Parameters.BaseConfig.PortBase+uint32(config.Parameters.BaseConfig.ChannelPortOffset)))
	log.Debugf("config.Parameters.BaseConfig.ChainRpcAddrs: %v", config.Parameters.BaseConfig.ChainRpcAddrs)
	dspConfig := &dspCfg.DspConfig{
		DBPath:               config.DspDBPath(),
		FsRepoRoot:           config.FsRepoRootPath(),
		FsFileRoot:           config.FsFileRootPath(),
		FsType:               dspCfg.FSType(config.Parameters.FsConfig.FsType),
		FsGcPeriod:           config.Parameters.FsConfig.FsGCPeriod,
		EnableBackup:         config.Parameters.FsConfig.EnableBackup,
		ChainRpcAddrs:        config.Parameters.BaseConfig.ChainRpcAddrs,
		BlockConfirm:         config.Parameters.BaseConfig.BlockConfirm,
		ChannelClientType:    config.Parameters.BaseConfig.ChannelClientType,
		ChannelListenAddr:    channelListenAddr,
		ChannelProtocol:      config.Parameters.BaseConfig.ChannelProtocol,
		ChannelDBPath:        config.ChannelDBPath(),
		ChannelRevealTimeout: config.Parameters.BaseConfig.ChannelRevealTimeout,
		ChannelSettleTimeout: config.Parameters.BaseConfig.ChannelSettleTimeout,
		MaxUnpaidPayment:     config.Parameters.BaseConfig.MaxUnpaidPayment,
		BlockDelay:           config.Parameters.BaseConfig.BlockDelay,
		AutoSetupDNSEnable:   config.Parameters.BaseConfig.AutoSetupDNSEnable,
		DnsNodeMaxNum:        config.Parameters.BaseConfig.DnsNodeMaxNum,
		SeedInterval:         config.Parameters.BaseConfig.SeedInterval,
		DnsChannelDeposit:    config.Parameters.BaseConfig.DnsChannelDeposit,
		Trackers:             config.Parameters.BaseConfig.Trackers,
		TrackerProtocol:      config.Parameters.BaseConfig.TrackerProtocol,
		DNSWalletAddrs:       config.Parameters.BaseConfig.DNSWalletAddrs,
	}
	log.Debugf("dspConfig.dbPath %v, repo: %s, channelDB: %s, wallet: %s, enable backup: %t", dspConfig.DBPath, dspConfig.FsRepoRoot, dspConfig.ChannelDBPath, config.WalletDatFilePath(), config.Parameters.FsConfig.EnableBackup)
	if err := dspCom.CreateDirIfNeed(config.ClientSqliteDBPath()); err != nil {
		return err
	}
	sqliteDB, err := storage.NewSQLiteStorage(config.ClientSqliteDBPath() + "/" + common.SQLITE_DB_NAME)
	if err != nil {
		log.Errorf("SQLite err %s", err)
		return err
	}
	endpoint.sqliteDB = sqliteDB

	//Skip init fs if Dsp doesn't start listen
	if !startListen {
		dspConfig.FsRepoRoot = ""
	}
	if !startChannel {
		dspConfig.ChannelListenAddr = ""
	}
	p2pActor, err := p2p_actor.NewP2PActor()
	if err != nil {
		return err
	}
	endpoint.p2pActor = p2pActor
	dspSrv := dsp.NewDsp(dspConfig, endpoint.Account, p2pActor.GetLocalPID())
	if dspSrv == nil {
		return errors.New("dsp server init failed")
	}
	endpoint.Dsp = dspSrv
	version, _ := endpoint.GetNodeVersion()
	if startListen {
		// start tracker net
		tkListenPort := int(config.Parameters.BaseConfig.PortBase + uint32(config.Parameters.BaseConfig.TrackerPortOffset))
		tkListenAddr := fmt.Sprintf("%s://%s:%d", config.Parameters.BaseConfig.TrackerProtocol, listenHost, tkListenPort)
		log.Debugf("TrackerProtocol: %v, listenAddr: %s", config.Parameters, tkListenAddr)
		if err := endpoint.startTrackerP2P(tkListenAddr, endpoint.Account); err != nil {
			return err
		}

		// start dsp net
		dspListenPort := int(config.Parameters.BaseConfig.PortBase + uint32(config.Parameters.BaseConfig.DspPortOffset))
		dspListenAddr := fmt.Sprintf("%s://%s:%d", config.Parameters.BaseConfig.DspProtocol, listenHost, dspListenPort)
		if err := endpoint.startDspP2P(dspListenAddr, endpoint.Account); err != nil {
			return err
		}
		log.Debugf("start dsp at %s", endpoint.dspPublicAddr)
		// start channel net
		listenAddr := fmt.Sprintf("%s://%s", config.Parameters.BaseConfig.ChannelProtocol, dspConfig.ChannelListenAddr)
		if err := endpoint.startChannelP2P(listenAddr, endpoint.Account); err != nil {
			return err
		}
		log.Debugf("start channel at %s", endpoint.channelPublicAddr)
		endpoint.updateStorageNodeHost()
		if err := endpoint.registerChannelEndpoint(); err != nil {
			return err
		}
		log.Debugf("update node finished")
		if endpoint.Account == nil {
			return errors.New("account is nil")
		}
		// setup filter block range before start
		endpoint.SetFilterBlockRange()
		if err := dspSrv.Start(); err != nil {
			return err
		}
		if startShare {
			// TODO: price needed to be discuss
			dspSrv.SetUnitPriceForAllFile(dspCom.ASSET_USDT, common.DSP_DOWNLOAD_UNIT_PRICE)
			endpoint.Dsp.PushLocalFilesToTrackers()
			go dspSrv.StartShareServices()
		}

	}
	// start channel only
	if !startListen && startChannel {
		if err := dspSrv.StartChannelService(); err != nil {
			return err
		}
	}

	endpoint.closeCh = make(chan struct{}, 1)
	if startListen {
		go endpoint.setupDNSNodeBackground()
		go endpoint.RegisterProgressCh()
		go endpoint.RegisterShareNotificationCh()
	}
	go endpoint.stateChangeService()
	log.Infof("edge start success. version: %s, block time: %d", version, config.BlockTime())
	log.Infof("dsp-go-sdk version: %s", dsp.Version)
	log.Infof("edge version: %s", Version)
	log.Infof("pylons version: %s", pylons.Version)
	log.Infof("max version: %s", max.Version)
	log.Infof("carrier version: %s", carNet.Version)
	go endpoint.notifyWhenStartup()
	return nil
}

// Stop. stop endpoint instance
func (e *Endpoint) Stop() error {
	if e.p2pActor != nil {
		err := e.p2pActor.Stop()
		if err != nil {
			return err
		}
	}
	if e.closeCh != nil {
		close(e.closeCh)
	}
	err := e.sqliteDB.Close()
	if err != nil {
		return err
	}
	e.ResetChannelProgress()
	return e.Dsp.Stop()
}

func (e *Endpoint) startDspP2P(dspListenAddr string, acc *account.Account) error {
	bPub := keypair.SerializePublicKey(acc.PubKey())
	networkKey := utils.NewNetworkEd25519KeyPair(bPub, []byte("dsp"))
	dspNetwork := network.NewP2P()
	dspNetwork.SetNetworkKey(networkKey)
	dspNetwork.SetHandler(e.Dsp.Receive)
	f := utils.TimeoutFunc(func() error {
		return dspNetwork.Start(dspListenAddr)
	})
	err := utils.DoWithTimeout(f, time.Duration(common.START_P2P_TIMEOUT)*time.Second)
	if err != nil {
		return err
	}

	e.p2pActor.SetDspNetwork(dspNetwork)
	e.dspNet = dspNetwork
	e.dspPublicAddr = dspNetwork.PublicAddr()
	return nil
}

func (e *Endpoint) startChannelP2P(channelListenAddr string, acc *account.Account) error {
	channelNetwork := network.NewP2P()
	bPub := keypair.SerializePublicKey(acc.PubKey())
	channelNetwork.Keys = utils.NewNetworkEd25519KeyPair(bPub, []byte("channel"))
	req.SetChannelPid(e.Dsp.Channel.GetChannelPid())
	f := utils.TimeoutFunc(func() error {
		return channelNetwork.Start(channelListenAddr)
	})
	err := utils.DoWithTimeout(f, time.Duration(common.START_P2P_TIMEOUT)*time.Second)
	if err != nil {
		return err
	}
	e.p2pActor.SetChannelNetwork(channelNetwork)
	e.channelPublicAddr = channelNetwork.PublicAddr()
	e.channelNet = channelNetwork
	return nil
}

func (e *Endpoint) startTrackerP2P(tkListenAddr string, acc *account.Account) error {
	tkSrc := tk.NewTrackerService(nil, acc.PublicKey, func(raw []byte) ([]byte, error) {
		return chainSdk.Sign(acc, raw)
	})
	tkActServer, err := tkActServer.NewTrackerActor(tkSrc)
	if err != nil {
		return err
	}
	dPub := keypair.SerializePublicKey(acc.PubKey())
	tkNetworkKey := utils.NewNetworkEd25519KeyPair(dPub, []byte("tk"))
	tkNet := tk_net.NewP2P()
	tkNet.SetNetworkKey(tkNetworkKey)
	tkNet.SetPID(tkActServer.GetLocalPID())
	tkNet.SetProxyServer(config.Parameters.BaseConfig.NATProxyServerAddrs)
	log.Debugf("goto start tk network %s", tkListenAddr)
	tk_net.TkP2p = tkNet
	tkActServer.SetNetwork(tkNet)
	tkActClient.SetTrackerServerPid(tkActServer.GetLocalPID())
	e.p2pActor.SetTrackerNet(tkActServer)

	f := utils.TimeoutFunc(func() error {
		err := tkNet.Start(tkListenAddr, config.Parameters.BaseConfig.TrackerNetworkId)
		if err != nil {
			return err
		}
		log.Debugf("tk network started, public ip %s", tkNet.PublicAddr())
		return nil
	})
	return utils.DoWithTimeout(f, time.Duration(common.START_P2P_TIMEOUT)*time.Second)
}

func (e *Endpoint) registerChannelEndpoint() error {
	walletAddr := e.Dsp.Account.Address
	publicAddr := e.channelNet.PublicAddr()
	for i := 0; i < common.MAX_REG_CHANNEL_TIMES; i++ {
		err := e.Dsp.RegNodeEndpoint(walletAddr, publicAddr)
		log.Debugf("register endpoint for channel %s, err %s", publicAddr, err)
		if err == nil {
			return nil
		}
		log.Errorf("register endpoint failed %s", err)
		select {
		case <-time.After(time.Duration(common.MAX_REG_CHANNEL_BACKOFF) * time.Second):
			continue
		case <-e.closeCh:
			return errors.New("stop register endpoint because edge is closed")
		}
	}
	log.Errorf("register channel endpoint timeout")
	return fmt.Errorf("register channel endpoint timeout")
}

func (e *Endpoint) updateStorageNodeHost() {
	nodeInfo, err := e.NodeQuery(e.Account.Address.ToBase58())
	if err != nil || nodeInfo == nil {
		return
	}
	publicAddr := dspActorClient.P2pGetPublicAddr()
	log.Debugf("update node info %s %s", string(nodeInfo.NodeAddr), publicAddr)
	if string(nodeInfo.NodeAddr) == publicAddr {
		log.Debugf("no need to update")
		return
	}
	log.Debugf("update it %v %v %v", publicAddr, nodeInfo.Volume, nodeInfo.ServiceTime)
	_, err = e.NodeUpdate(publicAddr, nodeInfo.Volume, nodeInfo.ServiceTime)
	log.Debugf("update node result")
	if err != nil {
		log.Errorf("update node addr failed, err %s", err)
	}
}

// SetupDNSNodeBackground. setup a dns node background when received first payments.
func (e *Endpoint) setupDNSNodeBackground() {
	if !config.Parameters.BaseConfig.AutoSetupDNSEnable {
		return
	}
	allChannels := e.Dsp.Channel.GetAllPartners()
	log.Debugf("setup dns node len %v", len(allChannels))
	setup := func() bool {
		syncing, derr := e.IsChannelProcessBlocks()
		// progress, derr := e.GetFilterBlockProgress()
		if derr != nil {
			log.Errorf("setup dns channel syncing block err %s", derr)
			return false
		}
		if syncing {
			log.Debugf("setup dns wait for init channel")
			return false
		}

		if len(allChannels) > 0 {
			err := e.Dsp.SetupDNSChannels()
			if err != nil {
				log.Errorf("SetupDNSChannels err %s", err)
				return false
			}
			return true
		}
		address, err := chainCom.AddressFromBase58(e.Dsp.WalletAddress())
		if err != nil {
			log.Errorf("setup dns failed address decoded failed")
			return false
		}
		bal, err := e.Dsp.Chain.Native.Usdt.BalanceOf(address)
		if bal == 0 || err != nil {
			log.Errorf("setup dns failed balance is 0")
			return false
		}
		if bal < chainCfg.DEFAULT_GAS_PRICE*chainCfg.DEFAULT_GAS_PRICE {
			log.Errorf("setup dns failed balance not enough %d", bal)
			return false
		}
		err = e.Dsp.SetupDNSChannels()
		if err != nil {
			log.Errorf("setup dns channel err %s", err)
			return false
		}
		return true
	}
	if setup() {
		return
	}
	ti := time.NewTicker(time.Duration(30) * time.Second)
	for {
		select {
		case <-ti.C:
			if setup() {
				return
			}
		case <-e.closeCh:
			ti.Stop()
			return
		}
	}
}

func (e *Endpoint) stateChangeService() {
	log.Debugf("start stateChangeService")
	ti := time.NewTicker(time.Duration(common.MAX_STATE_CHANGE_CHECK_INTERVAL) * time.Second)
	for {
		select {
		case <-ti.C:
			// check log file size
			go e.checkGoRoutineNum()
			go e.checkLogFileSize()
			if e.dspPublicAddr != e.dspNet.PublicAddr() {
				log.Debugf("dsp public address has change, old addr: %s, new addr:%s", e.dspPublicAddr, e.dspNet.PublicAddr())
				e.dspPublicAddr = e.dspNet.PublicAddr()
				go e.updateStorageNodeHost()
			}
			if e.channelPublicAddr != e.channelNet.PublicAddr() {
				log.Debugf("channel public address has change, old addr: %s, new addr:%s", e.channelPublicAddr, e.channelNet.PublicAddr())
				e.channelPublicAddr = e.channelNet.PublicAddr()
				go e.registerChannelEndpoint()
			}
			go e.checkOnlineDNS()
			go e.notifyChannelProgress()
			go e.notifyNewSmartContractEvent()
			go e.notifyNewNetworkState()
		case <-e.closeCh:
			ti.Stop()
			return
		}
	}
}

func (e *Endpoint) checkGoRoutineNum() {
	log.Debugf("go routine num: %d", runtime.NumGoroutine())
}

func (e *Endpoint) checkLogFileSize() {
	isNeedNewFile := log.CheckIfNeedNewFile()
	if !isNeedNewFile {
		return
	}
	log.ClosePrintLog()
	logPath := config.Parameters.BaseConfig.LogPath
	baseDir := config.Parameters.BaseConfig.BaseDir
	extra := ""
	logFullPath := filepath.Join(baseDir, logPath) + extra + "/"
	_, err := log.FileOpen(logFullPath)
	if err != nil {
		extra = strconv.FormatUint(utils.GetMilliSecTimestamp(), 10)
	}
	logFullPath = filepath.Join(baseDir, logPath) + extra + "/"
	log.InitLog(int(config.Parameters.BaseConfig.LogLevel), logFullPath, log.Stdout)
}

func (e *Endpoint) checkOnlineDNS() {
	if e.Dsp.DNS == nil || e.Dsp.DNS.DNSNode == nil {
		e.Dsp.BootstrapDNS()
	}
}
