package dsp

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/saveio/carrier/crypto"
	"github.com/saveio/carrier/crypto/ed25519"
	dsp "github.com/saveio/dsp-go-sdk"
	dspActorClient "github.com/saveio/dsp-go-sdk/actor/client"
	dspCom "github.com/saveio/dsp-go-sdk/common"
	dspCfg "github.com/saveio/dsp-go-sdk/config"
	"github.com/saveio/dsp-go-sdk/store"
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
)

var DspService *Endpoint

type Endpoint struct {
	Dsp          *dsp.Dsp
	Account      *account.Account
	AccountLabel string
	progress     sync.Map
	db           *store.LevelDBStore // TODO: remove this
	sqliteDB     *storage.SQLiteStorage
	closeCh      chan struct{}
	p2pActor     *p2p_actor.P2PActor
	dspNet       *network.Network
	channelNet   *network.Network
}

func Init(walletDir, pwd string) (*Endpoint, error) {
	this := &Endpoint{
		closeCh: make(chan struct{}, 1),
	}
	log.Debugf("walletDir: %s, %d", walletDir, len(walletDir))
	if len(walletDir) == 0 {
		return this, nil
	}
	_, err := os.Open(walletDir)
	if err != nil {
		return this, nil
	}
	wallet, err := wallet.OpenWallet(walletDir)
	if err != nil {
		log.Error("Client open wallet Error, msg:", err)
		return nil, err
	}
	this.Account, err = wallet.GetDefaultAccount([]byte(pwd))
	if err != nil {
		log.Error("Client get default account Error, msg:", err)
		return nil, err
	}
	accData, err := wallet.GetDefaultAccountData()
	if err != nil {
		return nil, err
	}
	this.AccountLabel = accData.Label
	config.SetCurrentUserWalletAddress(this.Account.Address.ToBase58())
	db, err := store.NewLevelDBStore(config.ClientLevelDBPath())
	if err != nil {
		log.Errorf("NewLevelDBStore err %s, config.ClientLevelDBPath():%v", err, config.ClientLevelDBPath())
		return nil, err
	}
	this.db = db
	sqliteDB, err := storage.NewSQLiteStorage(config.ClientSqliteDBPath())
	if err != nil {
		log.Errorf("sqlite err %s", err)
		return nil, err
	}
	this.sqliteDB = sqliteDB
	log.Debug("Endpoint init success")
	DspService = this
	return this, nil
}

func StartDspNode(endpoint *Endpoint, startListen, startShare, startChannel bool) error {
	channelListenAddr := fmt.Sprintf("%s:%d", "127.0.0.1", int(config.Parameters.BaseConfig.PortBase+uint32(config.Parameters.BaseConfig.ChannelPortOffset)))
	dspConfig := &dspCfg.DspConfig{
		DBPath:               config.DspDBPath(),
		FsRepoRoot:           config.FsRepoRootPath(),
		FsFileRoot:           config.FsFileRootPath(),
		FsType:               dspCfg.FSType(config.Parameters.FsConfig.FsType),
		FsGcPeriod:           config.Parameters.FsConfig.FsGCPeriod,
		EnableBackup:         config.Parameters.FsConfig.EnableBackup,
		ChainRpcAddr:         config.Parameters.BaseConfig.ChainRpcAddr,
		ChannelClientType:    config.Parameters.BaseConfig.ChannelClientType,
		ChannelListenAddr:    channelListenAddr,
		ChannelProtocol:      config.Parameters.BaseConfig.ChannelProtocol,
		ChannelRevealTimeout: config.Parameters.BaseConfig.ChannelRevealTimeout,
		ChannelDBPath:        config.ChannelDBPath(),
		ChannelSettleTimeout: config.Parameters.BaseConfig.ChannelSettleTimeout,
		AutoSetupDNSEnable:   config.Parameters.BaseConfig.AutoSetupDNSEnable,
		DnsNodeMaxNum:        config.Parameters.BaseConfig.DnsNodeMaxNum,
		SeedInterval:         config.Parameters.BaseConfig.SeedInterval,
		DnsChannelDeposit:    config.Parameters.BaseConfig.DnsChannelDeposit,
		Trackers:             config.Parameters.BaseConfig.Trackers,
	}
	log.Debugf("dspConfig.dbpath %v, repo: %s, channelDB: %s, wallet: %s, enable backup: %t", dspConfig.DBPath, dspConfig.FsRepoRoot, dspConfig.ChannelDBPath, config.WalletDatFilePath(), config.Parameters.FsConfig.EnableBackup)
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
	log.Debugf("Create Dsp success, version: %s", version)
	if startListen {
		// start dsp net
		dspListenPort := int(config.Parameters.BaseConfig.PortBase + uint32(config.Parameters.BaseConfig.DspPortOffset))
		if dspListenPort == 0 {
			log.Fatal("Not configure HttpRestPort port ")
			return nil
		}

		bPub := keypair.SerializePublicKey(endpoint.Account.PubKey())
		dspPub, dspPrivate, err := ed25519.GenerateKey(&accountReader{
			PublicKey: append(bPub, []byte("dsp")...),
		})
		networkKey := &crypto.KeyPair{
			PublicKey:  dspPub,
			PrivateKey: dspPrivate,
		}
		dspNetwork := network.NewP2P()
		dspNetwork.SetNetworkKey(networkKey)
		dspNetwork.SetHandler(dspSrv.Receive)
		dspNetwork.SetProxyServer(config.Parameters.BaseConfig.NATProxyServerAddr)
		dspListenAddr := fmt.Sprintf("%s://%s:%d", config.Parameters.BaseConfig.DspProtocol, "127.0.0.1", dspListenPort)
		err = dspNetwork.Start(dspListenAddr)
		if err != nil {
			return err
		}
		p2pActor.SetDspNetwork(dspNetwork)
		log.Debugf("start dsp at %s", dspNetwork.PublicAddr())
		// start channel net
		channelNetwork := network.NewP2P()
		channelPubKey, channelPrivateKey, err := ed25519.GenerateKey(&accountReader{
			PublicKey: append(bPub, []byte("channel")...),
		})
		if err != nil {
			return err
		}
		channelNetwork.Keys = &crypto.KeyPair{
			PublicKey:  channelPubKey,
			PrivateKey: channelPrivateKey,
		}
		log.Debugf("dsp pubkey:%s\n", hex.EncodeToString(networkKey.PublicKey))
		log.Debugf("channel pubkey:%s\n", hex.EncodeToString(channelNetwork.Keys.PublicKey))
		channelNetwork.SetProxyServer(config.Parameters.BaseConfig.NATProxyServerAddr)
		req.SetChannelPid(dspSrv.Channel.GetChannelPid())
		listenAddr := fmt.Sprintf("%s://%s", config.Parameters.BaseConfig.ChannelProtocol, dspConfig.ChannelListenAddr)
		log.Debugf("goto start channel network %s", listenAddr)
		err = channelNetwork.Start(listenAddr)
		if err != nil {
			return err
		}
		p2pActor.SetChannelNetwork(channelNetwork)
		endpoint.UpdateNodeIfNeeded()

		log.Debugf("update node finished")
		go endpoint.RegisterChannelEndpoint(dspSrv.CurrentAccount().Address, channelNetwork.PublicAddr())

		// setup filter block range before start
		endpoint.SetFilterBlockRange()
		err = dspSrv.Start()
		if err != nil {
			return err
		}
		if startShare {
			//[TODO] price needed to be discuss
			dspSrv.SetUnitPriceForAllFile(dspCom.ASSET_USDT, common.DSP_DOWNLOAD_UNIT_PRICE)
			endpoint.Dsp.PushLocalFilesToTrackers()
			go dspSrv.StartShareServices()
		}
		endpoint.dspNet = dspNetwork
		endpoint.channelNet = channelNetwork
	}
	// start channel only
	if !startListen && startChannel {
		err := dspSrv.StartChannelService()
		if err != nil {
			return err
		}
	}

	if startListen {
		go endpoint.SetupDNSNodeBackground()
		go endpoint.RegisterProgressCh()
		go endpoint.RegisterShareNotificationCh()
	}
	go endpoint.CheckLogFileSize()
	log.Info("Dsp start success.")
	return nil
}

func (this *Endpoint) RegisterChannelEndpoint(walletAddr chainCom.Address, publicAddr string) error {
	for i := 0; i < common.MAX_REG_CHANNEL_TIMES; i++ {
		err := this.Dsp.RegNodeEndpoint(walletAddr, publicAddr)
		log.Debugf("register endpoint for channel %s, err %s", publicAddr, err)
		if err == nil {
			return nil
		}
		log.Errorf("register endpoint failed %s", err)
		time.Sleep(time.Duration(common.MAX_REG_CHANNEL_BACKOFF) * time.Second)
	}
	log.Errorf("register channel endpoint timeout")
	// testGetIp, err := dspSrv.GetExternalIP(dspSrv.CurrentAccount().Address.ToBase58())
	// log.Debugf("test get ip %s err %s", testGetIp, err)
	return fmt.Errorf("register channel endpoint timeout")
}

func (this *Endpoint) UpdateNodeIfNeeded() {
	nodeInfo, err := this.NodeQuery(this.Account.Address.ToBase58())
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
	_, err = this.NodeUpdate(publicAddr, nodeInfo.Volume, nodeInfo.ServiceTime)
	log.Debugf("update node result")
	if err != nil {
		log.Errorf("update node addr failed, err %s", err)
		panic(err)
	}

}

// SetupDNSNodeBackground. setup a dns node background when received first payments.
func (this *Endpoint) SetupDNSNodeBackground() {
	if !config.Parameters.BaseConfig.AutoSetupDNSEnable {
		return
	}
	allChannels := this.Dsp.Channel.GetAllPartners()
	log.Debugf("setup dns node len %v", len(allChannels))
	setup := func() bool {
		progress, derr := this.GetFilterBlockProgress()
		if derr != nil {
			log.Errorf("setup dns failed filter block err %s", derr)
			return false
		}
		if progress.Progress != 1.0 {
			log.Debugf("setup dns wait for init channel")
			return false
		}

		if len(allChannels) > 0 {
			err := this.Dsp.SetupDNSChannels()
			if err != nil {
				log.Errorf("SetupDNSChannels err %s", err)
				return false
			}
			return true
		}
		address, err := chainCom.AddressFromBase58(this.Dsp.WalletAddress())
		if err != nil {
			log.Errorf("setup dns failed address decoded failed")
			return false
		}
		bal, err := this.Dsp.Chain.Native.Usdt.BalanceOf(address)
		if bal == 0 || err != nil {
			log.Errorf("setup dns failed balance is 0")
			return false
		}
		if bal < chainCfg.DEFAULT_GAS_PRICE*chainCfg.DEFAULT_GAS_PRICE {
			log.Errorf("setup dns failed balance not enough %d", bal)
			return false
		}
		err = this.Dsp.SetupDNSChannels()
		if err != nil {
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
		case <-this.closeCh:
			return
		}
	}
}

func (this *Endpoint) CheckLogFileSize() {
	ti := time.NewTicker(time.Minute)
	for {
		select {
		case <-ti.C:
			isNeedNewFile := log.CheckIfNeedNewFile()
			if isNeedNewFile {
				log.ClosePrintLog()
				logPath := config.Parameters.BaseConfig.LogPath
				baseDir := config.Parameters.BaseConfig.BaseDir
				logFullPath := filepath.Join(baseDir, logPath) + "/"
				log.InitLog(int(config.Parameters.BaseConfig.LogLevel), logFullPath, log.Stdout)
			}
		case <-this.closeCh:
			return
		}
	}
}

// Stop. stop endpoint instance
func (this *Endpoint) Stop() error {
	if this.p2pActor != nil {
		err := this.p2pActor.Stop()
		if err != nil {
			return err
		}
	}
	close(this.closeCh)
	err := this.db.Close()
	if err != nil {
		return err
	}
	err = this.sqliteDB.Close()
	if err != nil {
		return err
	}
	this.ResetChannelProgress()
	return this.Dsp.Stop()
}
