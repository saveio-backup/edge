package dsp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	carNet "github.com/saveio/carrier/network"
	"github.com/saveio/carrier/types/opcode"
	dspActorClient "github.com/saveio/dsp-go-sdk/actor/client"
	dspCfg "github.com/saveio/dsp-go-sdk/config"
	"github.com/saveio/dsp-go-sdk/dsp"
	dspSdk "github.com/saveio/dsp-go-sdk/dsp"
	dspNetCom "github.com/saveio/dsp-go-sdk/network/common"
	"github.com/saveio/dsp-go-sdk/network/message/pb"
	"github.com/saveio/dsp-go-sdk/types/state"
	"github.com/saveio/dsp-go-sdk/utils/async"
	"github.com/saveio/dsp-go-sdk/utils/crypto"
	dspOS "github.com/saveio/dsp-go-sdk/utils/os"
	uTime "github.com/saveio/dsp-go-sdk/utils/time"

	"github.com/saveio/dsp-go-sdk/core/chain"
	"github.com/saveio/dsp-go-sdk/core/fs"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp/cache"
	"github.com/saveio/edge/p2p/actor/req"
	p2p_actor "github.com/saveio/edge/p2p/actor/server"
	"github.com/saveio/edge/p2p/network"
	edgeUtils "github.com/saveio/edge/utils"
	"github.com/saveio/max/max"
	"github.com/saveio/pylons"
	"github.com/saveio/pylons/actor/msg_opcode"
	tkActClient "github.com/saveio/scan/p2p/actor/tracker/client"
	tkActServer "github.com/saveio/scan/p2p/actor/tracker/server"
	tk_net "github.com/saveio/scan/p2p/network"
	"github.com/saveio/scan/service/tk"
	chainSdk "github.com/saveio/themis-go-sdk/utils"
	"github.com/saveio/themis-go-sdk/wallet"
	"github.com/saveio/themis/account"
	chainCom "github.com/saveio/themis/common"
	chainCfg "github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
)

var DspService *Endpoint

type Endpoint struct {
	dsp               *dspSdk.Dsp
	account           *account.Account
	password          string
	accountLabel      string
	dspAccLock        *sync.Mutex
	progress          sync.Map
	closeCh           chan struct{}
	p2pActor          *p2p_actor.P2PActor
	dspNet            *network.Network
	dspPublicAddr     string
	channelNet        *network.Network
	channelPublicAddr string
	eventHub          *EventHub
	state             *LifeCycle
	cache             *cache.EdgeCache
}

func Init(walletDir, pwd string) (*Endpoint, error) {
	e := &Endpoint{
		closeCh:    make(chan struct{}, 1),
		eventHub:   NewEventHub(),
		state:      NewLifeCycle(),
		dspAccLock: new(sync.Mutex),
		cache:      cache.NewEdgeCache(),
	}
	DspService = e
	log.Debugf("walletDir: %s, %d", walletDir, len(walletDir))
	if len(walletDir) == 0 {
		return e, nil
	}
	if _, err := os.Open(walletDir); err != nil {
		log.Errorf("open wallet err %s", err)
		return nil, err
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
	e.setDspAccount(acc, accData.Label, pwd)
	config.SetCurrentUserWalletAddress(e.getDspWalletAddress())
	log.Debug("Endpoint init success %s", e.getDspWalletAddress())
	return e, nil
}

func StartDspNode(endpoint *Endpoint, startListen, startShare, startChannel bool) error {
	if err := endpoint.state.Start(); err != nil {
		return err
	}
	defer endpoint.state.Run()

	listenHost := "127.0.0.1"
	if len(config.Parameters.BaseConfig.PublicIP) > 0 {
		listenHost = config.Parameters.BaseConfig.PublicIP
	}
	channelListenAddr := fmt.Sprintf("%s:%d", listenHost,
		int(config.Parameters.BaseConfig.PortBase+uint32(config.Parameters.BaseConfig.ChannelPortOffset)))
	dspListenAddr := fmt.Sprintf("%s:%d", listenHost,
		int(config.Parameters.BaseConfig.PortBase+uint32(config.Parameters.BaseConfig.DspPortOffset)))
	log.Debugf("config.Parameters.BaseConfig.ChainRpcAddrs: %v", config.Parameters.BaseConfig.ChainRpcAddrs)
	dspConfig := &dspCfg.DspConfig{
		DBPath:               config.DspDBPath(),
		FsRepoRoot:           config.FsRepoRootPath(),
		FsFileRoot:           config.FsFileRootPath(),
		DspListenAddr:        dspListenAddr,
		DspProtocol:          config.Parameters.BaseConfig.DspProtocol,
		FsType:               int(config.Parameters.FsConfig.FsType),
		FsGcPeriod:           config.Parameters.FsConfig.FsGCPeriod,
		FsMaxStorage:         config.Parameters.FsConfig.FsMaxStorage,
		EnableBackup:         config.Parameters.FsConfig.EnableBackup,
		ChainRpcAddrs:        config.Parameters.BaseConfig.ChainRpcAddrs,
		BlockConfirm:         config.Parameters.DspConfig.BlockConfirm,
		ChannelClientType:    config.Parameters.DspConfig.ChannelClientType,
		ChannelListenAddr:    channelListenAddr,
		ChannelProtocol:      config.Parameters.BaseConfig.ChannelProtocol,
		ChannelDBPath:        config.ChannelDBPath(),
		ChannelRevealTimeout: config.Parameters.DspConfig.ChannelRevealTimeout,
		ChannelSettleTimeout: config.Parameters.DspConfig.ChannelSettleTimeout,
		MaxUnpaidPayment:     config.Parameters.DspConfig.MaxUnpaidPayment,
		BlockDelay:           config.Parameters.DspConfig.BlockDelay,
		AutoSetupDNSEnable:   config.Parameters.DspConfig.AutoSetupDNSEnable,
		DnsNodeMaxNum:        config.Parameters.DspConfig.DnsNodeMaxNum,
		SeedInterval:         config.Parameters.DspConfig.SeedInterval,
		DnsChannelDeposit:    config.Parameters.DspConfig.DnsChannelDeposit,
		Trackers:             config.Parameters.DspConfig.Trackers,
		TrackerProtocol:      config.Parameters.BaseConfig.TrackerProtocol,
		DNSWalletAddrs:       config.Parameters.DspConfig.DNSWalletAddrs,
		HealthCheckDNS:       config.Parameters.DspConfig.HealthCheckDNS,
		MaxUploadTask:        config.Parameters.DspConfig.MaxUploadTask,
		MaxDownloadTask:      config.Parameters.DspConfig.MaxDownloadTask,
		MaxShareTask:         config.Parameters.DspConfig.MaxShareTask,
		EnableLayer2:         config.Parameters.DspConfig.EnableLayer2,
		AllowLocalNode:       config.Parameters.DspConfig.AllowLocalNode,
		Mode:                 config.Parameters.DspConfig.Mode,
	}
	log.Debugf("dspConfig.dbPath %v, repo: %s, channelDB: %s, wallet: %s, enable backup: %t",
		dspConfig.DBPath, dspConfig.FsRepoRoot, dspConfig.ChannelDBPath, config.WalletDatFilePath(),
		config.Parameters.FsConfig.EnableBackup)
	if err := dspOS.CreateDirIfNeed(config.ClientSqliteDBPath()); err != nil {
		return err
	}

	if err := dspOS.CreateDirIfNeed(config.PlotPath()); err != nil {
		return err
	}
	// Skip init fs if Dsp doesn't start listen
	if !startListen {
		dspConfig.FsRepoRoot = ""
	}
	if !startChannel {
		dspConfig.ChannelListenAddr = ""
	}

	// init DSP
	p2pActor, err := p2p_actor.NewP2PActor()
	if err != nil {
		return err
	}
	p2pActor.SetDNSHostAddrCallback(func(walletAddr string) string {
		hostAddr, err := endpoint.GetDNSHostAddr(walletAddr)
		if err != nil {
			log.Errorf("get dns host addr err %s", err)
		}
		return hostAddr
	})
	endpoint.p2pActor = p2pActor
	dspSrv := dsp.NewDsp(dspConfig, endpoint.GetDspAccount(), p2pActor.GetLocalPID(), dspConfig.Mode)
	log.Debugf("new dsp svr %v", dspSrv)
	if dspSrv == nil {
		return errors.New("dsp server init failed")
	}
	if endpoint.GetDspAccount() == nil {
		return fmt.Errorf("dsp service account is nil")
	}
	endpoint.setDsp(dspSrv)
	// start dsp service
	if startListen {
		go func() {
			if err = endpoint.startDspService(listenHost); err != nil {
				log.Errorf("start dsp err %s", err)
			} else {
				endpoint.notifyWhenStartup()
			}
		}()
	}
	// start share service
	if startListen && startShare {
		go endpoint.startShareService()
	}
	// start channel only
	if !startListen && startChannel {
		go func() {
			if err := dspSrv.StartChannelService(); err != nil {
				log.Errorf("start dsp channel err %s", err)
			}
		}()
	}
	endpoint.closeCh = make(chan struct{}, 1)
	if startListen {
		go endpoint.setupDNSNodeBackground()
		go endpoint.RegisterProgressCh()
		go endpoint.RegisterShareNotificationCh()
	}
	go endpoint.stateChangeService()
	version, _ := endpoint.GetNodeVersion()
	log.Infof("edge start success. version: %s, block time: %d", version, config.BlockTime())
	log.Infof("dsp-go-sdk version: %s", dsp.Version)
	log.Infof("edge version: %s", Version)
	log.Infof("pylons version: %s", pylons.Version)
	log.Infof("max version: %s", max.Version)
	log.Infof("carrier version: %s", carNet.Version)
	return nil
}

func (this *Endpoint) SetDsp(d *dspSdk.Dsp) {
	this.dsp = d
	if this.dspAccLock == nil {
		this.dspAccLock = &sync.Mutex{}
	}
}

// Stop. stop endpoint instance
func (this *Endpoint) Stop() error {
	dsp := this.getDsp()
	if dsp.GetTaskMgr() != nil && dsp.GetTaskMgr().HasRunningTask() {
		log.Warnf("exist running task, cant stop")
		return fmt.Errorf("exist running task, cant stop")
	}
	log.Debugf("stop dsp without running task")
	if err := this.state.Stop(); err != nil {
		log.Errorf("stop dsp state err %s", err)
		return err
	}
	defer this.state.Terminate()
	if this.p2pActor != nil {
		log.Debugf("stop edge p2p module")
		err := this.p2pActor.Stop()
		if err != nil {
			log.Errorf("stop edge p2p err %s", err)
			return err
		}
		log.Debugf("stop edge p2p success")
	}

	if this.closeCh != nil {
		log.Debugf("stop edge with closing channel")
		close(this.closeCh)
	}
	log.Debugf("stop edge with closing channel success")
	this.ResetChannelProgress()
	return this.getDsp().Stop()
}

func (this *Endpoint) startDspService(listenHost string) error {
	// start tracker net
	tkHostAddr := &common.HostAddr{
		Protocol: config.Parameters.BaseConfig.TrackerProtocol,
		Address:  listenHost,
		Port: fmt.Sprintf("%d", int(config.Parameters.BaseConfig.PortBase+
			uint32(config.Parameters.BaseConfig.TrackerPortOffset))),
	}
	if err := this.startTrackerP2P(tkHostAddr, this.GetDspAccount()); err != nil {
		return err
	}
	// start dsp net
	dspHostAddr := &common.HostAddr{
		Protocol: config.Parameters.BaseConfig.DspProtocol,
		Address:  listenHost,
		Port: fmt.Sprintf("%d", int(config.Parameters.BaseConfig.PortBase+
			uint32(config.Parameters.BaseConfig.DspPortOffset))),
	}
	if err := this.startDspP2P(dspHostAddr, this.GetDspAccount()); err != nil {
		return err
	}
	log.Debugf("start dsp at %s", this.dspPublicAddr)
	// start channel net
	chHostAddr := &common.HostAddr{
		Protocol: config.Parameters.BaseConfig.ChannelProtocol,
		Address:  listenHost,
		Port: fmt.Sprintf("%d", int(config.Parameters.BaseConfig.PortBase+
			uint32(config.Parameters.BaseConfig.ChannelPortOffset))),
	}

	if err := this.startChannelP2P(chHostAddr, this.GetDspAccount()); err != nil {
		return err
	}
	log.Debugf("start channel at %s", this.channelPublicAddr)
	this.updateStorageNodeHost()
	log.Debugf("update node finished")
	if this.GetDspAccount() == nil {
		log.Debugf("account is nil")
		return errors.New("account is nil")
	}
	// setup filter block range before start
	this.SetFilterBlockRange()
	go func() {
		for {
			this.notifyChannelProgress()
			if this.getDsp().Running() {
				log.Debugf("return channel progress after runing")
				return
			}
			select {
			case <-time.After(time.Duration(2) * time.Second):
				continue
			case <-this.closeCh:
				return
			}
		}
	}()
	if err := this.getDsp().Start(); err != nil {
		log.Errorf("start dsp err", err)
		return err
	}
	log.Debugf("start dsp success")
	return nil
}

func (this *Endpoint) startDspP2P(hostAddr *common.HostAddr, acc *account.Account) error {
	networkKey := crypto.NewNetworkKeyPairWithAccount(acc)

	codes := make(map[opcode.Opcode]proto.Message)
	codes[dspNetCom.MSG_OP_CODE] = &pb.Message{}
	opts := []network.NetworkOption{
		network.WithKeys(networkKey),
		network.WithMsgHandler(this.getDsp().Receive),
		network.WithNetworkId(config.Parameters.BaseConfig.NetworkId),
		network.WithWalletAddrFromPeerId(crypto.AddressFromPubkeyHex),
		network.WithOpcodes(codes),
	}
	if len(config.Parameters.BaseConfig.IntranetIP) > 0 {
		opts = append(opts, network.WithIntranetIP(config.Parameters.BaseConfig.IntranetIP))
	}
	if len(config.Parameters.BaseConfig.NATProxyServerAddrs) > 0 {
		opts = append(opts, network.WithProxyAddrs(strings.Split(config.Parameters.BaseConfig.NATProxyServerAddrs, ",")))
	}
	dspNetwork := network.NewP2P(opts...)
	f := async.TimeoutFunc(func() error {
		return dspNetwork.Start(hostAddr.Protocol, hostAddr.Address, hostAddr.Port)
	})
	err := async.DoWithTimeout(f, time.Duration(common.START_P2P_TIMEOUT)*time.Second)
	if err != nil {
		return err
	}
	this.p2pActor.SetDspNetwork(dspNetwork)
	this.dspNet = dspNetwork
	this.dspPublicAddr = dspNetwork.PublicAddr()
	return nil
}

func (this *Endpoint) startChannelP2P(hostAddr *common.HostAddr, acc *account.Account) error {
	networkKey := crypto.NewNetworkKeyPairWithAccount(acc)
	opts := []network.NetworkOption{
		network.WithKeys(networkKey),
		network.WithAsyncRecvMsgDisabled(true),
		network.WithNetworkId(config.Parameters.BaseConfig.NetworkId),
		network.WithWalletAddrFromPeerId(crypto.AddressFromPubkeyHex),
		network.WithOpcodes(msg_opcode.OpCodes),
	}
	dsp := this.getDsp()
	log.Debugf("channel p2p dsp %v", dsp)
	if dsp.GetChannelPid() != nil {
		req.SetChannelPid(dsp.GetChannelPid())
	}
	if len(config.Parameters.BaseConfig.IntranetIP) > 0 {
		opts = append(opts, network.WithIntranetIP(config.Parameters.BaseConfig.IntranetIP))
	}
	channelNetwork := network.NewP2P(opts...)
	f := async.TimeoutFunc(func() error {
		return channelNetwork.Start(hostAddr.Protocol, hostAddr.Address, hostAddr.Port)
	})
	err := async.DoWithTimeout(f, time.Duration(common.START_P2P_TIMEOUT)*time.Second)
	if err != nil {
		return err
	}
	this.p2pActor.SetChannelNetwork(channelNetwork)
	this.channelPublicAddr = channelNetwork.PublicAddr()
	this.channelNet = channelNetwork
	return nil
}

func (this *Endpoint) startTrackerP2P(hostAddr *common.HostAddr, acc *account.Account) error {
	tkSrc := tk.NewTrackerService(nil, acc.PublicKey, func(raw []byte) ([]byte, error) {
		return chainSdk.Sign(acc, raw)
	})
	tkActServer, err := tkActServer.NewTrackerActor(tkSrc)
	if err != nil {
		return err
	}
	opts := []tk_net.NetworkOption{
		tk_net.WithKeys(crypto.NewNetworkKeyPairWithAccount(acc)),
		tk_net.WithPid(tkActServer.GetLocalPID()),
		tk_net.WithWalletAddrFromPeerId(crypto.AddressFromPubkeyHex),
		tk_net.WithNetworkId(config.Parameters.BaseConfig.TrackerNetworkId),
		tk_net.WithOpcodes(tk_net.TrackerOpCodes),
	}
	tkNet := tk_net.NewP2P(opts...)
	tk_net.TkP2p = tkNet
	tkActServer.SetNetwork(tkNet)
	tkActClient.SetTrackerServerPid(tkActServer.GetLocalPID())
	this.p2pActor.SetTrackerNet(tkActServer)
	f := async.TimeoutFunc(func() error {
		err := tkNet.Start(hostAddr.Protocol, hostAddr.Address, hostAddr.Port)
		if err != nil {
			return err
		}
		log.Debugf("tk network started, public ip %s", tkNet.PublicAddr())
		return nil
	})
	return async.DoWithTimeout(f, time.Duration(common.START_P2P_TIMEOUT)*time.Second)
}

func (this *Endpoint) updateStorageNodeHost() {
	walletAddr := this.getDspWalletAddress()
	if len(walletAddr) == 0 {
		return
	}
	nodeInfo, err := this.NodeQuery(walletAddr)
	if err != nil || nodeInfo == nil {
		return
	}
	publicAddr := dspActorClient.P2PGetPublicAddr()
	log.Debugf("update node info %s %s", string(nodeInfo.NodeAddr), publicAddr)
	if string(nodeInfo.NodeAddr) == publicAddr {
		log.Debugf("no need to update")
		return
	}
	if _, err := this.NodeUpdate(publicAddr, nodeInfo.Volume, nodeInfo.ServiceTime); err != nil {
		log.Errorf("update node addr failed, err %s", err)
	} else {
		log.Debugf("update it %v %v %v", publicAddr, nodeInfo.Volume, nodeInfo.ServiceTime)
	}
}

func (this *Endpoint) startShareService() {
	// TODO: price needed to be discuss
	_, files, err := this.getDsp().AllDownloadFiles()
	if err == nil {
		this.getDsp().PushFilesToTrackers(files)
	}
}

// SetupDNSNodeBackground. setup a dns node background when received first payments.
func (this *Endpoint) setupDNSNodeBackground() {
	if !config.Parameters.DspConfig.AutoSetupDNSEnable {
		return
	}
	allChannels := this.getDsp().GetAllPartners()
	log.Debugf("setup dns node len %v", len(allChannels))
	setup := func() bool {
		syncing, derr := this.IsChannelProcessBlocks()
		// progress, derr := this.GetFilterBlockProgress()
		if derr != nil {
			log.Errorf("setup dns channel syncing block err %s", derr)
			return false
		}
		if syncing {
			log.Debugf("setup dns wait for init channel")
			return false
		}

		if len(allChannels) > 0 {
			// err := this.getDsp().SetupDNSChannels()
			// if err != nil {
			// 	log.Errorf("SetupDNSChannels err %s", err)
			// 	return false
			// }
			return true
		}
		address, err := chainCom.AddressFromBase58(this.getDsp().WalletAddress())
		if err != nil {
			log.Errorf("setup dns failed address decoded failed")
			return false
		}
		bal, err := this.getDsp().BalanceOf(address)
		if bal == 0 || err != nil {
			log.Errorf("setup dns failed balance is 0")
			return false
		}
		if bal < chainCfg.DEFAULT_GAS_PRICE*chainCfg.DEFAULT_GAS_PRICE {
			log.Errorf("setup dns failed balance not enough %d", bal)
			return false
		}
		// err = this.getDsp().SetupDNSChannels()
		// if err != nil {
		// 	log.Errorf("setup dns channel err %s", err)
		// 	return false
		// }
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
		case <-this.closeCh:
			ti.Stop()
			return
		}
	}
}

func (this *Endpoint) stateChangeService() {
	log.Debugf("start stateChangeService")
	ti := time.NewTicker(time.Duration(common.MAX_STATE_CHANGE_CHECK_INTERVAL) * time.Second)
	for {
		select {
		case <-ti.C:
			log.Debugf("channel service....")
			// check log file size
			go this.checkGoRoutineNum()
			go this.checkLogFileSize()
			if this.dspNet != nil && this.dspPublicAddr != this.dspNet.PublicAddr() {
				log.Debugf("dsp public address has change, old addr: %s, new addr:%s",
					this.dspPublicAddr, this.dspNet.PublicAddr())
				this.dspPublicAddr = this.dspNet.PublicAddr()
				go this.updateStorageNodeHost()
			}
			go this.checkOnlineDNS()
			go this.notifyChannelProgress()
			go this.notifyNewSmartContractEvent()
			go this.notifyNewNetworkState()
		case <-this.closeCh:
			log.Debugf("stop channel service")
			ti.Stop()
			return
		}
	}
}

func (this *Endpoint) checkGoRoutineNum() {
	log.Debugf("go routine num: %d", runtime.NumGoroutine())
}

func (this *Endpoint) checkLogFileSize() {
	logSize, _ := log.GetLogFileSize()
	if logSize < 50*1024*1024 {
		return
	}
	this.initLog()
}

func (this *Endpoint) initLog() {
	log.ClosePrintLog()
	logPath := config.Parameters.BaseConfig.LogPath
	baseDir := config.Parameters.BaseConfig.BaseDir
	extra := ""
	logFullPath := filepath.Join(baseDir, logPath) + extra + "/"
	_, err := log.FileOpen(logFullPath)
	if err != nil {
		extra = strconv.FormatUint(uTime.GetMilliSecTimestamp(), 10)
	}
	logFullPath = filepath.Join(baseDir, logPath) + extra + "/"
	log.Debugf("log new path %s", logFullPath)
	log.InitLog(int(config.Parameters.BaseConfig.LogLevel), logFullPath, log.Stdout)
	// log.SetProcName("saveio")
	go edgeUtils.CleanOldestLogs(logFullPath, config.Parameters.BaseConfig.LogMaxSize)
}

func (this *Endpoint) checkOnlineDNS() {
	if this.getDsp().HasDNS() {
		return
	}
	this.getDsp().BootstrapDNS()
}
func (this *Endpoint) GetFS() *fs.Fs {

	return this.dsp.Fs
}
func (this *Endpoint) GetChain() *chain.Chain {

	return this.dsp.Chain
}
func (this *Endpoint) GetNodeState() state.ModuleState {
	return this.dsp.State()
}
