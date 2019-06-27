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
	channelNet "github.com/saveio/edge/p2p/networks/channel"
	dspNet "github.com/saveio/edge/p2p/networks/dsp"
	"github.com/saveio/themis-go-sdk/wallet"
	"github.com/saveio/themis/account"
	chainCom "github.com/saveio/themis/common"
	chainCfg "github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"
)

var DspService *Endpoint

type Endpoint struct {
	Dsp      *dsp.Dsp
	Account  *account.Account
	progress sync.Map
	db       *store.LevelDBStore // TODO: remove this
	sqliteDB *storage.SQLiteStorage
	closhCh  chan struct{}
	p2pActor *p2p_actor.P2PActor
}

func Init(walletDir, pwd string) (*Endpoint, error) {
	this := &Endpoint{
		closhCh: make(chan struct{}, 1),
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
	log.Debug("Endpoint init successed")
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
		ChainRpcAddr:         config.Parameters.BaseConfig.ChainRpcAddr,
		ChannelClientType:    config.Parameters.BaseConfig.ChannelClientType,
		ChannelListenAddr:    channelListenAddr,
		ChannelProtocol:      config.Parameters.BaseConfig.ChannelProtocol,
		ChannelRevealTimeout: config.Parameters.BaseConfig.ChannelRevealTimeout,
		ChannelDBPath:        config.ChannelDBPath(),
		AutoSetupDNSEnable:   config.Parameters.BaseConfig.AutoSetupDNSEnable,
		DnsNodeMaxNum:        config.Parameters.BaseConfig.DnsNodeMaxNum,
		SeedInterval:         config.Parameters.BaseConfig.SeedInterval,
		DnsChannelDeposit:    config.Parameters.BaseConfig.DnsChannelDeposit,
	}
	log.Debugf("dspConfig.dbpath %v", dspConfig.DBPath)
	log.Debugf("dspConfig.FsRepoRoot %v", dspConfig.FsRepoRoot)
	log.Debugf("dspConfig.ChannelDBPath %v", dspConfig.ChannelDBPath)
	log.Debugf("WalletDatFilePath %v", config.WalletDatFilePath())
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
	log.Debugf("dspConfig.ChannelListenAddr :%s ,acc %v\n", dspConfig.ChannelListenAddr, endpoint.Account)
	dspSrv := dsp.NewDsp(dspConfig, endpoint.Account, p2pActor.GetLocalPID())
	if dspSrv == nil {
		return errors.New("dsp server init failed")
	}
	log.Debugf("Create Dsp success %p", endpoint)
	endpoint.Dsp = dspSrv
	if startListen {
		// start dsp net
		dspListenPort := int(config.Parameters.BaseConfig.PortBase + uint32(config.Parameters.BaseConfig.DspPortOffset))
		if dspListenPort == 0 {
			log.Fatal("Not configure HttpRestPort port ")
			return nil
		}
		bPrivate := keypair.SerializePrivateKey(endpoint.Account.PrivKey())
		bPub := keypair.SerializePublicKey(endpoint.Account.PubKey())
		networkKey := &crypto.KeyPair{
			PrivateKey: bPrivate,
			PublicKey:  bPub,
		}
		log.Debugf("networkKey:%v", networkKey)
		dspNetwork := dspNet.NewNetwork()
		// dspNetwork.SetNetworkKey(networkKey)
		dspNetwork.SetHandler(dspSrv.Receive)
		dspNetwork.SetProxyServer(config.Parameters.BaseConfig.NATProxyServerAddr)
		dspListenAddr := fmt.Sprintf("%s://%s:%d", config.Parameters.BaseConfig.DspProtocol, "127.0.0.1", dspListenPort)
		log.Debugf("goto start dsp network %s", dspListenAddr)
		err := dspNetwork.Start(dspListenAddr)
		if err != nil {
			return err
		}
		p2pActor.SetDspNetwork(dspNetwork)
		log.Debugf("start dsp at %s", dspNetwork.PublicAddr())
		// start channel net
		channelNetwork := channelNet.NewP2P()
		// channelNetwork.Keys = networkKey
		log.Debugf("privteKey:%s, pubkey:%s\n", hex.EncodeToString(bPrivate), hex.EncodeToString(bPub))
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
		err = dspSrv.RegNodeEndpoint(dspSrv.CurrentAccount().Address, channelNetwork.PublicAddr())
		log.Debugf("register endpoint for channel %s", channelNetwork.PublicAddr())
		if err != nil {
			log.Errorf("register endpoint failed %s", err)
			return err
		}
		testGetIp, err := dspSrv.GetExternalIP(dspSrv.CurrentAccount().Address.ToBase58())
		log.Debugf("test get ip %s err %s", testGetIp, err)
		// setup filter block range before start
		endpoint.SetFilterBlockRange()
		log.Debugf("will start dsp")
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
	log.Info("Dsp start successed.")
	return nil
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
	if config.Parameters.BaseConfig.AutoSetupDNSEnable {
		return
	}
	allChannels := this.Dsp.Channel.GetAllPartners()
	log.Debugf("setup dns node len %v", len(allChannels))
	ti := time.NewTicker(time.Second)
	for {
		select {
		case <-ti.C:
			progress, derr := this.GetFilterBlockProgress()
			if derr != nil {
				log.Errorf("setup dns failed filter block err %s", derr)
				break
			}
			if progress.Progress != 1.0 {
				log.Debugf("setup dns wait for init channel")
				break
			}

			if len(allChannels) > 0 {
				err := this.Dsp.SetupDNSChannels()
				if err != nil {
					// panic(err)
					log.Errorf("set up dns failed %s", err)
				}
				return
			}
			address, err := chainCom.AddressFromBase58(this.Dsp.WalletAddress())
			if err != nil {
				log.Errorf("setup dns failed address decoded failed")
				break
			}
			bal, err := this.Dsp.Chain.Native.Usdt.BalanceOf(address)
			if bal == 0 || err != nil {
				log.Errorf("setup dns failed balance is 0")
				break
			}
			if bal < chainCfg.DEFAULT_GAS_PRICE*chainCfg.DEFAULT_GAS_PRICE {
				log.Errorf("setup dns failed balance not enough %d", bal)
				break
			}
			err = this.Dsp.SetupDNSChannels()
			if err != nil {
				break
			}
			return
		case <-this.closhCh:
			return
		}
	}
}

func (this *Endpoint) CheckLogFileSize() {
	ti := time.NewTicker(time.Second)
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
		case <-this.closhCh:
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
	close(this.closhCh)
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
