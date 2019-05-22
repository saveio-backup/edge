package dsp

import (
	"container/list"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
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
	chanCom "github.com/saveio/pylons/common"
	themisSdk "github.com/saveio/themis-go-sdk"
	"github.com/saveio/themis-go-sdk/wallet"
	"github.com/saveio/themis/account"
	chainCom "github.com/saveio/themis/common"
	chainCfg "github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"
	fs "github.com/saveio/themis/smartcontract/service/native/onifs"
)

type Endpoint struct {
	Account  *account.Account
	Chain    *themisSdk.Chain
	Dsp      *dsp.Dsp
	progress sync.Map
	db       *store.LevelDBStore
	sqliteDB *storage.SQLiteStorage
	closhCh  chan struct{}
}

func Init(walletDir, pwd string) (*Endpoint, error) {
	this := &Endpoint{
		closhCh: make(chan struct{}, 1),
	}
	log.Debugf("walletDir: %s, %d", walletDir, len(walletDir))
	chain := themisSdk.NewChain()
	chain.NewRestClient().SetAddress(config.Parameters.BaseConfig.ChainRestAddr)
	this.Chain = chain
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
	this.Chain.SetDefaultAccount(this.Account)
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
	log.Debugf("dspConfig.ChannelListenAddr :%s ,acc %v\n", dspConfig.ChannelListenAddr, endpoint.Account)
	dspSrv := dsp.NewDsp(dspConfig, endpoint.Account, p2pActor.GetLocalPID())
	if dspSrv == nil {
		return errors.New("dsp server init failed")
	}
	log.Debug("Create Dsp success")
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

		dspNetwork := dspNet.NewNetwork()
		dspNetwork.SetNetworkKey(networkKey)
		dspNetwork.SetHandler(dspSrv.Receive)
		dspNetwork.SetProxyServer(config.Parameters.BaseConfig.NATProxyServerAddr)
		dspListenAddr := fmt.Sprintf("%s://%s:%d", config.Parameters.BaseConfig.DspProtocol, "127.0.0.1", dspListenPort)
		log.Debugf("goto start dsp network")
		err := dspNetwork.Start(dspListenAddr)
		if err != nil {
			return err
		}
		p2pActor.SetDspNetwork(dspNetwork)
		log.Debugf("start dsp at %s", dspNetwork.PublicAddr())
		// start channel net
		channelNetwork := channelNet.NewP2P()
		channelNetwork.Keys = networkKey
		log.Debugf("privteKey:%s, pubkey:%s\n", hex.EncodeToString(bPrivate), hex.EncodeToString(bPub))
		channelNetwork.SetProxyServer(config.Parameters.BaseConfig.NATProxyServerAddr)
		req.SetChannelPid(dspSrv.Channel.GetChannelPid())
		listenAddr := fmt.Sprintf("%s://%s", config.Parameters.BaseConfig.ChannelProtocol, dspConfig.ChannelListenAddr)
		log.Debugf("goto start channel network")
		err = channelNetwork.Start(listenAddr)
		if err != nil {
			return err
		}
		p2pActor.SetChannelNetwork(channelNetwork)
		log.Debugf("channelNetwork.PublicAddr(): %s", channelNetwork.PublicAddr())
		endpoint.UpdateNodeIfNeeded()
		err = dspSrv.RegNodeEndpoint(dspSrv.CurrentAccount().Address, channelNetwork.PublicAddr())
		if err != nil {
			log.Errorf("register endpoint failed %s", err)
			return err
		}
		// time.Sleep(time.Duration(5) * time.Second)
		err = dspSrv.Start()
		if err != nil {
			return err
		}
		if startShare {
			//[TODO] price needed to be discuss
			dspSrv.SetUnitPriceForAllFile(dspCom.ASSET_USDT, common.DSP_DOWNLOAD_UNIT_PRICE)
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
		return
	}
	_, err = this.NodeUpdate(publicAddr, nodeInfo.Volume, nodeInfo.ServiceTime)
	if err != nil {
		log.Errorf("update node addr failed, err %s", err)
	}
	// TODO: update to trackes
}

// SetupDNSNodeBackground. setup a dns node background when received first payments.
func (this *Endpoint) SetupDNSNodeBackground() {
	if config.Parameters.BaseConfig.AutoSetupDNSEnable {
		return
	}
	allChannels := this.Dsp.Channel.GetAllPartners()
	log.Debugf("setup dns node len %v", len(allChannels))
	if len(allChannels) > 0 {
		err := this.Dsp.SetupDNSChannels()
		if err != nil {
			panic(err)
		}
		return
	}
	ti := time.NewTicker(time.Second)
	for {
		select {
		case <-ti.C:
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

// Stop. stop endpoint instance
func (this *Endpoint) Stop() error {
	// TODO: stop networks
	close(this.closhCh)
	err := this.db.Close()
	if err != nil {
		return err
	}
	err = this.sqliteDB.Close()
	if err != nil {
		return err
	}
	return this.Dsp.Stop()
}

//Dsp api
func (this *Endpoint) RegisterNode(addr string, volume, serviceTime uint64) (string, error) {
	tx, err := this.Dsp.RegisterNode(addr, volume, serviceTime)
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", err
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) UnregisterNode() (string, error) {
	tx, err := this.Dsp.UnregisterNode()
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", err
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) NodeQuery(walletAddr string) (*fs.FsNodeInfo, error) {
	info, err := this.Dsp.QueryNode(walletAddr)
	if err != nil {
		log.Errorf("query node err %s", err)
		return nil, err
	}
	log.Infof("node info pledge %d", info.Pledge)
	log.Infof("node info profit %d", info.Profit)
	log.Infof("node info volume %d", info.Volume)
	log.Infof("node info restvol %d", info.RestVol)
	log.Infof("node info service time %d", info.ServiceTime)
	log.Infof("node info wallet address %s", info.WalletAddr.ToBase58())
	log.Infof("node info node address %s", info.NodeAddr)

	return info, nil
}

func (this *Endpoint) NodeUpdate(addr string, volume, serviceTime uint64) (string, error) {
	tx, err := this.Dsp.UpdateNode(addr, volume, serviceTime)
	if err != nil {
		log.Errorf("update node err:%s", err)
		return "", err
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) NodeWithdrawProfit() (string, error) {
	tx, err := this.Dsp.NodeWithdrawProfit()
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", err
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

//oniChannel api
func (this *Endpoint) OpenPaymentChannel(partnerAddr string) (chanCom.ChannelID, error) {
	return this.Dsp.Channel.OpenChannel(partnerAddr)
}

func (this *Endpoint) ClosePaymentChannel(regAddr, tokenAddr, partnerAddr string, retryTimeout float64) {
	//[TODO] call channel close function of dsp-go-sdk
	return
}

func (this *Endpoint) DepositToChannel(partnerAddr string, totalDeposit uint64) error {
	return this.Dsp.Channel.SetDeposit(partnerAddr, totalDeposit)
}

func (this *Endpoint) Transfer(paymentId int32, amount uint64, to string) error {
	return this.Dsp.Channel.MediaTransfer(paymentId, amount, to)
}

func (this *Endpoint) GetChannelListByOwnerAddress(addr string, tokenAddr string) *list.List {
	//[TODO] call dsp-go-sdk function to return channel list
	//[NOTE] addr and token Addr should NOT be needed. addr mean PaymentNetworkID
	//tokenAddr mean TokenAddress. Need comfirm the behavior when integrate dsp-go-sdk with oniChannel
	return list.New()
}

func (this *Endpoint) QuerySpecialChannelDeposit(partnerAddr string) (uint64, error) {
	return this.Dsp.Channel.GetTotalDepositBalance(partnerAddr)
}

func (this *Endpoint) QuerySpecialChannelAvaliable(partnerAddr string) (uint64, error) {
	return this.Dsp.Channel.GetAvaliableBalance(partnerAddr)
}

// ChannelWithdraw. withdraw amount of asset from channel
func (this *Endpoint) ChannelWithdraw(partnerAddr string, amount uint64) error {
	totalWithdraw, err := this.Dsp.Channel.GetTotalWithdraw(partnerAddr)
	if err != nil {
		return err
	}
	bal := amount + totalWithdraw
	if bal-amount != totalWithdraw {
		return errors.New("withdraw overflow")
	}
	success, err := this.Dsp.Channel.Withdraw(partnerAddr, bal)
	if err != nil {
		return err
	}
	if !success {
		return errors.New("withdraw failed")
	}
	return nil
}

// ChannelCooperativeSettle. settle channel cooperatively
func (this *Endpoint) ChannelCooperativeSettle(partnerAddr string) error {
	return this.Dsp.Channel.CooperativeSettle(partnerAddr)
}
